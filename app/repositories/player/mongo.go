package player

import (
	"context"
	apperrors "dungeons/app/errors"
	"dungeons/app/models"
	"dungeons/app/mongodb"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const collectionName = "players"

type MongoRepository struct {
	db      *mongo.Database
	timeout time.Duration
}

func NewMongoRepository(db *mongo.Database, timeout time.Duration) *MongoRepository {
	return &MongoRepository{db: db, timeout: timeout}
}

func (r *MongoRepository) EnsureIndexes(ctx context.Context) error {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.db.Collection(collectionName).Indexes().CreateMany(cctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "display_name", Value: 1}}},
		{Keys: bson.D{{Key: "customID", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
	if err != nil {
		return fmt.Errorf("ensure player indexes: %w", err)
	}
	return nil
}

func (r *MongoRepository) Create(ctx context.Context, p models.Player) error {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.db.Collection(collectionName).InsertOne(cctx, p)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate player: %w", apperrors.ErrConflict)
		}
		return fmt.Errorf("insert player: %w", err)
	}
	return nil
}

func (r *MongoRepository) GetByID(ctx context.Context, id string) (models.Player, error) {
	var p models.Player
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	if err := r.db.Collection(collectionName).FindOne(cctx, bson.M{"customID": id}).Decode(&p); err != nil {
		if err == mongo.ErrNoDocuments {
			return p, fmt.Errorf("player id %s: %w", id, apperrors.ErrNotFound)
		}
		return p, fmt.Errorf("find player by id: %w", err)
	}
	return p, nil
}

func (r *MongoRepository) GetByEmail(ctx context.Context, email string) (models.Player, error) {
	var p models.Player
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	if err := r.db.Collection(collectionName).FindOne(cctx, bson.M{"email": email}).Decode(&p); err != nil {
		if err == mongo.ErrNoDocuments {
			return p, fmt.Errorf("player email %s: %w", email, apperrors.ErrNotFound)
		}
		return p, fmt.Errorf("find player by email: %w", err)
	}
	return p, nil
}

func (r *MongoRepository) List(ctx context.Context, params models.QueryParams) ([]models.Player, error) {
	q := params.Normalize()
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()

	cursor, err := r.db.Collection(collectionName).Find(cctx, bson.M{}, options.Find().SetSkip(q.Skip()).SetLimit(q.Limit).SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("list players: %w", err)
	}
	defer cursor.Close(cctx)

	players := make([]models.Player, 0)
	for cursor.Next(cctx) {
		var p models.Player
		if err := cursor.Decode(&p); err != nil {
			return nil, fmt.Errorf("decode player: %w", err)
		}
		players = append(players, p)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("player cursor: %w", err)
	}
	return players, nil
}

func (r *MongoRepository) UpdateDisplayName(ctx context.Context, id, displayName string, updatedAt time.Time) (models.Player, error) {
	var updated models.Player
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()

	err := r.db.Collection(collectionName).FindOneAndUpdate(
		cctx,
		bson.M{"customID": id},
		bson.M{"$set": bson.M{"display_name": displayName, "updated_at": updatedAt}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return updated, fmt.Errorf("player id %s: %w", id, apperrors.ErrNotFound)
		}
		return updated, fmt.Errorf("update player: %w", err)
	}
	return updated, nil
}

func (r *MongoRepository) IncrementGold(ctx context.Context, id string, delta int64, updatedAt time.Time) (models.Player, error) {
	var updated models.Player
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()

	err := r.db.Collection(collectionName).FindOneAndUpdate(
		cctx,
		bson.M{"customID": id},
		bson.M{"$inc": bson.M{"gold": delta}, "$set": bson.M{"updated_at": updatedAt}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return updated, fmt.Errorf("player id %s: %w", id, apperrors.ErrNotFound)
		}
		return updated, fmt.Errorf("increment gold: %w", err)
	}
	return updated, nil
}

func (r *MongoRepository) SetGold(ctx context.Context, id string, gold int64, updatedAt time.Time) (models.Player, error) {
	var updated models.Player
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()

	err := r.db.Collection(collectionName).FindOneAndUpdate(
		cctx,
		bson.M{"customID": id},
		bson.M{"$set": bson.M{"gold": gold, "updated_at": updatedAt}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return updated, fmt.Errorf("player id %s: %w", id, apperrors.ErrNotFound)
		}
		return updated, fmt.Errorf("set gold: %w", err)
	}
	return updated, nil
}
