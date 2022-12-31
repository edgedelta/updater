package encoders

import (
	"fmt"
	"io"
)

type Encoder interface {
	Write([]any) error
	Close() error
}

type EncodingType string

const (
	EncodingJSON EncodingType = "json"
)

func New(writer io.Writer, encoding EncodingType) (Encoder, error) {
	switch encoding {
	case EncodingJSON:
		return NewJSONEncoder(writer), nil
	}
	return nil, fmt.Errorf("unknown encoding type: %q", encoding)
}
