package protocol

import "sync"

var msgPool = sync.Pool{
	New: func() interface{} {
		header := Header([12]byte{})
		header[0] = magicNumber

		return &Message{
			Header: &header,
		}
	},
}

// GetPooledMsg gets a pooled message.
func GetMsgs() *Message {
	return msgPool.Get().(*Message)
}

var poolUint32Data = sync.Pool{
	New: func() interface{} {
		data := make([]byte, 4)
		return &data
	},
}