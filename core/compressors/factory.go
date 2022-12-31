package compressors

import (
	"compress/gzip"
	"fmt"
	"io"
)

type Compressor interface {
	io.WriteCloser
}

type CompressionType string

const (
	CompressionGzip CompressionType = "gzip"
	CompressionNoOp                 = ""
)

func New(writer io.Writer, compression CompressionType) (Compressor, error) {
	switch compression {
	case CompressionGzip:
		return gzip.NewWriter(writer), nil
	case CompressionNoOp:
		return NewNoOpCompressor(writer), nil
	}
	return nil, fmt.Errorf("unknown compression type: %q", compression)
}
