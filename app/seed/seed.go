package seed

import (
	"context"
	"dungeons/app/models"
	"dungeons/app/mongodb"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func Run(ctx context.Context, db *mongo.Database, timeout time.Duration) error {
	now := time.Now().UTC()
	cctx, cancel := mongodb.WithTimeout(ctx, timeout)
	defer cancel()

	hashMJ, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash seed mj password: %w", err)
	}
	hashPlayer, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash seed player password: %w", err)
	}

	players := []models.Player{
		{
			ID:           "seed-mj",
			DisplayName:  "Seed MJ",
			Gold:         5000,
			CreatedAt:    now,
			UpdatedAt:    now,
			Email:        "mj@seed.local",
			PasswordHash: string(hashMJ),
			Role:         models.RoleMJ,
		},
		{
			ID:           "seed-player",
			DisplayName:  "Seed Player",
			Gold:         1000,
			CreatedAt:    now,
			UpdatedAt:    now,
			Email:        "player@seed.local",
			PasswordHash: string(hashPlayer),
			Role:         models.RolePlayer,
		},
	}
	for _, p := range players {
		_, err := db.Collection("players").UpdateOne(cctx, bson.M{"customID": p.ID}, bson.M{"$set": p}, options.UpdateOne().SetUpsert(true))
		if err != nil {
			return fmt.Errorf("upsert seed player %s: %w", p.ID, err)
		}
	}

	items := []models.ItemDef{
		{
			ID:          "seed-item-sword",
			Type:        "weapon",
			Rarity:      "common",
			Name:        "Rusty Sword",
			Description: "Old but reliable",
			Stats:       map[string]any{"attack": 5},
			Tradable:    true,
			BaseValue:   25,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "seed-item-potion",
			Type:        "consumable",
			Rarity:      "common",
			Name:        "Minor Potion",
			Description: "Restores a bit of health",
			Stats:       map[string]any{"heal": 20},
			Tradable:    true,
			BaseValue:   15,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	for _, item := range items {
		_, err := db.Collection("item_defs").UpdateOne(cctx, bson.M{"_id": item.ID}, bson.M{"$set": item}, options.UpdateOne().SetUpsert(true))
		if err != nil {
			return fmt.Errorf("upsert seed item %s: %w", item.ID, err)
		}
	}

	inventories := []models.InventoryEntry{
		{ID: "seed-player:seed-item-potion", PlayerID: "seed-player", ItemID: "seed-item-potion", Qty: 5, UpdatedAt: now},
		{ID: "seed-mj:seed-item-sword", PlayerID: "seed-mj", ItemID: "seed-item-sword", Qty: 3, UpdatedAt: now},
	}
	for _, inv := range inventories {
		_, err := db.Collection("inventory").UpdateOne(cctx, bson.M{"_id": inv.ID}, bson.M{"$set": inv}, options.UpdateOne().SetUpsert(true))
		if err != nil {
			return fmt.Errorf("upsert seed inventory %s: %w", inv.ID, err)
		}
	}

	dungeon := models.Dungeon{
		ID:          "seed-dungeon-1",
		Title:       "Seed Dungeon",
		Description: "Starter published dungeon",
		CreatedBy:   "seed-mj",
		AreaName:    "Paris Center",
		Status:      models.DungeonStatusPublished,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := db.Collection("dungeons").UpdateOne(cctx, bson.M{"_id": dungeon.ID}, bson.M{"$set": dungeon}, options.UpdateOne().SetUpsert(true)); err != nil {
		return fmt.Errorf("upsert seed dungeon: %w", err)
	}

	steps := []models.BossStep{
		{
			ID:              "seed-step-1",
			DungeonID:       dungeon.ID,
			Order:           1,
			Name:            "Gatekeeper",
			Location:        models.BossLocation{Lat: 48.8566, Lon: 2.3522, RadiusMeters: 80},
			ZoneDescription: "Near city hall",
			Difficulty:      2,
			Rewards:         models.Rewards{Gold: 50, Items: []models.RewardItem{{ItemID: "seed-item-potion", Qty: 1}}},
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "seed-step-2",
			DungeonID:       dungeon.ID,
			Order:           2,
			Name:            "Catacomb Guardian",
			Location:        models.BossLocation{Lat: 48.8570, Lon: 2.3530, RadiusMeters: 120},
			ZoneDescription: "Second checkpoint",
			Difficulty:      4,
			Rewards:         models.Rewards{Gold: 120, Items: []models.RewardItem{{ItemID: "seed-item-sword", Qty: 1}}},
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}
	for _, step := range steps {
		if _, err := db.Collection("boss_steps").UpdateOne(cctx, bson.M{"_id": step.ID}, bson.M{"$set": step}, options.UpdateOne().SetUpsert(true)); err != nil {
			return fmt.Errorf("upsert seed step %s: %w", step.ID, err)
		}
	}

	listing := models.Listing{
		ID:           "seed-listing-1",
		SellerID:     "seed-mj",
		ItemID:       "seed-item-sword",
		Qty:          1,
		PricePerUnit: 200,
		Status:       models.ListingStatusActive,
		CreatedAt:    now,
	}
	if _, err := db.Collection("auction_listings").UpdateOne(cctx, bson.M{"_id": listing.ID}, bson.M{"$set": listing}, options.UpdateOne().SetUpsert(true)); err != nil {
		return fmt.Errorf("upsert seed listing: %w", err)
	}

	return nil
}
