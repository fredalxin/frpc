package client

import (
	"sync"
	"bufio"
	"frpc/protocol"
	"time"
	"net"
	"frpc/log"
	"rpcx/util"
	"frpc/core"
	"errors"
	"io"
	"context"
)

type ServiceError string

func (e ServiceError) Error() string {
	return string(e)
}

var (
	ErrShutdown         = errors.New("connection is shut down")
	ErrUnsupportedCodec = errors.New("unsupported codec")
)

type seqKey struct{}

type Client struct {
	option   Option
	reqMutex sync.Mutex // protects following
	r        *bufio.Reader
	conn     net.Conn
	mutex    sync.Mutex // protects following
	seq      uint64
	pending  map[uint64]*Call
	closing  bool // user has called Close
	shutdown bool // server has told us to stop
}

type Option struct {
	SerializeType     protocol.SerializeType
	CompressType      protocol.CompressType
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	ConnTimeout       time.Duration
	Heartbeat         bool
	HeartbeatInterval time.Duration
}

type Call struct {
	ServicePath   string
	ServiceMethod string            // The name of the service and method to call.
	Metadata      map[string]string //metadata
	ResMetadata   map[string]string
	Args          interface{} // The argument to the function (*struct).
	Reply         interface{} // The reply from the function (*struct).
	Error         error       // After completion, the error status.
	Done          chan *Call
	Raw           bool
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		log.Debug("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
}

func (client *Client) handleResponse(reader io.Reader) {
	var err error
	var res = protocol.NewMessage()
	for err == nil {
		err := res.Decode(reader)
		if err != nil {
			break
		}
		seq := res.Seq()
		client.mutex.Lock()
		call := client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()

		switch {
		case call == nil:
		case res.MessageStatusType() == protocol.Error:
			call.Error = ServiceError(res.Metadata[protocol.ServiceError])
			call.ResMetadata = res.Metadata
			//
			call.done()
		default:
			data := res.Payload
			if len(data) > 0 {
				if res.CompressType() == protocol.Gzip {
					data, err = util.Unzip(data)
					if err != nil {
					}
					call.Error = ServiceError("unzip payload: " + err.Error())
				}
				codec := core.Codecs[res.SerializeType()]
				if codec == nil {
					call.Error = ServiceError(ErrUnsupportedCodec.Error())
				} else {
					err = codec.Decode(data, call.Reply)
					if err != nil {
						call.Error = ServiceError(err.Error())
					}
				}
			}
			call.ResMetadata = res.Metadata
			call.done()
		}
		res.Reset()
	}
	client.mutex.Lock()
	client.shutdown = true
	closing := client.closing
	if err == io.EOF {
		if closing {
			err = ErrShutdown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
	client.mutex.Unlock()
	if err != nil && err != io.EOF && !closing {
		log.Error("rpcx: client protocol error:", err)
	}
}

func (client *Client) heartbeat() {
	t := time.NewTicker(client.option.HeartbeatInterval)
	for range t.C {
		if client.shutdown || client.closing {
			return
		}
		err := client.Call(context.Background(), "", "", nil, nil)
		if err != nil {
			log.Warnf("failed to heartbeat to %s", client.conn.RemoteAddr().String())
		}
	}
}

func (client *Client) Go(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call {
	call := new(Call)
	call.ServicePath = servicePath
	call.ServiceMethod = serviceMethod
	meta := ctx.Value(core.ReqMetaDataKey)
	if meta != nil {
		call.Metadata = meta.(map[string]string)
	}
	call.Args = args
	call.Reply = reply
	if done == nil {
		//buffer
		done = make(chan *Call, 10)
	} else {
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	client.send(ctx, call)
	return call
}

func (client *Client) Call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error {
	//熔断 待开发
	return client.call(ctx, servicePath, serviceMethod, args, reply)
}

func (client *Client) call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error {
	seq := new(uint64)
	context.WithValue(ctx, seqKey{}, seq)
	Done := client.Go(ctx, servicePath, serviceMethod, args, reply, make(chan *Call, 1)).Done
	var err error
	select {
	case <-ctx.Done():
		client.mutex.Lock()
		call := client.pending[*seq]
		delete(client.pending, *seq)
		client.mutex.Unlock()
		if call != nil {
			call.Error = ctx.Err()
			call.done()
		}
		return ctx.Err()
	case call := <-Done:
		err = call.Error
		meta := ctx.Value(core.ResMetaDataKey)
		if meta != nil && len(call.ResMetadata) > 0 {
			resMeta := meta.(map[string]string)
			for k, v := range call.ResMetadata {
				resMeta[k] = v
			}
		}
	}
	return err
}

func (client *Client) send(ctx context.Context, call *Call) {
	//client.reqMutex.Lock()
	//defer client.reqMutex.Unlock()
	//
	//// Register this call.
	//client.mutex.Lock()
	//if client.shutdown || client.closing {
	//	call.Error = ErrShutdown
	//	client.mutex.Unlock()
	//	call.done()
	//	return
	//}
	////pending使用map实现的，rpc请求都会生存一个唯一递增的seq, seq就是用来标记请求的，这个很像tcp包的seq
	//
	//seq := client.seq
	//client.seq++
	//client.pending[seq] = call
	//client.mutex.Unlock()
	//
	//// Encode and send the request.
	//client.request.Seq = seq
	//client.request.ServiceMethod = call.ServiceMethod
	//err := client.codec.WriteRequest(&client.request, call.Args)
	//if err != nil {
	//	client.mutex.Lock()
	//	call = client.pending[seq]
	//	delete(client.pending, seq)
	//	client.mutex.Unlock()
	//	if call != nil {
	//		call.Error = err
	//		call.done()
	//	}
	//}
	client.mutex.Lock()
	if client.shutdown || client.closing {
		call.Error = ErrShutdown
		client.mutex.Unlock()
		call.done()
		return
	}

	codec := core.Codecs[client.option.SerializeType]
	if codec == nil {
		call.Error = ErrUnsupportedCodec
		client.mutex.Unlock()
		call.done()
		return
	}

	if client.pending == nil {
		client.pending = make(map[uint64]*Call)
	}
	seq := client.seq
	client.seq++
	client.pending[seq] = call
	client.mutex.Unlock()

	if cseq, ok := ctx.Value(seqKey{}).(*uint64); ok {
		*cseq = seq
	}

	req := protocol.GetMsgs()
	req.SetMessageType(protocol.Request)
	req.SetSeq(seq)

	if call.ServicePath == "" && call.ServiceMethod == "" {
		req.SetHeartbeat(true)
	} else {
		req.SetSerializeType(client.option.SerializeType)
		if call.Metadata != nil {
			req.Metadata = call.Metadata
		}

		req.ServicePath = call.ServicePath
		req.ServiceMethod = call.ServiceMethod

		data, err := codec.Encode(call.Args)
		if err != nil {
			call.Error = err
			call.done()
			return
		}
		if len(data) > 1024 && client.option.CompressType == protocol.Gzip {
			data, err = util.Zip(data)
			if err != nil {
				call.Error = err
				call.done()
				return
			}

			req.SetCompressType(client.option.CompressType)
		}
		req.Payload = data
	}

	data := req.Encode()

	_, err := client.conn.Write(data)
	if err != nil {
		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()
		if call != nil {
			call.Error = err
			call.done()
		}
	}
	protocol.FreeMsg(req)

	if req.IsOneway() {
		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()
		if call != nil {
			call.done()
		}
	}
}
