package protocol

import (
	"io"
	"encoding/binary"
	"errors"
	"frpc/util"
	"bytes"
)

const (
	magicNumber byte = 0x08
)

const (
	// ServiceError contains error info of service invocation
	ServiceError = "__rpcx_error__"
)

type MessageType byte

const (
	Request  MessageType = iota
	Response
)

type CompressType byte

const (
	// None does not compress.
	None CompressType = iota
	// Gzip uses gzip compression.
	Gzip
)

type MessageStatusType byte

const (
	// Normal is normal requests and responses.
	Normal MessageStatusType = iota
	// Error indicates some errors occur.
	Error
)

type SerializeType byte

const (
	// Json
	JSON SerializeType = iota
	// ProtoBuffer
	ProtoBuffer
	// MsgPack
	MsgPack
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

func NewMessage() *Message {
	header := Header([12]byte{})
	header[0] = magicNumber

	return &Message{
		Header: &header,
	}
}

// CheckMagicNumber checks whether header starts rpcx magic number.
func (h Header) CheckMagicNumber() bool {
	return h[0] == magicNumber
}

// Version returns version of rpcx protocol.
func (h Header) Version() byte {
	return h[1]
}

// SetVersion sets version for this header.
func (h *Header) SetVersion(v byte) {
	h[1] = v
}

// MessageType returns the message type.
func (h Header) MessageType() MessageType {
	return MessageType(h[2]&0x80) >> 7
}

// SetMessageType sets message type.
func (h *Header) SetMessageType(mt MessageType) {
	h[2] = h[2] | (byte(mt) << 7)
}

// IsHeartbeat returns whether the message is heartbeat message.
func (h Header) IsHeartbeat() bool {
	return h[2]&0x40 == 0x40
}

// SetHeartbeat sets the heartbeat flag.
func (h *Header) SetHeartbeat(hb bool) {
	if hb {
		h[2] = h[2] | 0x40
	} else {
		h[2] = h[2] &^ 0x40
	}
}

// IsOneway returns whether the message is one-way message.
// If true, server won't send responses.
func (h Header) IsOneway() bool {
	return h[2]&0x20 == 0x20
}

// SetOneway sets the oneway flag.
func (h *Header) SetOneway(oneway bool) {
	if oneway {
		h[2] = h[2] | 0x20
	} else {
		h[2] = h[2] &^ 0x20
	}
}

// CompressType returns compression type of messages.
func (h Header) CompressType() CompressType {
	return CompressType((h[2] & 0x1C) >> 2)
}

// SetCompressType sets the compression type.
func (h *Header) SetCompressType(ct CompressType) {
	h[2] = h[2] | ((byte(ct) << 2) & 0x1C)
}

// MessageStatusType returns the message status type.
func (h Header) MessageStatusType() MessageStatusType {
	return MessageStatusType(h[2] & 0x03)
}

// SetMessageStatusType sets message status type.
func (h *Header) SetMessageStatusType(mt MessageStatusType) {
	h[2] = h[2] | (byte(mt) & 0x03)
}

// SerializeType returns serialization type of payload.
func (h Header) SerializeType() SerializeType {
	return SerializeType((h[3] & 0xF0) >> 4)
}

// SetSerializeType sets the serialization type.
func (h *Header) SetSerializeType(st SerializeType) {
	h[3] = h[3] | (byte(st) << 4)
}

// Seq returns sequence number of messages.
func (h Header) Seq() uint64 {
	return binary.BigEndian.Uint64(h[4:])
}

// SetSeq sets  sequence number.
func (h *Header) SetSeq(seq uint64) {
	binary.BigEndian.PutUint64(h[4:], seq)
}

func (m Message) Copy() *Message {
	header := *m.Header
	c := GetMsgs()
	c.Header = &header
	c.ServicePath = m.ServicePath
	c.ServiceMethod = m.ServiceMethod
	return c
}

func (message *Message) Decode(reader io.Reader) error {
	// parse header
	_, err := io.ReadFull(reader, message.Header[:])
	if err != nil {
		return err
	}
	p := poolUint32Data.Get()
	lenData := p.(*[]byte)
	_, err = io.ReadFull(reader, *lenData)
	if err != nil {
		poolUint32Data.Put(lenData)
		return err
	}
	l := binary.BigEndian.Uint32(*lenData)
	poolUint32Data.Put(lenData)

	if MaxMessageLength > 0 && int(l) > MaxMessageLength {
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
	l = binary.BigEndian.Uint32(data[n: n+4])
	n = n + 4
	nEnd = n + int(l)
	message.ServiceMethod = util.ByteToString(data[n:nEnd])
	n = nEnd

	// parse meta
	l = binary.BigEndian.Uint32(data[n: n+4])
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
	l = binary.BigEndian.Uint32(data[n: n+4])
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
		sl := binary.BigEndian.Uint32(data[n: n+4])
		n = n + 4
		if n+sl > l-4 {
			return m, ErrMetaKVMissing
		}
		k := util.ByteToString(data[n: n+sl])
		n = n + sl

		// value
		sl = binary.BigEndian.Uint32(data[n: n+4])
		n = n + 4
		if n+sl > l {
			return m, ErrMetaKVMissing
		}
		v := util.ByteToString(data[n: n+sl])
		n = n + sl
		m[k] = v
	}

	return m, nil
}

func (m *Message) Encode() []byte {
	meta := encodeMetadata(m.Metadata)

	spl := len(m.ServicePath)
	sml := len(m.ServiceMethod)

	allL := (4 + spl) + (4 + sml) + (4 + len(meta)) + (4 + len(m.Payload))
	metaStart := 12 + 4 + (4 + spl) + (4 + sml)
	payLoadStart := metaStart + (4 + len(meta))
	l := 12 + 4 + allL
	data := make([]byte, l)
	copy(data, m.Header[:])

	binary.BigEndian.PutUint32(data[12:16], uint32(allL))

	binary.BigEndian.PutUint32(data[16:20], uint32(spl))
	copy(data[20:20+spl], util.StringToByte(m.ServicePath))

	binary.BigEndian.PutUint32(data[20+spl:24+spl], uint32(sml))
	copy(data[24+spl:metaStart], util.StringToByte(m.ServiceMethod))

	binary.BigEndian.PutUint32(data[metaStart:metaStart+4], uint32(len(meta)))
	copy(data[metaStart+4:], meta)

	binary.BigEndian.PutUint32(data[payLoadStart:payLoadStart+4], uint32(len(m.Payload)))
	copy(data[payLoadStart+4:], m.Payload)

	return data
}

func encodeMetadata(m map[string]string) []byte {
	if len(m) == 0 {
		return []byte{}
	}
	var buf bytes.Buffer
	var d = make([]byte, 4)
	for k, v := range m {
		binary.BigEndian.PutUint32(d, uint32(len(k)))
		buf.Write(d)
		buf.Write(util.StringToByte(k))
		binary.BigEndian.PutUint32(d, uint32(len(v)))
		buf.Write(d)
		buf.Write(util.StringToByte(v))
	}
	return buf.Bytes()
}

func (m *Message) Reset() {
	resetHeader(m.Header)
	m.Metadata = nil
	m.Payload = m.Payload[:0]
	m.data = m.data[:0]
	m.ServicePath = ""
	m.ServiceMethod = ""
}

var zeroHeaderArray Header
var zeroHeader = zeroHeaderArray[1:]

func resetHeader(h *Header) {
	copy(h[1:], zeroHeader)
}
