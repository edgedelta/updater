package compressors

import (
	"compress/gzip"
	"fmt"
	"io"

	"github.com/edgedelta/updater/core"
)

type Compressor interface {
	io.WriteCloser
	Flush() error
}

func New(writer io.Writer, compression core.CompressionType) (Compressor, error) {
	switch compression {
	case core.CompressionGzip:
		return gzip.NewWriter(writer), nil
	case core.CompressionNoOp:
		return NewNoOpCompressor(writer), nil
	}
	return nil, fmt.Errorf("unknown compression type: %q", compression)
}
