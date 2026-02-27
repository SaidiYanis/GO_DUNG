package run

import (
	"context"
	apperrors "dungeons/app/errors"
	"dungeons/app/models"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
)

type runRepoStub struct {
	run     models.Run
	record  models.AttemptRecord
	hasReco bool
}

func (s *runRepoStub) EnsureIndexes(context.Context) error         { return nil }
func (s *runRepoStub) CreateRun(context.Context, models.Run) error { return nil }
func (s *runRepoStub) HasActiveRun(context.Context, string, string) (bool, error) {
	return false, nil
}
func (s *runRepoStub) GetRunByID(context.Context, string) (models.Run, error) { return s.run, nil }
func (s *runRepoStub) ListRunsByPlayer(context.Context, string, models.QueryParams) ([]models.Run, error) {
	return nil, nil
}
func (s *runRepoStub) ReplaceRun(context.Context, models.Run) (models.Run, error) {
	return models.Run{}, errors.New("not implemented")
}
func (s *runRepoStub) CreateAttemptRecord(context.Context, models.AttemptRecord) error { return nil }
func (s *runRepoStub) GetAttemptRecord(context.Context, string, string) (models.AttemptRecord, error) {
	if s.hasReco {
		return s.record, nil
	}
	return models.AttemptRecord{}, apperrors.ErrNotFound
}
func (s *runRepoStub) UpdateAttemptRecord(context.Context, string, any, bool) error { return nil }

type dungeonRepoStub struct {
	dungeon models.Dungeon
	step    models.BossStep
	steps   []models.BossStep
}

func (s *dungeonRepoStub) GetDungeonByID(context.Context, string) (models.Dungeon, error) {
	return s.dungeon, nil
}
func (s *dungeonRepoStub) GetStep(context.Context, string, string) (models.BossStep, error) {
	return s.step, nil
}
func (s *dungeonRepoStub) ListStepsByDungeon(context.Context, string) ([]models.BossStep, error) {
	return s.steps, nil
}

type playerRepoStub struct{}

func (playerRepoStub) GetByID(context.Context, string) (models.Player, error) {
	return models.Player{}, nil
}
func (playerRepoStub) IncrementGold(context.Context, string, int64, time.Time) (models.Player, error) {
	return models.Player{}, nil
}

type inventoryRepoStub struct{}

func (inventoryRepoStub) AddItem(context.Context, string, string, int64, time.Time) error { return nil }

func TestAttemptWrongStepOrder(t *testing.T) {
	lat := 48.8566
	lon := 2.3522
	runs := &runRepoStub{run: models.Run{ID: "run-1", DungeonID: "d-1", PlayerID: "p-1", State: models.RunStateActive, CurrentStep: 2}}
	dungeons := &dungeonRepoStub{step: models.BossStep{ID: "s-1", DungeonID: "d-1", Order: 1, Location: models.BossLocation{Lat: 48.8566, Lon: 2.3522, RadiusMeters: 100}}}

	svc := New(runs, dungeons, playerRepoStub{}, inventoryRepoStub{}, validator.New(), nil)
	_, err := svc.Attempt(context.Background(), "p-1", "run-1", "s-1", models.AttemptRequest{Lat: &lat, Lon: &lon, IdempotencyKey: "idem-key-123"})
	if !errors.Is(err, apperrors.ErrWrongStepOrder) {
		t.Fatalf("expected wrong step order error, got %v", err)
	}
}

func TestAttemptIdempotentReplay(t *testing.T) {
	lat := 48.8566
	lon := 2.3522
	record := models.AttemptRecord{
		Response:      map[string]any{"runId": "run-1", "stepId": "s-1", "distanceMeters": 10.0},
		RewardApplied: true,
	}
	runs := &runRepoStub{
		run:     models.Run{ID: "run-1", DungeonID: "d-1", PlayerID: "p-1", State: models.RunStateActive, CurrentStep: 1},
		record:  record,
		hasReco: true,
	}
	dungeons := &dungeonRepoStub{step: models.BossStep{ID: "s-1", DungeonID: "d-1", Order: 1, Location: models.BossLocation{Lat: 48.8566, Lon: 2.3522, RadiusMeters: 100}}}

	svc := New(runs, dungeons, playerRepoStub{}, inventoryRepoStub{}, validator.New(), nil)
	resp, err := svc.Attempt(context.Background(), "p-1", "run-1", "s-1", models.AttemptRequest{Lat: &lat, Lon: &lon, IdempotencyKey: "idem-key-123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Idempotency {
		t.Fatalf("expected idempotent replay response")
	}
	if resp.RunID != "run-1" || resp.StepID != "s-1" {
		t.Fatalf("unexpected replay payload: %#v", resp)
	}
}
