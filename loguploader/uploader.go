package loguploader

import (
	"context"
	"io"
	"sync/atomic"
	"time"

	"github.com/edgedelta/updater/api"

	zerolog "github.com/rs/zerolog/log"
)

const (
	uploaderChanBufferSize = 100
	uploaderFlushInterval  = time.Minute
)

type Uploader struct {
	name       string
	incoming   chan string
	buffer     []string
	cl         *api.Client
	isRunning  int32
	stopDoneCh chan struct{}
	stopCh     chan struct{}
}

type incomingLogWriter struct {
	buffer chan<- string
}

func (w *incomingLogWriter) Write(b []byte) (int, error) {
	w.buffer <- string(b)
	return len(b), nil
}

func New(ctx context.Context, name string, cl *api.Client) *Uploader {
	return &Uploader{
		name:      name,
		incoming:  make(chan string, uploaderChanBufferSize),
		buffer:    make([]string, 0),
		cl:        cl,
		isRunning: 0,
	}
}

func (u *Uploader) Name() string {
	return u.name
}

func (u *Uploader) Writer() io.Writer {
	return &incomingLogWriter{buffer: u.incoming}
}

func (u *Uploader) Run() {
	if atomic.CompareAndSwapInt32(&u.isRunning, 0, 1) {
		u.stopCh = make(chan struct{})
		u.stopDoneCh = make(chan struct{})
		go u.run()
		zerolog.Debug().Msgf("log.Uploader %s started running", u.name)
	}
}

func (u *Uploader) StopBlocking() <-chan struct{} {
	if atomic.CompareAndSwapInt32(&u.isRunning, 1, 0) {
		u.stop()
		return u.stopDoneCh
	}
	ch := make(chan struct{})
	close(ch)
	return ch
}

func (u *Uploader) stop() {
	zerolog.Debug().Msgf("Stopping log.Uploader %s", u.name)
	close(u.incoming)
	close(u.stopCh)
}

func (u *Uploader) run() {
	ticker := time.NewTicker(uploaderFlushInterval)
	for {
		select {
		case <-u.stopCh:
			zerolog.Debug().Msgf("log.Uploader %s got stop signal, will drain remaining logs", u.name)
			u.drain()
			close(u.stopDoneCh)
			return
		case l := <-u.incoming:
			u.process(l)
		case <-ticker.C:
			u.flush()
		}
	}
}

func (u *Uploader) drain() {
	for l := range u.incoming {
		u.process(l)
	}
	u.flush()
}

func (u *Uploader) process(l string) {
	u.buffer = append(u.buffer, l)
}

func (u *Uploader) flush() {
	size := len(u.buffer)
	b := make([]interface{}, 0, len(u.buffer))
	for _, it := range u.buffer {
		b = append(b, it)
	}
	if err := u.cl.UploadLogs(b); err != nil {
		zerolog.Fatal().Msgf("api.Client.UploadLogs: %v", err)
	}
	u.buffer = make([]string, 0)
	zerolog.Debug().Msgf("log.Uploader %s flushed %d lines of logs", u.name, size)
}
