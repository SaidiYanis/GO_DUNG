package inventory

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
	inventoryCollection = "inventory"
	itemsCollection     = "item_defs"
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
	_, err := r.db.Collection(inventoryCollection).Indexes().CreateMany(cctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "playerId", Value: 1}, {Key: "itemId", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "playerId", Value: 1}}},
	})
	if err != nil {
		return fmt.Errorf("inventory indexes: %w", err)
	}
	return nil
}

func (r *MongoRepository) ListInventory(ctx context.Context, playerID string) ([]models.InventoryEntry, error) {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	cursor, err := r.db.Collection(inventoryCollection).Find(cctx, bson.M{"playerId": playerID, "qty": bson.M{"$gt": 0}})
	if err != nil {
		return nil, fmt.Errorf("list inventory: %w", err)
	}
	defer cursor.Close(cctx)

	items := make([]models.InventoryEntry, 0)
	for cursor.Next(cctx) {
		var entry models.InventoryEntry
		if err := cursor.Decode(&entry); err != nil {
			return nil, fmt.Errorf("decode inventory: %w", err)
		}
		items = append(items, entry)
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("inventory cursor: %w", err)
	}
	return items, nil
}

func (r *MongoRepository) AddItem(ctx context.Context, playerID, itemID string, qty int64, updatedAt time.Time) error {
	if qty <= 0 {
		return fmt.Errorf("qty must be positive: %w", apperrors.ErrValidation)
	}
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.db.Collection(inventoryCollection).UpdateOne(cctx,
		bson.M{"playerId": playerID, "itemId": itemID},
		bson.M{"$inc": bson.M{"qty": qty}, "$set": bson.M{"updatedAt": updatedAt}, "$setOnInsert": bson.M{"_id": playerID + ":" + itemID, "playerId": playerID, "itemId": itemID}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return fmt.Errorf("add inventory item: %w", err)
	}
	return nil
}

func (r *MongoRepository) RemoveItem(ctx context.Context, playerID, itemID string, qty int64, updatedAt time.Time) error {
	if qty <= 0 {
		return fmt.Errorf("qty must be positive: %w", apperrors.ErrValidation)
	}
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	res, err := r.db.Collection(inventoryCollection).UpdateOne(cctx,
		bson.M{"playerId": playerID, "itemId": itemID, "qty": bson.M{"$gte": qty}},
		bson.M{"$inc": bson.M{"qty": -qty}, "$set": bson.M{"updatedAt": updatedAt}},
	)
	if err != nil {
		return fmt.Errorf("remove inventory item: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("not enough item quantity: %w", apperrors.ErrConflict)
	}
	return nil
}

func (r *MongoRepository) GetItemDef(ctx context.Context, itemID string) (models.ItemDef, error) {
	var item models.ItemDef
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	if err := r.db.Collection(itemsCollection).FindOne(cctx, bson.M{"_id": itemID}).Decode(&item); err != nil {
		if err == mongo.ErrNoDocuments {
			return item, fmt.Errorf("item %s: %w", itemID, apperrors.ErrNotFound)
		}
		return item, fmt.Errorf("find item def: %w", err)
	}
	return item, nil
}

func (r *MongoRepository) UpsertItemDef(ctx context.Context, item models.ItemDef) error {
	cctx, cancel := mongodb.WithTimeout(ctx, r.timeout)
	defer cancel()
	_, err := r.db.Collection(itemsCollection).UpdateOne(cctx, bson.M{"_id": item.ID}, bson.M{"$set": item}, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("upsert item def: %w", err)
	}
	return nil
}
