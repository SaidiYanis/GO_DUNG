package seed

import (
	"context"
	"dungeons/app/mongodb"
	"dungeons/app/server"
	"errors"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func ExecuteFromEnv() {
	if os.Getenv("MODE") == "" {
		if err := godotenv.Load(); err != nil {
			var pathErr *os.PathError
			if !errors.As(err, &pathErr) {
				log.Fatal().Err(err).Msg("load env")
			}
		}
	}

	srv := &server.Dungeons{}
	srv.ParseParameters()

	ctx, cancel := context.WithTimeout(context.Background(), srv.DBTimeout)
	defer cancel()

	client, err := mongodb.OpenMongoDB(ctx, srv.DBHost)
	if err != nil {
		log.Fatal().Err(err).Msg("connect mongo")
	}
	defer client.Disconnect(context.Background())

	if err := Run(context.Background(), client.Database(srv.DBName), srv.DBTimeout); err != nil {
		log.Fatal().Err(err).Msg("seed failed")
	}
	log.Info().Msg("bootloader completed")
}
