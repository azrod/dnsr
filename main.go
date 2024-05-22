package main

import (
	"os"

	"github.com/azrod/dnsr/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	log.Level(zerolog.DebugLevel)
	cmd.Execute()
}
