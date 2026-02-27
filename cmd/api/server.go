package main

import (
	"context"
	"dungeons/app/auth"
	auctioncontroller "dungeons/app/controllers/auction"
	dungeoncontroller "dungeons/app/controllers/dungeon"
	inventorycontroller "dungeons/app/controllers/inventory"
	playercontroller "dungeons/app/controllers/player"
	runcontroller "dungeons/app/controllers/run"
	"dungeons/app/mongodb"
	auctionrepo "dungeons/app/repositories/auction"
	dungeonrepo "dungeons/app/repositories/dungeon"
	inventoryrepo "dungeons/app/repositories/inventory"
	playerrepo "dungeons/app/repositories/player"
	runrepo "dungeons/app/repositories/run"
	auctionroutes "dungeons/app/routes/auction"
	dungeonroutes "dungeons/app/routes/dungeon"
	inventoryroutes "dungeons/app/routes/inventory"
	playerroutes "dungeons/app/routes/player"
	runroutes "dungeons/app/routes/run"
	"dungeons/app/seed"
	"dungeons/app/server"
	auctionservice "dungeons/app/services/auction"
	dungeonservice "dungeons/app/services/dungeon"
	inventoryservice "dungeons/app/services/inventory"
	playerservice "dungeons/app/services/player"
	runservice "dungeons/app/services/run"
	"errors"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func newDungeonsServer() error {
	if os.Getenv("MODE") == "" {
		if err := godotenv.Load(); err != nil {
			var pathErr *os.PathError
			if !errors.As(err, &pathErr) {
				return err
			}
		}
	}

	srv := &server.Dungeons{}
	srv.ParseParameters()

	switch srv.LogFormat {
	case "HUMAN":
		log.Logger = log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	case "JSON":
	default:
		log.Logger = log.Logger.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true})
	}

	srv.Router = setupRouter()

	ctx, cancel := context.WithTimeout(context.Background(), srv.DBTimeout)
	defer cancel()

	client, err := mongodb.OpenMongoDB(ctx, srv.DBHost)
	if err != nil {
		return err
	}
	srv.MongoClient = client
	srv.Database = client.Database(srv.DBName)

	validate := validator.New()

	playerRepository := playerrepo.NewMongoRepository(srv.Database, srv.DBTimeout)
	dungeonRepository := dungeonrepo.NewMongoRepository(srv.Database, srv.DBTimeout)
	runRepository := runrepo.NewMongoRepository(srv.Database, srv.DBTimeout)
	inventoryRepository := inventoryrepo.NewMongoRepository(srv.Database, srv.DBTimeout)
	auctionRepository := auctionrepo.NewMongoRepository(srv.Database, srv.DBTimeout)

	playerSvc := playerservice.New(playerRepository, validate, playerservice.NewHMACTokenSigner(srv.TokenKey), srv.TokenTTL)
	dungeonSvc := dungeonservice.New(dungeonRepository, validate)
	runSvc := runservice.New(runRepository, dungeonRepository, playerRepository, inventoryRepository, validate, srv.MongoClient)
	inventorySvc := inventoryservice.New(inventoryRepository)
	auctionSvc := auctionservice.New(auctionRepository, inventoryRepository, playerRepository, validate, srv.MongoClient)

	for _, ensure := range []func(context.Context) error{
		playerSvc.EnsureIndexes,
		dungeonSvc.EnsureIndexes,
		runSvc.EnsureIndexes,
		inventorySvc.EnsureIndexes,
		auctionSvc.EnsureIndexes,
	} {
		if err := ensure(context.Background()); err != nil {
			return err
		}
	}

	if srv.SeedOnBoot {
		if err := seed.Run(context.Background(), srv.Database, srv.DBTimeout); err != nil {
			return err
		}
	}

	playerHandler := playercontroller.New(playerSvc)
	dungeonHandler := dungeoncontroller.New(dungeonSvc)
	runHandler := runcontroller.New(runSvc)
	inventoryHandler := inventorycontroller.New(inventorySvc)
	auctionHandler := auctioncontroller.New(auctionSvc)

	authMiddleware := auth.RequireAuth(srv.TokenKey)
	v1 := srv.Router.Group("/v1")
	playerroutes.SetupRouter(v1, playerHandler, authMiddleware)
	dungeonroutes.SetupRouter(v1, dungeonHandler, authMiddleware)
	runroutes.SetupRouter(v1, runHandler, authMiddleware)
	inventoryroutes.SetupRouter(v1, inventoryHandler, authMiddleware)
	auctionroutes.SetupRouter(v1, auctionHandler, authMiddleware)

	server.SetServer(srv)
	return nil
}
