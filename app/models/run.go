package models

import "time"

type RunState string

const (
	RunStateActive    RunState = "active"
	RunStateCompleted RunState = "completed"
	RunStateAbandoned RunState = "abandoned"
)

type KilledStep struct {
	BossStepID string    `bson:"bossStepId" json:"bossStepId"`
	KilledAt   time.Time `bson:"killedAt" json:"killedAt"`
	AttemptID  string    `bson:"attemptId" json:"attemptId"`
}

type Run struct {
	ID          string       `bson:"_id" json:"id"`
	DungeonID   string       `bson:"dungeonId" json:"dungeonId"`
	PlayerID    string       `bson:"playerId" json:"playerId"`
	State       RunState     `bson:"state" json:"state"`
	CurrentStep int          `bson:"currentStep" json:"currentStep"`
	KilledSteps []KilledStep `bson:"killedSteps" json:"killedSteps"`
	StartedAt   time.Time    `bson:"startedAt" json:"startedAt"`
	EndedAt     *time.Time   `bson:"endedAt,omitempty" json:"endedAt,omitempty"`
	UpdatedAt   time.Time    `bson:"updatedAt" json:"updatedAt"`
}

type StartRunRequest struct {
	DungeonID string `json:"dungeonId" validate:"required,min=1,max=64"`
}

type AttemptRequest struct {
	Lat            *float64 `json:"lat" validate:"required"`
	Lon            *float64 `json:"lon" validate:"required"`
	DeviceTime     string   `json:"deviceTime" validate:"omitempty,max=64"`
	GPSAccuracyM   *float64 `json:"gpsAccuracyMeters" validate:"omitempty,gte=0"`
	IdempotencyKey string   `json:"idempotencyKey" validate:"required,min=8,max=128"`
}

type AttemptRecord struct {
	ID             string    `bson:"_id" json:"id"`
	RunID          string    `bson:"runId" json:"runId"`
	StepID         string    `bson:"stepId" json:"stepId"`
	PlayerID       string    `bson:"playerId" json:"playerId"`
	IdempotencyKey string    `bson:"idempotencyKey" json:"idempotencyKey"`
	RewardApplied  bool      `bson:"rewardApplied" json:"rewardApplied"`
	Response       any       `bson:"response" json:"response"`
	CreatedAt      time.Time `bson:"createdAt" json:"createdAt"`
}

type AttemptResponse struct {
	RunID       string      `json:"runId"`
	StepID      string      `json:"stepId"`
	DistanceM   float64     `json:"distanceMeters"`
	Rewards     Rewards     `json:"rewards"`
	Run         Run         `json:"run"`
	Player      Player      `json:"player"`
	Idempotency bool        `json:"idempotentReplay"`
	Proof       interface{} `json:"proof,omitempty"`
}
