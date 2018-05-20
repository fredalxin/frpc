package client

import (
	"bufio"
	"context"
	"errors"
	"frpc/core"
	err "frpc/error"
	"frpc/log"
	"frpc/protocol"
	"frpc/selector"
	"frpc/util"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type ServiceError string

func (e ServiceError) Error() string {
	return string(e)
}

var (
	ErrShutdown         = errors.New("connection is shut down")
	ErrUnsupportedCodec = errors.New("unsupported codec")
	ErrClientNoServer   = errors.New("can not found any server")
	// ErrServerUnavailable selected server is unavailable.
	ErrServerUnavailable = errors.New("selected server is unavilable")
)

type seqKey struct{}

type Client struct {
	option   Option
	reqMutex sync.Mutex // protects following
	r        *bufio.Reader
	conn     net.Conn
	mutex    sync.RWMutex // protects following
	seq      uint64
	pending  map[uint64]*Call
	closing  bool // user has called Close
	shutdown bool // server has told us to stop
	//for xclient use
	cachedClient      map[string]*Client
	servicePath       string
	serverMessageChan chan<- *protocol.Message

	registry RegistryClient
	selector selector.Selector
}

func NewClient() *Client {
	return newClient().
		ConnTimeout(10 * time.Second).
		Serialize(protocol.MsgPack).
		Compress(protocol.None).
		Retries(3).
		FailMode(core.FailFast)
}

func newClient() *Client {
	return &Client{}
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
		log.Debug("frpc: discarding Call reply due to insufficient Done chan capacity")
	}
}

func (client *Client) handleResponse() {
	var err error
	var res = protocol.NewMessage()
	for err == nil {
		err = res.Decode(client.r)
		if err != nil {
			break
		}
		seq := res.Seq()
		var call *Call
		isServerMessage := (res.MessageType() == protocol.Request && !res.IsHeartbeat() && res.IsOneway())
		if !isServerMessage {
			client.mutex.Lock()
			call = client.pending[seq]
			delete(client.pending, seq)
			client.mutex.Unlock()
		}

		switch {
		case call == nil:
			if isServerMessage {
				//if client.ServerMessageChan != nil {
				//	go client.handleServerRequest(res)
				//}
				continue
			}
		case res.MessageStatusType() == protocol.Error:
			call.Error = ServiceError(res.Metadata[protocol.ServiceError])
			call.ResMetadata = res.Metadata
			//
			call.done()
		case res.IsHeartbeat():
			log.Info("client receive heartbeat from server")
			call.ResMetadata = res.Metadata
			call.done()
		default:
			data := res.Payload
			if len(data) > 0 {
				if res.CompressType() == protocol.Gzip {
					data, err = util.Unzip(data)
					if err != nil {
						call.Error = ServiceError("unzip payload: " + err.Error())
					}
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
	//todo
	//if err != nil && err != io.EOF && !closing {
	//	log.Error("frpc: client protocol error:", err)
	//}
}

func (client *Client) heartbeat() {
	t := time.NewTicker(client.option.HeartbeatInterval)
	for range t.C {
		if client.shutdown || client.closing {
			return
		}
		log.Info("client heartbeat")
		err := client.CallDirect(context.Background(), "", "", nil, nil)
		if err != nil {
			log.Warnf("failed to heartbeat to %s", client.conn.RemoteAddr().String())
		}
	}
}

//需要处理selectmode
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
			log.Panic("frpc: done channel is unbuffered")
		}
	}
	call.Done = done
	client.send(ctx, call)
	return call
}

func (c *Client) CallProxy(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {
	return c.Call(ctx, c.servicePath, serviceMethod, args, reply)
}

func (c *Client) Call(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error {
	if c.selector == nil {
		c.Selector(core.Random)
	}
	if c.registry.Discovery == nil {
		errorStr := "please set registry first,or use callDirect"
		log.Errorf(errorStr)
		return errors.New(errorStr)
	}

	cname, client, err := c.selectClient(ctx, servicePath, serviceMethod, args)
	if err != nil {
		if c.option.failMode == core.FailFast {
			return err
		}
	}
	switch c.option.failMode {
	case core.FailTry:
		retries := c.option.Retries
		for retries > 0 {
			retries--
			if client != nil {
				err = client.CallDirect(ctx, servicePath, serviceMethod, args, reply)
				if err == nil {
					return nil
				}
			}
			c.removeClient(cname, client)
			client, _ = c.getCachedClient(cname)
		}
		return err
	case core.FailOver:
		retries := c.option.Retries
		for retries > 0 {
			retries--
			if client != nil {
				err = client.CallDirect(ctx, servicePath, serviceMethod, args, reply)
				if err == nil {
					return nil
				}
			}
			c.removeClient(cname, client)
			cname, client, _ = c.selectClient(ctx, servicePath, serviceMethod, args)
		}
		return err
	default: //failfast
		err = client.CallDirect(ctx, servicePath, serviceMethod, args, reply)
		if err != nil {
			if _, ok := err.(ServiceError); !ok {
				c.removeClient(cname, client)
			}
		}
		return err
	}
}

func (client *Client) CallDirect(ctx context.Context, servicePath, serviceMethod string, args interface{}, reply interface{}) error {
	if client.option.Breaker.breaker != nil {
		return client.option.Breaker.breaker.Call(func() error {
			return client.call(ctx, servicePath, serviceMethod, args, reply)
		}, client.option.Breaker.timeout)
	}
	return client.call(ctx, servicePath, serviceMethod, args, reply)
}

func (c *Client) removeClient(cname string, client *Client) {
	c.mutex.Lock()
	cl := c.cachedClient[cname]
	if cl == client {
		delete(c.cachedClient, cname)
	}
	c.mutex.Unlock()

	if client != nil {
		client.UnregisterServerMessageChan()
		client.Close()
	}
}

func (c *Client) selectClient(ctx context.Context, servicePath, serviceMethod string, args interface{}) (string, *Client, error) {
	cname := c.selector.Select(ctx, servicePath, serviceMethod, args)
	if cname == "" {
		return "", nil, ErrClientNoServer
	}

	client, err := c.getCachedClient(cname)
	return cname, client, err
}

func (c *Client) getCachedClient(cname string) (*Client, error) {
	c.mutex.RLock()
	client := c.cachedClient[cname]
	if client != nil {
		if !client.closing && !client.shutdown {
			c.mutex.RUnlock()
			return client, nil
		}
	}
	c.mutex.RUnlock()

	//double check
	c.mutex.Lock()
	client = c.cachedClient[cname]
	if client == nil {
		network, addr := splitNetworkAndAddress(cname)
		//todo 属性赋值
		client = &Client{
			option:   c.option,
			registry: c.registry,
			selector: c.selector,
		}
		err := client.Connect(network, addr)
		if err != nil {
			c.mutex.Unlock()
			return nil, err
		}

		client.RegisterServerMessageChan(c.serverMessageChan)
		c.cachedClient[cname] = client
	}
	c.mutex.Unlock()

	return client, nil
}

func splitNetworkAndAddress(server string) (string, string) {
	ss := strings.SplitN(server, "@", 2)
	if len(ss) == 1 {
		return "tcp", server
	}

	return ss[0], ss[1]
}

func (client *Client) RegisterServerMessageChan(ch chan<- *protocol.Message) {
	client.serverMessageChan = ch
}

func (client *Client) UnregisterServerMessageChan() {
	client.serverMessageChan = nil
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

func (c *Client) Close() error {
	c.shutdown = true
	var errs []error
	c.mutex.Lock()
	for k, v := range c.cachedClient {
		e := v.close()
		if e != nil {
			errs = append(errs, e)
		}
		delete(c.cachedClient, k)
	}
	c.mutex.Unlock()
	go func() {
		defer func() {
			if r := recover(); r != nil {

			}
		}()
		c.registry.Discovery.RemoveWatcher(c.registry.ch)
		close(c.registry.ch)
	}()
	if len(errs) > 0 {
		return err.NewMultiError(errs)
	}
	return nil
}

func (client *Client) close() error {
	client.mutex.Lock()
	for seq, call := range client.pending {
		if call != nil {
			call.Error = ErrShutdown
			call.done()
		}
		delete(client.pending, seq)
	}

	if client.closing || client.shutdown {
		client.mutex.Unlock()
		return ErrShutdown
	}
	client.closing = true
	client.mutex.Unlock()
	return client.conn.Close()
}
