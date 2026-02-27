package run

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
	runsCollection     = "runs"
	attemptsCollection = "attempts"
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

	if _, err := r.db.Collection(runsCollection).Indexes().CreateMany(cctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "playerId", Value: 1}, {Key: "state", Value: 1}}},
		{Keys: bson.D{{Key: "dungeonId", Value: 1}}},
		{
			Keys: bson.D{{Key: "playerId", Value: 1}, {Key: "dungeonId", Value: 1}, {Key: "state", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"state": models.RunStateActive}),
		},
	}); err != nil {
		return fmt.Errorf("run indexes: %w", err)
	}

	if _, err := r.db.Collection(attemptsCollection).Indexes().CreateMany(cctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "runId", Value: 1}, {Key: "stepId", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "runId", Value: 1}, {Key: "stepId", Value: 1}, {Key: "idempotencyKey", Value: 1}}},
	}); err != nil {
		return fmt.Errorf("attempt indexes: %w", err)
	}
	return nil
}

func (r *MongoRepository) CreateRun(ctx context.Context, run models.Run) error {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.db.Collection(runsCollection).InsertOne(cctx, run)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("active run already exists: %w", apperrors.ErrConflict)
		}
		return fmt.Errorf("insert run: %w", err)
	}
	return nil
}

func (r *MongoRepository) HasActiveRun(ctx context.Context, playerID, dungeonID string) (bool, error) {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	count, err := r.db.Collection(runsCollection).CountDocuments(cctx, bson.M{
		"playerId":  playerID,
		"dungeonId": dungeonID,
		"state":     models.RunStateActive,
	})
	if err != nil {
		return false, fmt.Errorf("count active runs: %w", err)
	}
	return count > 0, nil
}

func (r *MongoRepository) GetRunByID(ctx context.Context, id string) (models.Run, error) {
	var run models.Run
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	if err := r.db.Collection(runsCollection).FindOne(cctx, bson.M{"_id": id}).Decode(&run); err != nil {
		if err == mongo.ErrNoDocuments {
			return run, fmt.Errorf("run id %s: %w", id, apperrors.ErrNotFound)
		}
		return run, fmt.Errorf("find run: %w", err)
	}
	return run, nil
}

func (r *MongoRepository) ListRunsByPlayer(ctx context.Context, playerID string, params models.QueryParams) ([]models.Run, error) {
	q := params.Normalize()
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()

	cursor, err := r.db.Collection(runsCollection).Find(cctx, bson.M{"playerId": playerID}, options.Find().SetSkip(q.Skip()).SetLimit(q.Limit).SetSort(bson.D{{Key: "startedAt", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("list runs: %w", err)
	}
	defer cursor.Close(cctx)

	runs := make([]models.Run, 0)
	for cursor.Next(cctx) {
		var run models.Run
		if err := cursor.Decode(&run); err != nil {
			return nil, fmt.Errorf("decode run: %w", err)
		}
		runs = append(runs, run)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("runs cursor: %w", err)
	}
	return runs, nil
}

func (r *MongoRepository) ReplaceRun(ctx context.Context, run models.Run) (models.Run, error) {
	var out models.Run
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	err := r.db.Collection(runsCollection).FindOneAndReplace(cctx, bson.M{"_id": run.ID}, run, options.FindOneAndReplace().SetReturnDocument(options.After)).Decode(&out)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return out, fmt.Errorf("run id %s: %w", run.ID, apperrors.ErrNotFound)
		}
		return out, fmt.Errorf("replace run: %w", err)
	}
	return out, nil
}

func (r *MongoRepository) CreateAttemptRecord(ctx context.Context, record models.AttemptRecord) error {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.db.Collection(attemptsCollection).InsertOne(cctx, record)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("duplicate attempt key: %w", apperrors.ErrAlreadyHandled)
		}
		return fmt.Errorf("insert attempt record: %w", err)
	}
	return nil
}

func (r *MongoRepository) GetAttemptRecord(ctx context.Context, runID, stepID string) (models.AttemptRecord, error) {
	var rec models.AttemptRecord
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	if err := r.db.Collection(attemptsCollection).FindOne(cctx, bson.M{"runId": runID, "stepId": stepID}).Decode(&rec); err != nil {
		if err == mongo.ErrNoDocuments {
			return rec, fmt.Errorf("attempt record missing: %w", apperrors.ErrNotFound)
		}
		return rec, fmt.Errorf("find attempt record: %w", err)
	}
	return rec, nil
}

func (r *MongoRepository) UpdateAttemptRecord(ctx context.Context, id string, response any, rewardApplied bool) error {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	res, err := r.db.Collection(attemptsCollection).UpdateOne(
		cctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"response": response, "rewardApplied": rewardApplied}},
	)
	if err != nil {
		return fmt.Errorf("update attempt record: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("attempt record %s missing: %w", id, apperrors.ErrNotFound)
	}
	return nil
}
