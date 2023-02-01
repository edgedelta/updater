package log

import (
	"io"
	"os"
	"sync/atomic"

	"github.com/rs/zerolog"
)

var (
	logger     atomic.Pointer[zerolog.Logger]
	customTags atomic.Pointer[map[string]string]
	errCount   int32
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.TimestampFieldName = "timestamp"
	zerolog.MessageFieldName = "raw"

	errCount = 0

	logger.Store(newLogger(os.Stdout))
	customTags.Store(&map[string]string{})
}

func SetWriters(wrs ...io.Writer) {
	l := newLogger(wrs...)
	logger.Swap(l)
}

func SetCustomTags(m map[string]string) {
	customTags.Store(&m)
}

func Debug(format string, args ...any) {
	l := logger.Load().Debug()
	for k, v := range *customTags.Load() {
		l.Str(k, v)
	}
	l.Msgf(format, args...)
}

func Info(format string, args ...any) {
	l := logger.Load().Info()
	for k, v := range *customTags.Load() {
		l.Str(k, v)
	}
	l.Msgf(format, args...)
}

func Warn(format string, args ...any) {
	l := logger.Load().Warn()
	for k, v := range *customTags.Load() {
		l.Str(k, v)
	}
	l.Msgf(format, args...)
}

func Error(format string, args ...any) {
	l := logger.Load().Error()
	for k, v := range *customTags.Load() {
		l.Str(k, v)
	}
	l.Msgf(format, args...)
	atomic.AddInt32(&errCount, 1)
}

func Fatal(format string, args ...any) {
	l := logger.Load().Fatal()
	for k, v := range *customTags.Load() {
		l.Str(k, v)
	}
	l.Msgf(format, args...)
}

func ErrorCount() int32 {
	return atomic.LoadInt32(&errCount)
}

func newLogger(wrs ...io.Writer) *zerolog.Logger {
	multi := zerolog.MultiLevelWriter(wrs...)
	l := zerolog.New(multi).With().Timestamp().Logger().Level(zerolog.DebugLevel)
	return &l
}
