package codec

import (
	"io"
)

type Header struct {
	ServiceMethod string
	Seq           uint64
	Err           error
}

type Codec interface {
	io.Closer
	ReadHeader(header *Header) error
	ReadBody(v interface{}) error
	Write(header *Header, v interface{}) error
}

// NewCodecFunc Codec的构造方法
type NewCodecFunc func(closer io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
