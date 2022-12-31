package compressors

import "io"

type NoOpCompressor struct {
	wr io.Writer
}

func NewNoOpCompressor(wr io.Writer) *NoOpCompressor {
	return &NoOpCompressor{wr: wr}
}

func (c *NoOpCompressor) Write(b []byte) (n int, err error) {
	return c.wr.Write(b)
}

func (c *NoOpCompressor) Close() error {
	return nil
}
