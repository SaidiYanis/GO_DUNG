package auction

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
	listingsCollection = "auction_listings"
	tradesCollection   = "auction_trades"
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
	if _, err := r.db.Collection(listingsCollection).Indexes().CreateMany(cctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "sellerId", Value: 1}}},
	}); err != nil {
		return fmt.Errorf("listing indexes: %w", err)
	}
	if _, err := r.db.Collection(tradesCollection).Indexes().CreateMany(cctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "listingId", Value: 1}}},
	}); err != nil {
		return fmt.Errorf("trade indexes: %w", err)
	}
	return nil
}

func (r *MongoRepository) CreateListing(ctx context.Context, listing models.Listing) error {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.db.Collection(listingsCollection).InsertOne(cctx, listing)
	if err != nil {
		return fmt.Errorf("insert listing: %w", err)
	}
	return nil
}

func (r *MongoRepository) ListActive(ctx context.Context, params models.QueryParams) ([]models.Listing, error) {
	q := params.Normalize()
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	now := time.Now().UTC()
	filter := bson.M{
		"status": models.ListingStatusActive,
		"$or": []bson.M{
			{"expiresAt": bson.M{"$exists": false}},
			{"expiresAt": nil},
			{"expiresAt": bson.M{"$gt": now}},
		},
	}
	cursor, err := r.db.Collection(listingsCollection).Find(cctx, filter, options.Find().SetSkip(q.Skip()).SetLimit(q.Limit).SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("list listings: %w", err)
	}
	defer cursor.Close(cctx)

	listings := make([]models.Listing, 0)
	for cursor.Next(cctx) {
		var l models.Listing
		if err := cursor.Decode(&l); err != nil {
			return nil, fmt.Errorf("decode listing: %w", err)
		}
		listings = append(listings, l)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("listing cursor: %w", err)
	}
	return listings, nil
}

func (r *MongoRepository) GetByID(ctx context.Context, id string) (models.Listing, error) {
	var listing models.Listing
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	if err := r.db.Collection(listingsCollection).FindOne(cctx, bson.M{"_id": id}).Decode(&listing); err != nil {
		if err == mongo.ErrNoDocuments {
			return listing, fmt.Errorf("listing id %s: %w", id, apperrors.ErrNotFound)
		}
		return listing, fmt.Errorf("find listing: %w", err)
	}
	return listing, nil
}

func (r *MongoRepository) ReplaceListing(ctx context.Context, listing models.Listing) (models.Listing, error) {
	var out models.Listing
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	err := r.db.Collection(listingsCollection).FindOneAndReplace(cctx, bson.M{"_id": listing.ID}, listing, options.FindOneAndReplace().SetReturnDocument(options.After)).Decode(&out)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return out, fmt.Errorf("listing id %s: %w", listing.ID, apperrors.ErrNotFound)
		}
		return out, fmt.Errorf("replace listing: %w", err)
	}
	return out, nil
}

func (r *MongoRepository) InsertTrade(ctx context.Context, trade models.Trade) error {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.db.Collection(tradesCollection).InsertOne(cctx, trade)
	if err != nil {
		return fmt.Errorf("insert trade: %w", err)
	}
	return nil
}
