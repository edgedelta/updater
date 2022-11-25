package main

import (
	"context"
	"os"

	"github.com/edgedelta/updater"

	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()
	path := os.Getenv("KO_DATA_PATH")
	if path == "" {
		log.Error().Msgf("KO_DATA_PATH is not set")
		os.Exit(1)
	}
	updater, err := updater.NewUpdater(ctx, path+"/config.yml")
	if err != nil {
		log.Error().Msgf("Failed to construct new Updater, err: %v\n", err)
		os.Exit(1)
	}
	if err := updater.Run(ctx); err != nil {
		log.Error().Msgf("Runtime error occured, err: %v\n", err)
		os.Exit(1)
	}
}
