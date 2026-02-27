package inventory

import (
	"context"
	"dungeons/app/models"
	"fmt"
)

type Repository interface {
	EnsureIndexes(ctx context.Context) error
	ListInventory(ctx context.Context, playerID string) ([]models.InventoryEntry, error)
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) EnsureIndexes(ctx context.Context) error {
	if err := s.repo.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("inventory ensure indexes: %w", err)
	}
	return nil
}

func (s *Service) GetInventory(ctx context.Context, playerID string) (models.InventoryResponse, error) {
	entries, err := s.repo.ListInventory(ctx, playerID)
	if err != nil {
		return models.InventoryResponse{}, fmt.Errorf("list inventory: %w", err)
	}
	items := make([]models.InventoryItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, models.InventoryItem{ItemID: e.ItemID, Qty: e.Qty})
	}
	return models.InventoryResponse{PlayerID: playerID, Items: items}, nil
}
