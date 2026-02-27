package models

import "time"

type DungeonStatus string

const (
	DungeonStatusDraft     DungeonStatus = "draft"
	DungeonStatusPublished DungeonStatus = "published"
	DungeonStatusArchived  DungeonStatus = "archived"
)

type Dungeon struct {
	ID          string        `bson:"_id" json:"id"`
	Title       string        `bson:"title" json:"title"`
	Description string        `bson:"description" json:"description"`
	CreatedBy   string        `bson:"createdBy" json:"createdBy"`
	AreaName    string        `bson:"areaName" json:"areaName"`
	Status      DungeonStatus `bson:"status" json:"status"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time     `bson:"updatedAt" json:"updatedAt"`
}

type BossLocation struct {
	Lat          float64 `bson:"lat" json:"lat"`
	Lon          float64 `bson:"lon" json:"lon"`
	RadiusMeters float64 `bson:"radiusMeters" json:"radiusMeters"`
}

type RewardItem struct {
	ItemID string `bson:"itemId" json:"itemId"`
	Qty    int64  `bson:"qty" json:"qty"`
}

type Rewards struct {
	Gold  int64        `bson:"gold" json:"gold"`
	Items []RewardItem `bson:"items" json:"items"`
}

type BossStep struct {
	ID              string       `bson:"_id" json:"id"`
	DungeonID       string       `bson:"dungeonId" json:"dungeonId"`
	Order           int          `bson:"order" json:"order"`
	Name            string       `bson:"name" json:"name"`
	Location        BossLocation `bson:"location" json:"location"`
	ZoneDescription string       `bson:"zoneDescription" json:"zoneDescription"`
	Difficulty      int          `bson:"difficulty" json:"difficulty"`
	Rewards         Rewards      `bson:"rewards" json:"rewards"`
	CreatedAt       time.Time    `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time    `bson:"updatedAt" json:"updatedAt"`
}

type CreateDungeonRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=120"`
	Description string `json:"description" validate:"required,min=3,max=1024"`
	AreaName    string `json:"areaName" validate:"required,min=2,max=120"`
}

type UpdateDungeonRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=120"`
	Description string `json:"description" validate:"required,min=3,max=1024"`
	AreaName    string `json:"areaName" validate:"required,min=2,max=120"`
	Status      string `json:"status" validate:"omitempty,oneof=draft published archived"`
}

type CreateBossStepRequest struct {
	Order           int          `json:"order" validate:"required,min=1"`
	Name            string       `json:"name" validate:"required,min=2,max=120"`
	Location        BossLocation `json:"location" validate:"required"`
	ZoneDescription string       `json:"zoneDescription" validate:"required,min=2,max=512"`
	Difficulty      int          `json:"difficulty" validate:"required,min=1,max=10"`
	Rewards         Rewards      `json:"rewards" validate:"required"`
}

type UpdateBossStepRequest struct {
	Name            string       `json:"name" validate:"required,min=2,max=120"`
	Location        BossLocation `json:"location" validate:"required"`
	ZoneDescription string       `json:"zoneDescription" validate:"required,min=2,max=512"`
	Difficulty      int          `json:"difficulty" validate:"required,min=1,max=10"`
	Rewards         Rewards      `json:"rewards" validate:"required"`
}

type ReorderBossStepsRequest struct {
	StepIDs []string `json:"stepIds" validate:"required,min=1,dive,required"`
}
