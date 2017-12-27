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
)

type ServiceError string

func (e ServiceError) Error() string {
	return string(e)
}

var (
	ErrShutdown         = errors.New("connection is shut down")
	ErrUnsupportedCodec = errors.New("unsupported codec")
)

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

}
