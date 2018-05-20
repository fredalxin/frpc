package codec

import (
	"encoding/json"
	"fmt"
	pb "github.com/golang/protobuf/proto"
	mp "github.com/vmihailenco/msgpack"
)

type Codec interface {
	Encode(i interface{}) ([]byte, error)
	Decode(data []byte, i interface{}) error
}

// JSONCodec uses json marshaler and unmarshaler.
type JSONCodec struct{}

// Encode encodes an object into slice of bytes.
func (c JSONCodec) Encode(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

// Decode decodes an object from slice of bytes.
func (c JSONCodec) Decode(data []byte, i interface{}) error {
	return json.Unmarshal(data, i)
}

type PBCodec struct{}

// Encode encodes an object into slice of bytes.
func (c PBCodec) Encode(i interface{}) ([]byte, error) {
	if m, ok := i.(pb.Marshaler); ok {
		return m.Marshal()
	}
	if m, ok := i.(pb.Message); ok {
		return pb.Marshal(m)
	}
	return nil, fmt.Errorf("%T is not a proto.Unmarshaler", i)
}

// Decode decodes an object from slice of bytes.
func (c PBCodec) Decode(data []byte, i interface{}) error {
	if m, ok := i.(pb.Unmarshaler); ok {
		return m.Unmarshal(data)
	}
	if m, ok := i.(pb.Message); ok {
		return pb.Unmarshal(data, m)
	}
	return fmt.Errorf("%T is not a proto.Unmarshaler", i)
}

type MsgpackCodec struct{}

// Encode encodes an object into slice of bytes.
func (c MsgpackCodec) Encode(i interface{}) ([]byte, error) {
	return mp.Marshal(i)
}

// Decode decodes an object from slice of bytes.
func (c MsgpackCodec) Decode(data []byte, i interface{}) error {
	return mp.Unmarshal(data, i)
}
