package models

import "time"

type ItemDef struct {
	ID          string         `bson:"_id" json:"id"`
	Type        string         `bson:"type" json:"type"`
	Rarity      string         `bson:"rarity" json:"rarity"`
	Name        string         `bson:"name" json:"name"`
	Description string         `bson:"description" json:"description"`
	Stats       map[string]any `bson:"stats" json:"stats,omitempty"`
	Tradable    bool           `bson:"tradable" json:"tradable"`
	BaseValue   int64          `bson:"baseValue" json:"baseValue"`
	CreatedAt   time.Time      `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time      `bson:"updatedAt" json:"updatedAt"`
}

type InventoryEntry struct {
	ID        string    `bson:"_id" json:"id"`
	PlayerID  string    `bson:"playerId" json:"playerId"`
	ItemID    string    `bson:"itemId" json:"itemId"`
	Qty       int64     `bson:"qty" json:"qty"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

type InventoryItem struct {
	ItemID string `json:"itemId"`
	Qty    int64  `json:"qty"`
}

type InventoryResponse struct {
	PlayerID string          `json:"playerId"`
	Items    []InventoryItem `json:"items"`
}

type ListingStatus string

const (
	ListingStatusActive    ListingStatus = "active"
	ListingStatusSold      ListingStatus = "sold"
	ListingStatusCancelled ListingStatus = "cancelled"
	ListingStatusExpired   ListingStatus = "expired"
)

type Listing struct {
	ID           string        `bson:"_id" json:"id"`
	SellerID     string        `bson:"sellerId" json:"sellerId"`
	BuyerID      string        `bson:"buyerId,omitempty" json:"buyerId,omitempty"`
	ItemID       string        `bson:"itemId" json:"itemId"`
	Qty          int64         `bson:"qty" json:"qty"`
	PricePerUnit int64         `bson:"pricePerUnit" json:"pricePerUnit"`
	Status       ListingStatus `bson:"status" json:"status"`
	CreatedAt    time.Time     `bson:"createdAt" json:"createdAt"`
	ExpiresAt    *time.Time    `bson:"expiresAt,omitempty" json:"expiresAt,omitempty"`
}

type Trade struct {
	ID         string    `bson:"_id" json:"id"`
	BuyerID    string    `bson:"buyerId" json:"buyerId"`
	SellerID   string    `bson:"sellerId" json:"sellerId"`
	ListingID  string    `bson:"listingId" json:"listingId"`
	ItemID     string    `bson:"itemId" json:"itemId"`
	Qty        int64     `bson:"qty" json:"qty"`
	TotalPrice int64     `bson:"totalPrice" json:"totalPrice"`
	CreatedAt  time.Time `bson:"createdAt" json:"createdAt"`
}

type CreateListingRequest struct {
	ItemID       string `json:"itemId" validate:"required,min=1,max=64"`
	Qty          int64  `json:"qty" validate:"required,min=1"`
	PricePerUnit int64  `json:"pricePerUnit" validate:"required,min=1"`
	ExpiresInH   int64  `json:"expiresInHours" validate:"omitempty,min=1,max=720"`
}

type BuyListingRequest struct {
	Qty int64 `json:"qty" validate:"required,min=1"`
}
