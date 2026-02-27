package auction

import (
	"context"
	apperrors "dungeons/app/errors"
	"dungeons/app/functions"
	"dungeons/app/models"
	"dungeons/app/mongodb"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AuctionRepository interface {
	EnsureIndexes(ctx context.Context) error
	CreateListing(ctx context.Context, listing models.Listing) error
	ListActive(ctx context.Context, params models.QueryParams) ([]models.Listing, error)
	GetByID(ctx context.Context, id string) (models.Listing, error)
	ReplaceListing(ctx context.Context, listing models.Listing) (models.Listing, error)
	InsertTrade(ctx context.Context, trade models.Trade) error
}

type InventoryRepository interface {
	GetItemDef(ctx context.Context, itemID string) (models.ItemDef, error)
	AddItem(ctx context.Context, playerID, itemID string, qty int64, updatedAt time.Time) error
	RemoveItem(ctx context.Context, playerID, itemID string, qty int64, updatedAt time.Time) error
}

type PlayerRepository interface {
	GetByID(ctx context.Context, id string) (models.Player, error)
	IncrementGold(ctx context.Context, id string, delta int64, updatedAt time.Time) (models.Player, error)
	SetGold(ctx context.Context, id string, gold int64, updatedAt time.Time) (models.Player, error)
}

type Service struct {
	auction   AuctionRepository
	inventory InventoryRepository
	players   PlayerRepository
	validate  *validator.Validate
	client    *mongo.Client
	now       func() time.Time
}

func New(auction AuctionRepository, inventory InventoryRepository, players PlayerRepository, validate *validator.Validate, client *mongo.Client) *Service {
	return &Service{
		auction:   auction,
		inventory: inventory,
		players:   players,
		validate:  validate,
		client:    client,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) EnsureIndexes(ctx context.Context) error {
	if err := s.auction.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("auction ensure indexes: %w", err)
	}
	return nil
}

func (s *Service) CreateListing(ctx context.Context, sellerID string, req models.CreateListingRequest) (models.Listing, error) {
	if err := s.validate.Struct(req); err != nil {
		return models.Listing{}, fmt.Errorf("validate create listing: %w", apperrors.ErrValidation)
	}
	item, err := s.inventory.GetItemDef(ctx, req.ItemID)
	if err != nil {
		return models.Listing{}, fmt.Errorf("load item def: %w", err)
	}
	if !item.Tradable {
		return models.Listing{}, fmt.Errorf("item not tradable: %w", apperrors.ErrConflict)
	}

	now := s.now()
	listing := models.Listing{
		ID:           functions.NewUUID(),
		SellerID:     sellerID,
		ItemID:       req.ItemID,
		Qty:          req.Qty,
		PricePerUnit: req.PricePerUnit,
		Status:       models.ListingStatusActive,
		CreatedAt:    now,
	}
	if req.ExpiresInH > 0 {
		expires := now.Add(time.Duration(req.ExpiresInH) * time.Hour)
		listing.ExpiresAt = &expires
	}

	err = mongodb.WithTransaction(ctx, s.client, func(txCtx context.Context) error {
		if err := s.inventory.RemoveItem(txCtx, sellerID, req.ItemID, req.Qty, now); err != nil {
			return fmt.Errorf("remove seller inventory for listing: %w", err)
		}
		if err := s.auction.CreateListing(txCtx, listing); err != nil {
			return fmt.Errorf("create listing: %w", err)
		}
		return nil
	})
	if err != nil {
		return models.Listing{}, fmt.Errorf("transaction create listing: %w", err)
	}

	return listing, nil
}

func (s *Service) ListActive(ctx context.Context, params models.QueryParams) ([]models.Listing, error) {
	listings, err := s.auction.ListActive(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list active listings: %w", err)
	}
	return listings, nil
}

func (s *Service) Buy(ctx context.Context, buyerID, listingID string, req models.BuyListingRequest) (models.Listing, error) {
	if err := s.validate.Struct(req); err != nil {
		return models.Listing{}, fmt.Errorf("validate buy listing: %w", apperrors.ErrValidation)
	}
	listing, err := s.auction.GetByID(ctx, listingID)
	if err != nil {
		return models.Listing{}, fmt.Errorf("load listing: %w", err)
	}
	if listing.Status != models.ListingStatusActive {
		return models.Listing{}, fmt.Errorf("listing is not active: %w", apperrors.ErrConflict)
	}
	if listing.SellerID == buyerID {
		return models.Listing{}, fmt.Errorf("seller cannot buy own listing: %w", apperrors.ErrConflict)
	}
	if req.Qty > listing.Qty {
		return models.Listing{}, fmt.Errorf("requested qty exceeds listing qty: %w", apperrors.ErrConflict)
	}
	if listing.ExpiresAt != nil && listing.ExpiresAt.Before(s.now()) {
		return models.Listing{}, fmt.Errorf("listing expired: %w", apperrors.ErrConflict)
	}

	now := s.now()
	totalPrice := req.Qty * listing.PricePerUnit
	var out models.Listing

	err = mongodb.WithTransaction(ctx, s.client, func(txCtx context.Context) error {
		buyer, err := s.players.GetByID(txCtx, buyerID)
		if err != nil {
			return fmt.Errorf("load buyer: %w", err)
		}
		if buyer.Gold < totalPrice {
			return fmt.Errorf("insufficient funds: %w", apperrors.ErrInsufficient)
		}

		if _, err := s.players.SetGold(txCtx, buyerID, buyer.Gold-totalPrice, now); err != nil {
			return fmt.Errorf("debit buyer wallet: %w", err)
		}
		if _, err := s.players.IncrementGold(txCtx, listing.SellerID, totalPrice, now); err != nil {
			return fmt.Errorf("credit seller wallet: %w", err)
		}
		if err := s.inventory.AddItem(txCtx, buyerID, listing.ItemID, req.Qty, now); err != nil {
			return fmt.Errorf("transfer item to buyer inventory: %w", err)
		}

		if req.Qty == listing.Qty {
			listing.Status = models.ListingStatusSold
			listing.BuyerID = buyerID
			listing.Qty = 0
		} else {
			listing.Qty -= req.Qty
		}
		out, err = s.auction.ReplaceListing(txCtx, listing)
		if err != nil {
			return fmt.Errorf("update listing status after buy: %w", err)
		}

		trade := models.Trade{
			ID:         functions.NewUUID(),
			BuyerID:    buyerID,
			SellerID:   listing.SellerID,
			ListingID:  listing.ID,
			ItemID:     listing.ItemID,
			Qty:        req.Qty,
			TotalPrice: totalPrice,
			CreatedAt:  now,
		}
		if err := s.auction.InsertTrade(txCtx, trade); err != nil {
			return fmt.Errorf("insert trade: %w", err)
		}
		return nil
	})
	if err != nil {
		return models.Listing{}, fmt.Errorf("transaction buy listing: %w", err)
	}
	return out, nil
}

func (s *Service) Cancel(ctx context.Context, sellerID, listingID string) (models.Listing, error) {
	listing, err := s.auction.GetByID(ctx, listingID)
	if err != nil {
		return models.Listing{}, fmt.Errorf("load listing: %w", err)
	}
	if listing.SellerID != sellerID {
		return models.Listing{}, fmt.Errorf("cannot cancel foreign listing: %w", apperrors.ErrForbidden)
	}
	if listing.Status != models.ListingStatusActive {
		return models.Listing{}, fmt.Errorf("listing cannot be cancelled in current state: %w", apperrors.ErrConflict)
	}

	now := s.now()
	var out models.Listing
	err = mongodb.WithTransaction(ctx, s.client, func(txCtx context.Context) error {
		if listing.Qty > 0 {
			if err := s.inventory.AddItem(txCtx, sellerID, listing.ItemID, listing.Qty, now); err != nil {
				return fmt.Errorf("restore inventory on cancel: %w", err)
			}
		}
		listing.Status = models.ListingStatusCancelled
		out, err = s.auction.ReplaceListing(txCtx, listing)
		if err != nil {
			return fmt.Errorf("update listing to cancelled: %w", err)
		}
		return nil
	})
	if err != nil {
		return models.Listing{}, fmt.Errorf("transaction cancel listing: %w", err)
	}
	return out, nil
}
