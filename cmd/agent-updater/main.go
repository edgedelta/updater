package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/edgedelta/updater"

	"github.com/rs/zerolog/log"
)

var (
	configPath = flag.String("config", "", "Local config path")
)

func main() {
	flag.Parse()
	if err := validateFlags(); err != nil {
		log.Error().Msgf("Failed to validate the flags, err: %v", err)
		os.Exit(1)
	}
	ctx := context.Background()
	updater, err := updater.NewUpdater(ctx, *configPath)
	if err != nil {
		log.Error().Msgf("Failed to construct new Updater, err: %v", err)
		os.Exit(1)
	}
	if err := updater.Run(ctx); err != nil {
		log.Error().Msgf("Runtime error occured, err: %v", err)
		os.Exit(1)
	}
}

func validateFlags() error {
	if *configPath == "" {
		return fmt.Errorf("--config must be specified")
	}
	return nil
}
