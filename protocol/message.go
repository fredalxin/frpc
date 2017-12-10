package protocol

import (
	"io"
	"encoding/binary"
	"errors"
	"frpc/util"
)

const (
	magicNumber byte = 0x08
)

var MaxMessageLength = 0

var (
	// ErrMetaKVMissing some keys or values are mssing.
	ErrMetaKVMissing = errors.New("wrong metadata lines. some keys or values are missing")
	// ErrMessageToLong message is too long
	ErrMessageToLong = errors.New("message is too long")
)

type Header [12]byte

type Message struct {
	*Header
	ServicePath   string
	ServiceMethod string
	Metadata      map[string]string
	Payload       []byte
	data          []byte
}

func (message *Message) Decode(reader io.Reader) error {
	// parse header
	_, err := io.ReadFull(reader, message.Header[:])
	if err !=nil{
		return err
	}
	lenData := poolUint32Data.Get().(*[]byte)
	_, err = io.ReadFull(reader, *lenData)
	if err!=nil{
		poolUint32Data.Put(lenData)
		return err
	}
	l := binary.BigEndian.Uint32(*lenData)
	poolUint32Data.Put(l)

	if MaxMessageLength > 0 && int(l) > MaxMessageLength{
		return ErrMessageToLong
	}

	data := make([]byte, int(l))
	_, err = io.ReadFull(reader, data)
	if err != nil {
		return err
	}
	message.data = data

	n := 0
	// parse servicePath
	l = binary.BigEndian.Uint32(data[n:4])
	n = n + 4
	nEnd := n + int(l)
	message.ServicePath = util.ByteToString(data[n:nEnd])
	n = nEnd

	// parse serviceMethod
	l = binary.BigEndian.Uint32(data[n : n+4])
	n = n + 4
	nEnd = n + int(l)
	message.ServiceMethod = util.ByteToString(data[n:nEnd])
	n = nEnd

	// parse meta
	l = binary.BigEndian.Uint32(data[n : n+4])
	n = n + 4
	nEnd = n + int(l)

	if l > 0 {
		message.Metadata, err = decodeMetadata(l, data[n:nEnd])
		if err != nil {
			return err
		}
	}
	n = nEnd

	// parse payload
	l = binary.BigEndian.Uint32(data[n : n+4])
	_ = l
	n = n + 4
	message.Payload = data[n:]

	return err
}

func decodeMetadata(l uint32, data []byte) (map[string]string, error) {
	m := make(map[string]string, 10)
	n := uint32(0)
	for n < l {
		// parse one key and value
		// key
		sl := binary.BigEndian.Uint32(data[n : n+4])
		n = n + 4
		if n+sl > l-4 {
			return m, ErrMetaKVMissing
		}
		k := util.ByteToString(data[n : n+sl])
		n = n + sl

		// value
		sl = binary.BigEndian.Uint32(data[n : n+4])
		n = n + 4
		if n+sl > l {
			return m, ErrMetaKVMissing
		}
		v := util.ByteToString(data[n : n+sl])
		n = n + sl
		m[k] = v
	}

	return m, nil
}
