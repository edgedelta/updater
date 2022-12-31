package encoders

import (
	"encoding/json"
	"fmt"
	"io"
)

type JSONEncoder struct {
	wr io.Writer
}

func NewJSONEncoder(wr io.Writer) *JSONEncoder {
	return &JSONEncoder{wr: wr}
}

func (e *JSONEncoder) Write(objects []any) error {
	encoder := json.NewEncoder(e.wr)
	for _, o := range objects {
		if err := encoder.Encode(o); err != nil {
			return fmt.Errorf("json.Encoder.Encode: %v", err)
		}
	}
	return nil
}

func (e *JSONEncoder) Close() error {
	return nil
}
