package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/edgedelta/updater"
	"github.com/edgedelta/updater/log"
	"github.com/edgedelta/updater/loguploader"
)

var (
	configPath  = flag.String("config", "", "Local config path")
	logUploader *loguploader.Uploader
)

const (
	gracefulShutdownPeriod = time.Minute
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(fmt.Errorf("%v", err), "Main goroutine panicked")
		}
		handleGracefulShutdown()
	}()
	flag.Parse()
	if err := validateFlags(); err != nil {
		log.Fatal("Failed to validate the flags, err: %v", err)
	}
	ctx := context.Background()
	updater, err := updater.NewUpdater(ctx, *configPath)
	if err != nil {
		log.Fatal("Failed to construct new Updater, err: %v", err)
	}
	if updater.LogUploaderEnabled() {
		logUploader = loguploader.New(ctx, "self_log_uploader", updater.APIClient())
		log.SetWriters(os.Stdout, logUploader.Writer())
		logUploader.Run()
	}
	if err := updater.Run(ctx); err != nil {
		log.Error(err, "Runtime error")
	}
}

func validateFlags() error {
	if *configPath == "" {
		return errors.New("--config must be specified")
	}
	return nil
}

func handleGracefulShutdown() {
	if logUploader == nil {
		log.Info("Log uploader is not running, exiting")
		return
	}
	// It's important to first remove the log uploader's writer from logger and then
	// stop the log uploader to prevent memory leak
	log.SetWriters(os.Stdout)
	logUploaderStopped := logUploader.StopBlocking()

	log.Info("Shutdown period %.0fm started", gracefulShutdownPeriod.Minutes())
	t := time.NewTimer(gracefulShutdownPeriod)

	for {
		select {
		case <-logUploaderStopped:
			log.Info("Log uploader %s stopped", logUploader.Name())
			return
		case <-t.C:
			log.Warn("Could not stop log uploader %s within the graceful shutdown period (%.0fm)", logUploader.Name(), gracefulShutdownPeriod.Minutes())
			return
		}
	}
}
