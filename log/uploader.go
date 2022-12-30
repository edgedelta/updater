package log

import (
	"context"
	"io"
	"log"
	"sync/atomic"
	"time"

	"github.com/edgedelta/updater/api"

	zerolog "github.com/rs/zerolog/log"
)

const (
	uploaderChanBufferSize = 100
	uploaderFlushInterval  = 5 * time.Second
)

type Uploader struct {
	name       string
	incoming   chan []byte
	buffer     []string
	cl         *api.Client
	isRunning  int32
	stopDoneCh chan struct{}
	stopCh     chan struct{}
}

type incomingLogWriter struct {
	buffer chan<- []byte
}

func (w *incomingLogWriter) Write(b []byte) (int, error) {
	w.buffer <- b
	return len(b), nil
}

func NewUploader(ctx context.Context, name string, cl *api.Client) *Uploader {
	u := &Uploader{
		name:      name,
		incoming:  make(chan []byte, uploaderChanBufferSize),
		buffer:    make([]string, 0),
		cl:        cl,
		isRunning: 0,
	}
	return u
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

func (u *Uploader) Stop() {
	if atomic.CompareAndSwapInt32(&u.isRunning, 1, 0) {
		u.stop()
	}
}

func (u *Uploader) StopBlocking() <-chan struct{} {
	if atomic.CompareAndSwapInt32(&u.isRunning, 1, 0) {
		u.stop()
		return u.stopDoneCh
	}
	ch := make(chan struct{})
	ch <- struct{}{}
	return ch
}

func (u *Uploader) stop() {
	zerolog.Debug().Msgf("Stopping log.Uploader %s", u.name)
	u.stopCh <- struct{}{}
}

func (u *Uploader) run() {
	ticker := time.NewTicker(uploaderFlushInterval)
	for {
		select {
		case <-u.stopCh:
			u.drain()
			u.stopDoneCh <- struct{}{}
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

func (u *Uploader) process(l []byte) {
	u.buffer = append(u.buffer, string(l))
}

func (u *Uploader) flush() {
	size := len(u.buffer)
	if err := u.cl.UploadLogs(u.buffer); err != nil {
		log.Fatalf("api.Client.UploadLogs: %v", err)
	}
	u.buffer = make([]string, 0)
	zerolog.Debug().Msgf("log.Uploader %s flushed %d lines of logs", u.name, size)
}
