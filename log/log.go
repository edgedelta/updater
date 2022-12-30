package log

import (
	"io"
	"os"
	"sync/atomic"

	"github.com/edgedelta/updater/core"

	"github.com/rs/zerolog"
)

var (
	logger   atomic.Pointer[zerolog.Logger]
	selfInfo *core.RuntimeInfo
)

func init() {
	logger.Store(newLogger(os.Stdout))
	selfInfo = core.GetSelfInfo()
}

func SetWriters(wrs ...io.Writer) {
	l := newLogger(wrs...)
	logger.Swap(l)
}

func Debug(format string, args ...any) {
	logger.Load().Debug().Str("k8s_namespace", selfInfo.Namespace).
		Str("k8s_node", selfInfo.Node).Str("k8s_pod", selfInfo.Pod).
		Msgf(format, args)
}

func Info(format string, args ...any) {
	logger.Load().Info().Str("k8s_namespace", selfInfo.Namespace).
		Str("k8s_node", selfInfo.Node).Str("k8s_pod", selfInfo.Pod).
		Msgf(format, args)
}

func Error(err error, format string, args ...any) {
	logger.Load().Err(err).Str("k8s_namespace", selfInfo.Namespace).
		Str("k8s_node", selfInfo.Node).Str("k8s_pod", selfInfo.Pod).
		Stack().Msgf(format, args)
}

func Fatal(format string, args ...any) {
	logger.Load().Fatal().Str("k8s_namespace", selfInfo.Namespace).
		Str("k8s_node", selfInfo.Node).Str("k8s_pod", selfInfo.Pod).
		Msgf(format, args)
}

func newLogger(wrs ...io.Writer) *zerolog.Logger {
	multi := zerolog.MultiLevelWriter(wrs...)
	l := zerolog.New(multi).With().Timestamp().Logger()
	return &l
}
