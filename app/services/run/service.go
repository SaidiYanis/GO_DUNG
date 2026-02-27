package run

import (
	"context"
	apperrors "dungeons/app/errors"
	"dungeons/app/functions"
	"dungeons/app/geo"
	"dungeons/app/models"
	"dungeons/app/mongodb"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RunRepository interface {
	EnsureIndexes(ctx context.Context) error
	CreateRun(ctx context.Context, run models.Run) error
	HasActiveRun(ctx context.Context, playerID, dungeonID string) (bool, error)
	GetRunByID(ctx context.Context, id string) (models.Run, error)
	ListRunsByPlayer(ctx context.Context, playerID string, params models.QueryParams) ([]models.Run, error)
	ReplaceRun(ctx context.Context, run models.Run) (models.Run, error)
	CreateAttemptRecord(ctx context.Context, record models.AttemptRecord) error
	GetAttemptRecord(ctx context.Context, runID, stepID string) (models.AttemptRecord, error)
	UpdateAttemptRecord(ctx context.Context, id string, response any, rewardApplied bool) error
}

type DungeonRepository interface {
	GetDungeonByID(ctx context.Context, id string) (models.Dungeon, error)
	GetStep(ctx context.Context, dungeonID, stepID string) (models.BossStep, error)
	ListStepsByDungeon(ctx context.Context, dungeonID string) ([]models.BossStep, error)
}

type PlayerEconomyRepository interface {
	GetByID(ctx context.Context, id string) (models.Player, error)
	IncrementGold(ctx context.Context, id string, delta int64, updatedAt time.Time) (models.Player, error)
}

type InventoryRepository interface {
	AddItem(ctx context.Context, playerID, itemID string, qty int64, updatedAt time.Time) error
}

type Service struct {
	runs      RunRepository
	dungeons  DungeonRepository
	players   PlayerEconomyRepository
	inventory InventoryRepository
	validate  *validator.Validate
	client    *mongo.Client
	now       func() time.Time
}

func New(runs RunRepository, dungeons DungeonRepository, players PlayerEconomyRepository, inventory InventoryRepository, validate *validator.Validate, client *mongo.Client) *Service {
	return &Service{
		runs:      runs,
		dungeons:  dungeons,
		players:   players,
		inventory: inventory,
		validate:  validate,
		client:    client,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) EnsureIndexes(ctx context.Context) error {
	if err := s.runs.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("run ensure indexes: %w", err)
	}
	return nil
}

func (s *Service) Start(ctx context.Context, playerID string, req models.StartRunRequest) (models.Run, error) {
	if err := s.validate.Struct(req); err != nil {
		return models.Run{}, fmt.Errorf("validate start run request: %w", apperrors.ErrValidation)
	}
	dungeon, err := s.dungeons.GetDungeonByID(ctx, req.DungeonID)
	if err != nil {
		return models.Run{}, fmt.Errorf("get dungeon for run: %w", err)
	}
	if dungeon.Status != models.DungeonStatusPublished {
		return models.Run{}, fmt.Errorf("dungeon not published: %w", apperrors.ErrValidation)
	}
	if _, err := s.players.GetByID(ctx, playerID); err != nil {
		return models.Run{}, fmt.Errorf("get player for run: %w", err)
	}
	exists, err := s.runs.HasActiveRun(ctx, playerID, req.DungeonID)
	if err != nil {
		return models.Run{}, fmt.Errorf("check active run: %w", err)
	}
	if exists {
		return models.Run{}, fmt.Errorf("an active run already exists for this dungeon: %w", apperrors.ErrConflict)
	}
	now := s.now()
	run := models.Run{
		ID:          functions.NewUUID(),
		DungeonID:   req.DungeonID,
		PlayerID:    playerID,
		State:       models.RunStateActive,
		CurrentStep: 1,
		KilledSteps: make([]models.KilledStep, 0),
		StartedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.runs.CreateRun(ctx, run); err != nil {
		return models.Run{}, fmt.Errorf("create run: %w", err)
	}
	return run, nil
}

func (s *Service) List(ctx context.Context, playerID string, params models.QueryParams) ([]models.Run, error) {
	runs, err := s.runs.ListRunsByPlayer(ctx, playerID, params)
	if err != nil {
		return nil, fmt.Errorf("list runs: %w", err)
	}
	return runs, nil
}

func (s *Service) Get(ctx context.Context, playerID, runID string) (models.Run, error) {
	run, err := s.runs.GetRunByID(ctx, runID)
	if err != nil {
		return models.Run{}, fmt.Errorf("get run: %w", err)
	}
	if run.PlayerID != playerID {
		return models.Run{}, fmt.Errorf("run owner mismatch: %w", apperrors.ErrForbidden)
	}
	return run, nil
}

func (s *Service) Attempt(ctx context.Context, playerID, runID, stepID string, req models.AttemptRequest) (models.AttemptResponse, error) {
	var empty models.AttemptResponse
	if err := s.validate.Struct(req); err != nil {
		return empty, fmt.Errorf("validate attempt request: %w", apperrors.ErrValidation)
	}

	run, err := s.runs.GetRunByID(ctx, runID)
	if err != nil {
		return empty, fmt.Errorf("load run: %w", err)
	}
	if run.PlayerID != playerID {
		return empty, fmt.Errorf("run owner mismatch: %w", apperrors.ErrForbidden)
	}
	if run.State != models.RunStateActive {
		return empty, fmt.Errorf("run is not active: %w", apperrors.ErrConflict)
	}

	step, err := s.dungeons.GetStep(ctx, run.DungeonID, stepID)
	if err != nil {
		return empty, fmt.Errorf("load step: %w", err)
	}
	if step.Order != run.CurrentStep {
		return empty, fmt.Errorf("expected step order %d got %d: %w", run.CurrentStep, step.Order, apperrors.ErrWrongStepOrder)
	}

	distance := geo.HaversineMeters(*req.Lat, *req.Lon, step.Location.Lat, step.Location.Lon)
	if distance > step.Location.RadiusMeters {
		return empty, fmt.Errorf("distance %.2f exceeds %.2f: %w", distance, step.Location.RadiusMeters, apperrors.ErrNotInRange)
	}

	if existing, err := s.runs.GetAttemptRecord(ctx, runID, stepID); err == nil {
		if existing.IdempotencyKey != "" && existing.IdempotencyKey != req.IdempotencyKey {
			return empty, fmt.Errorf("attempt already handled with another idempotency key: %w", apperrors.ErrAlreadyHandled)
		}
		if !existing.RewardApplied {
			return empty, fmt.Errorf("attempt already in progress: %w", apperrors.ErrAlreadyHandled)
		}
		resp, convErr := decodeAttemptResponse(existing.Response)
		if convErr != nil {
			return empty, fmt.Errorf("decode cached attempt response: %w", convErr)
		}
		resp.Idempotency = true
		return resp, nil
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return empty, fmt.Errorf("check attempt replay state: %w", err)
	}

	steps, err := s.dungeons.ListStepsByDungeon(ctx, run.DungeonID)
	if err != nil {
		return empty, fmt.Errorf("list steps for completion check: %w", err)
	}
	now := s.now()
	record := models.AttemptRecord{
		ID:             functions.NewUUID(),
		RunID:          runID,
		StepID:         stepID,
		PlayerID:       playerID,
		IdempotencyKey: req.IdempotencyKey,
		RewardApplied:  false,
		CreatedAt:      now,
	}

	var response models.AttemptResponse
	txErr := mongodb.WithTransaction(ctx, s.client, func(txCtx context.Context) error {
		if err := s.runs.CreateAttemptRecord(txCtx, record); err != nil {
			return fmt.Errorf("create attempt idempotency record: %w", err)
		}

		updatedPlayer, err := s.players.IncrementGold(txCtx, playerID, step.Rewards.Gold, now)
		if err != nil {
			return fmt.Errorf("apply gold reward: %w", err)
		}
		for _, item := range step.Rewards.Items {
			if err := s.inventory.AddItem(txCtx, playerID, item.ItemID, item.Qty, now); err != nil {
				return fmt.Errorf("apply inventory reward item %s: %w", item.ItemID, err)
			}
		}

		run.KilledSteps = append(run.KilledSteps, models.KilledStep{BossStepID: stepID, KilledAt: now, AttemptID: record.ID})
		run.CurrentStep++
		if run.CurrentStep > len(steps) {
			run.State = models.RunStateCompleted
			run.EndedAt = &now
		}
		run.UpdatedAt = now

		updatedRun, err := s.runs.ReplaceRun(txCtx, run)
		if err != nil {
			return fmt.Errorf("update run progression: %w", err)
		}

		response = models.AttemptResponse{
			RunID:       runID,
			StepID:      stepID,
			DistanceM:   distance,
			Rewards:     step.Rewards,
			Run:         updatedRun,
			Player:      updatedPlayer,
			Idempotency: false,
		}

		if err := s.runs.UpdateAttemptRecord(txCtx, record.ID, response, true); err != nil {
			return fmt.Errorf("persist attempt replay response: %w", err)
		}

		return nil
	})
	if txErr != nil {
		if errors.Is(txErr, apperrors.ErrAlreadyHandled) {
			record, err := s.runs.GetAttemptRecord(ctx, runID, stepID)
			if err != nil {
				return empty, fmt.Errorf("load existing attempt after duplicate key: %w", err)
			}
			if record.IdempotencyKey != "" && record.IdempotencyKey != req.IdempotencyKey {
				return empty, fmt.Errorf("attempt already handled with another idempotency key: %w", apperrors.ErrAlreadyHandled)
			}
			resp, convErr := decodeAttemptResponse(record.Response)
			if convErr != nil {
				return empty, fmt.Errorf("decode existing attempt response: %w", convErr)
			}
			resp.Idempotency = true
			return resp, nil
		}
		return empty, fmt.Errorf("attempt transaction: %w", txErr)
	}
	return response, nil
}

func decodeAttemptResponse(raw any) (models.AttemptResponse, error) {
	var response models.AttemptResponse
	payload, err := json.Marshal(raw)
	if err != nil {
		return response, fmt.Errorf("marshal stored response: %w", err)
	}
	if err := json.Unmarshal(payload, &response); err != nil {
		return response, fmt.Errorf("unmarshal stored response: %w", err)
	}
	return response, nil
}
