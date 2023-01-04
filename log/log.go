package log

import (
	"io"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/edgedelta/updater/core"

	"github.com/rs/zerolog"
)

var (
	logger   atomic.Pointer[zerolog.Logger]
	selfInfo *core.RuntimeInfo
)

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	logger.Store(newLogger(os.Stdout))

	var err error
	selfInfo, err = core.GetSelfInfo()
	if err != nil {
		log.Fatalf("core.GetSelfInfo: %v\n", err)
	}
}

func SetWriters(wrs ...io.Writer) {
	l := newLogger(wrs...)
	logger.Swap(l)
}

func Debug(format string, args ...any) {
	logger.Load().Debug().Str("k8s_namespace", selfInfo.Namespace).
		Str("k8s_node", selfInfo.Node).Str("k8s_pod", selfInfo.Pod).
		Msgf(format, args...)
}

func Info(format string, args ...any) {
	logger.Load().Info().Str("k8s_namespace", selfInfo.Namespace).
		Str("k8s_node", selfInfo.Node).Str("k8s_pod", selfInfo.Pod).
		Msgf(format, args...)
}

func Warn(format string, args ...any) {
	logger.Load().Warn().Str("k8s_namespace", selfInfo.Namespace).
		Str("k8s_node", selfInfo.Node).Str("k8s_pod", selfInfo.Pod).
		Msgf(format, args...)
}

func Error(err error, format string, args ...any) {
	logger.Load().Err(err).Str("k8s_namespace", selfInfo.Namespace).
		Str("k8s_node", selfInfo.Node).Str("k8s_pod", selfInfo.Pod).
		Stack().Msgf(format, args...)
}

func Fatal(format string, args ...any) {
	logger.Load().Fatal().Str("k8s_namespace", selfInfo.Namespace).
		Str("k8s_node", selfInfo.Node).Str("k8s_pod", selfInfo.Pod).
		Msgf(format, args...)
}

func newLogger(wrs ...io.Writer) *zerolog.Logger {
	multi := zerolog.MultiLevelWriter(wrs...)
	l := zerolog.New(multi).With().Timestamp().Logger().Level(zerolog.DebugLevel)
	return &l
}
