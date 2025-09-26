package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shotgun-cli/shotgun-cli/cmd"
)

func main() {
	// Set up zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Execute the root command
	if err := cmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Failed to execute command")
		os.Exit(1)
	}
}