package core

import (
	"frpc/codec"
	"frpc/protocol"
)

const DefaultRPCPath = "/_fprc_"

var (
	Codecs = map[protocol.SerializeType]codec.Codec{
		protocol.JSON:        &codec.JSONCodec{},
		protocol.ProtoBuffer: &codec.PBCodec{},
		protocol.MsgPack:     &codec.MsgpackCodec{},
	}
)

type ContextKey string

// ReqMetaDataKey is used to set metatdata in context of requests.
var ReqMetaDataKey = ContextKey("__req_metadata")

// ResMetaDataKey is used to set metatdata in context of responses.
var ResMetaDataKey = ContextKey("__res_metadata")

var StartRequestContextKey = ContextKey("start-parse-request")
