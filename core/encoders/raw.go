package encoders

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/edgedelta/updater/core"
)

var (
	tabRe     = regexp.MustCompile(`\\t`)
	carrRe    = regexp.MustCompile(`\\r`)
	newlineRe = regexp.MustCompile(`\\n`)
)

type DelimitedRawEncoder struct {
	wr    io.Writer
	delim string
}

func NewDelimitedRawEncoder(wr io.Writer, opts *core.EncodingOptions) *DelimitedRawEncoder {
	delim := ""
	if opts != nil {
		delim = opts.Delimiter
	}
	delim = unescape(delim)
	return &DelimitedRawEncoder{wr: wr, delim: delim}
}

func (e *DelimitedRawEncoder) Write(objects []interface{}) error {
	var sb strings.Builder
	for _, o := range objects {
		sb.WriteString(strings.TrimSpace(fmt.Sprintf("%v", o)))
		sb.WriteString(e.delim)
	}
	_, err := e.wr.Write([]byte(sb.String()))
	return err
}

func (e *DelimitedRawEncoder) Close() error {
	return nil
}

func unescape(s string) string {
	ss := tabRe.ReplaceAllString(s, "\t")
	ss = carrRe.ReplaceAllString(ss, "\r")
	return newlineRe.ReplaceAllString(ss, "\n")
}
