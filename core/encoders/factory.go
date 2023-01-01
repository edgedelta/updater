package encoders

import (
	"fmt"
	"io"

	"github.com/edgedelta/updater/core"
)

type Encoder interface {
	Write([]interface{}) error
	Close() error
}

func New(writer io.Writer, encoding *core.EncodingConfig) (Encoder, error) {
	switch encoding.Type {
	case core.EncodingJSON:
		return NewJSONEncoder(writer), nil
	case core.EncodingRaw:
		return NewDelimitedRawEncoder(writer, encoding.Opts), nil
	}
	return nil, fmt.Errorf("unknown encoding type: %q", encoding.Type)
}
