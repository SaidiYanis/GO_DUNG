package dungeon

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

const (
	dungeonsCollection = "dungeons"
	stepsCollection    = "boss_steps"
)

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

	if _, err := r.db.Collection(dungeonsCollection).Indexes().CreateMany(cctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "createdBy", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
	}); err != nil {
		return fmt.Errorf("dungeon indexes: %w", err)
	}

	if _, err := r.db.Collection(stepsCollection).Indexes().CreateMany(cctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "dungeonId", Value: 1}, {Key: "order", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "dungeonId", Value: 1}}},
	}); err != nil {
		return fmt.Errorf("step indexes: %w", err)
	}
	return nil
}

func (r *MongoRepository) CreateDungeon(ctx context.Context, d models.Dungeon) error {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.db.Collection(dungeonsCollection).InsertOne(cctx, d)
	if err != nil {
		return fmt.Errorf("insert dungeon: %w", err)
	}
	return nil
}

func (r *MongoRepository) UpdateDungeon(ctx context.Context, d models.Dungeon) (models.Dungeon, error) {
	var out models.Dungeon
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()

	err := r.db.Collection(dungeonsCollection).FindOneAndReplace(cctx, bson.M{"_id": d.ID}, d, options.FindOneAndReplace().SetReturnDocument(options.After)).Decode(&out)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return out, fmt.Errorf("dungeon id %s: %w", d.ID, apperrors.ErrNotFound)
		}
		return out, fmt.Errorf("update dungeon: %w", err)
	}
	return out, nil
}

func (r *MongoRepository) GetDungeonByID(ctx context.Context, id string) (models.Dungeon, error) {
	var d models.Dungeon
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	if err := r.db.Collection(dungeonsCollection).FindOne(cctx, bson.M{"_id": id}).Decode(&d); err != nil {
		if err == mongo.ErrNoDocuments {
			return d, fmt.Errorf("dungeon id %s: %w", id, apperrors.ErrNotFound)
		}
		return d, fmt.Errorf("find dungeon: %w", err)
	}
	return d, nil
}

func (r *MongoRepository) ListDungeonsByFilter(ctx context.Context, filter bson.M, params models.QueryParams) ([]models.Dungeon, error) {
	q := params.Normalize()
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()

	cursor, err := r.db.Collection(dungeonsCollection).Find(cctx, filter, options.Find().SetSkip(q.Skip()).SetLimit(q.Limit).SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("list dungeons: %w", err)
	}
	defer cursor.Close(cctx)

	out := make([]models.Dungeon, 0)
	for cursor.Next(cctx) {
		var d models.Dungeon
		if err := cursor.Decode(&d); err != nil {
			return nil, fmt.Errorf("decode dungeon: %w", err)
		}
		out = append(out, d)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("dungeon cursor: %w", err)
	}
	return out, nil
}

func (r *MongoRepository) CreateStep(ctx context.Context, step models.BossStep) error {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.db.Collection(stepsCollection).InsertOne(cctx, step)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate step order: %w", apperrors.ErrConflict)
		}
		return fmt.Errorf("insert step: %w", err)
	}
	return nil
}

func (r *MongoRepository) UpdateStep(ctx context.Context, step models.BossStep) (models.BossStep, error) {
	var out models.BossStep
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()

	err := r.db.Collection(stepsCollection).FindOneAndReplace(cctx, bson.M{"_id": step.ID, "dungeonId": step.DungeonID}, step, options.FindOneAndReplace().SetReturnDocument(options.After)).Decode(&out)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return out, fmt.Errorf("step id %s: %w", step.ID, apperrors.ErrNotFound)
		}
		return out, fmt.Errorf("update step: %w", err)
	}
	return out, nil
}

func (r *MongoRepository) GetStep(ctx context.Context, dungeonID, stepID string) (models.BossStep, error) {
	var step models.BossStep
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	if err := r.db.Collection(stepsCollection).FindOne(cctx, bson.M{"_id": stepID, "dungeonId": dungeonID}).Decode(&step); err != nil {
		if err == mongo.ErrNoDocuments {
			return step, fmt.Errorf("step id %s: %w", stepID, apperrors.ErrNotFound)
		}
		return step, fmt.Errorf("find step: %w", err)
	}
	return step, nil
}

func (r *MongoRepository) ListStepsByDungeon(ctx context.Context, dungeonID string) ([]models.BossStep, error) {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()

	cursor, err := r.db.Collection(stepsCollection).Find(cctx, bson.M{"dungeonId": dungeonID}, options.Find().SetSort(bson.D{{Key: "order", Value: 1}}))
	if err != nil {
		return nil, fmt.Errorf("list steps: %w", err)
	}
	defer cursor.Close(cctx)

	steps := make([]models.BossStep, 0)
	for cursor.Next(cctx) {
		var step models.BossStep
		if err := cursor.Decode(&step); err != nil {
			return nil, fmt.Errorf("decode step: %w", err)
		}
		steps = append(steps, step)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("steps cursor: %w", err)
	}
	return steps, nil
}

func (r *MongoRepository) ReorderSteps(ctx context.Context, dungeonID string, orderByStepID map[string]int, updatedAt time.Time) error {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	collection := r.db.Collection(stepsCollection)

	for stepID, order := range orderByStepID {
		res, err := collection.UpdateOne(cctx, bson.M{"_id": stepID, "dungeonId": dungeonID}, bson.M{"$set": bson.M{"order": order, "updatedAt": updatedAt}})
		if err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return fmt.Errorf("duplicate step order: %w", apperrors.ErrConflict)
			}
			return fmt.Errorf("reorder step %s: %w", stepID, err)
		}
		if res.MatchedCount == 0 {
			return fmt.Errorf("step id %s missing: %w", stepID, apperrors.ErrNotFound)
		}
	}
	return nil
}
