package main

import (
	"dungeons/app/server"
	"os"

	"github.com/rs/zerolog/log"
)

func main() {
	if err := newDungeonsServer(); err != nil {
		log.Fatal().Err(err).Msg("Unable to create new server")
		os.Exit(51)
	}
	log.Debug().Msg("API launched with human readable log")

	srv := server.GetServer()
	srv.ListenAndServe()
}
