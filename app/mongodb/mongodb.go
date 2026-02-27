package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func OpenMongoDB(ctx context.Context, uri string) (*mongo.Client, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("ping mongo: %w", err)
	}
	return client, nil
}

func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return context.WithTimeout(parent, timeout)
}

func WithTransaction(ctx context.Context, client *mongo.Client, fn func(context.Context) error) error {
	session, err := client.StartSession()
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}
	defer session.EndSession(ctx)

	if err := mongo.WithSession(ctx, session, func(sc context.Context) error {
		if err := session.StartTransaction(); err != nil {
			return fmt.Errorf("start transaction: %w", err)
		}
		if err := fn(sc); err != nil {
			_ = session.AbortTransaction(sc)
			return err
		}
		if err := session.CommitTransaction(sc); err != nil {
			return fmt.Errorf("commit transaction: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("transaction body: %w", err)
	}

	return nil
}
