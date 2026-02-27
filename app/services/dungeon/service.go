package dungeon

import (
	"context"
	apperrors "dungeons/app/errors"
	"dungeons/app/functions"
	"dungeons/app/models"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Repository interface {
	EnsureIndexes(ctx context.Context) error
	CreateDungeon(ctx context.Context, d models.Dungeon) error
	UpdateDungeon(ctx context.Context, d models.Dungeon) (models.Dungeon, error)
	GetDungeonByID(ctx context.Context, id string) (models.Dungeon, error)
	ListDungeonsByFilter(ctx context.Context, filter bson.M, params models.QueryParams) ([]models.Dungeon, error)
	CreateStep(ctx context.Context, step models.BossStep) error
	UpdateStep(ctx context.Context, step models.BossStep) (models.BossStep, error)
	GetStep(ctx context.Context, dungeonID, stepID string) (models.BossStep, error)
	ListStepsByDungeon(ctx context.Context, dungeonID string) ([]models.BossStep, error)
	ReorderSteps(ctx context.Context, dungeonID string, orderByStepID map[string]int, updatedAt time.Time) error
}

type Service struct {
	repo     Repository
	validate *validator.Validate
	now      func() time.Time
}

func New(repo Repository, validate *validator.Validate) *Service {
	return &Service{
		repo:     repo,
		validate: validate,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) EnsureIndexes(ctx context.Context) error {
	if err := s.repo.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("dungeon ensure indexes: %w", err)
	}
	return nil
}

func (s *Service) CreateDungeon(ctx context.Context, mjID string, req models.CreateDungeonRequest) (models.Dungeon, error) {
	if err := s.validate.Struct(req); err != nil {
		return models.Dungeon{}, fmt.Errorf("validate create dungeon: %w", apperrors.ErrValidation)
	}
	now := s.now()
	d := models.Dungeon{
		ID:          functions.NewUUID(),
		Title:       req.Title,
		Description: req.Description,
		CreatedBy:   mjID,
		AreaName:    req.AreaName,
		Status:      models.DungeonStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateDungeon(ctx, d); err != nil {
		return models.Dungeon{}, fmt.Errorf("create dungeon: %w", err)
	}
	return d, nil
}

func (s *Service) UpdateDungeon(ctx context.Context, mjID, dungeonID string, req models.UpdateDungeonRequest) (models.Dungeon, error) {
	if err := s.validate.Struct(req); err != nil {
		return models.Dungeon{}, fmt.Errorf("validate update dungeon: %w", apperrors.ErrValidation)
	}
	d, err := s.repo.GetDungeonByID(ctx, dungeonID)
	if err != nil {
		return models.Dungeon{}, fmt.Errorf("get dungeon: %w", err)
	}
	if d.CreatedBy != mjID {
		return models.Dungeon{}, fmt.Errorf("cannot update foreign dungeon: %w", apperrors.ErrForbidden)
	}
	d.Title = req.Title
	d.Description = req.Description
	d.AreaName = req.AreaName
	if req.Status != "" {
		d.Status = models.DungeonStatus(req.Status)
	}
	d.UpdatedAt = s.now()
	updated, err := s.repo.UpdateDungeon(ctx, d)
	if err != nil {
		return models.Dungeon{}, fmt.Errorf("update dungeon: %w", err)
	}
	return updated, nil
}

func (s *Service) PublishDungeon(ctx context.Context, mjID, dungeonID string) (models.Dungeon, error) {
	d, err := s.repo.GetDungeonByID(ctx, dungeonID)
	if err != nil {
		return models.Dungeon{}, fmt.Errorf("get dungeon: %w", err)
	}
	if d.CreatedBy != mjID {
		return models.Dungeon{}, fmt.Errorf("cannot publish foreign dungeon: %w", apperrors.ErrForbidden)
	}
	steps, err := s.repo.ListStepsByDungeon(ctx, dungeonID)
	if err != nil {
		return models.Dungeon{}, fmt.Errorf("list steps: %w", err)
	}
	if len(steps) == 0 {
		return models.Dungeon{}, fmt.Errorf("cannot publish empty dungeon: %w", apperrors.ErrValidation)
	}
	for _, st := range steps {
		if st.Location.RadiusMeters <= 0 {
			return models.Dungeon{}, fmt.Errorf("step %s has invalid radius: %w", st.ID, apperrors.ErrValidation)
		}
	}
	d.Status = models.DungeonStatusPublished
	d.UpdatedAt = s.now()
	updated, err := s.repo.UpdateDungeon(ctx, d)
	if err != nil {
		return models.Dungeon{}, fmt.Errorf("publish dungeon: %w", err)
	}
	return updated, nil
}

func (s *Service) ListPublished(ctx context.Context, params models.QueryParams) ([]models.Dungeon, error) {
	list, err := s.repo.ListDungeonsByFilter(ctx, bson.M{"status": models.DungeonStatusPublished}, params)
	if err != nil {
		return nil, fmt.Errorf("list published dungeons: %w", err)
	}
	return list, nil
}

func (s *Service) GetPublishedByID(ctx context.Context, id string) (models.Dungeon, []models.BossStep, error) {
	d, err := s.repo.GetDungeonByID(ctx, id)
	if err != nil {
		return models.Dungeon{}, nil, fmt.Errorf("get dungeon: %w", err)
	}
	if d.Status != models.DungeonStatusPublished {
		return models.Dungeon{}, nil, fmt.Errorf("dungeon is not published: %w", apperrors.ErrNotFound)
	}
	steps, err := s.repo.ListStepsByDungeon(ctx, id)
	if err != nil {
		return models.Dungeon{}, nil, fmt.Errorf("list steps: %w", err)
	}
	return d, steps, nil
}

func (s *Service) CreateStep(ctx context.Context, mjID, dungeonID string, req models.CreateBossStepRequest) (models.BossStep, error) {
	if err := s.validate.Struct(req); err != nil {
		return models.BossStep{}, fmt.Errorf("validate create step: %w", apperrors.ErrValidation)
	}
	if req.Location.RadiusMeters <= 0 {
		return models.BossStep{}, fmt.Errorf("radiusMeters must be positive: %w", apperrors.ErrValidation)
	}
	d, err := s.repo.GetDungeonByID(ctx, dungeonID)
	if err != nil {
		return models.BossStep{}, fmt.Errorf("get dungeon: %w", err)
	}
	if d.CreatedBy != mjID {
		return models.BossStep{}, fmt.Errorf("cannot modify foreign dungeon: %w", apperrors.ErrForbidden)
	}
	now := s.now()
	step := models.BossStep{
		ID:              functions.NewUUID(),
		DungeonID:       dungeonID,
		Order:           req.Order,
		Name:            req.Name,
		Location:        req.Location,
		ZoneDescription: req.ZoneDescription,
		Difficulty:      req.Difficulty,
		Rewards:         req.Rewards,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateStep(ctx, step); err != nil {
		return models.BossStep{}, fmt.Errorf("create step: %w", err)
	}
	return step, nil
}

func (s *Service) UpdateStep(ctx context.Context, mjID, dungeonID, stepID string, req models.UpdateBossStepRequest) (models.BossStep, error) {
	if err := s.validate.Struct(req); err != nil {
		return models.BossStep{}, fmt.Errorf("validate update step: %w", apperrors.ErrValidation)
	}
	if req.Location.RadiusMeters <= 0 {
		return models.BossStep{}, fmt.Errorf("radiusMeters must be positive: %w", apperrors.ErrValidation)
	}
	d, err := s.repo.GetDungeonByID(ctx, dungeonID)
	if err != nil {
		return models.BossStep{}, fmt.Errorf("get dungeon: %w", err)
	}
	if d.CreatedBy != mjID {
		return models.BossStep{}, fmt.Errorf("cannot modify foreign dungeon: %w", apperrors.ErrForbidden)
	}
	step, err := s.repo.GetStep(ctx, dungeonID, stepID)
	if err != nil {
		return models.BossStep{}, fmt.Errorf("get step: %w", err)
	}
	step.Name = req.Name
	step.Location = req.Location
	step.ZoneDescription = req.ZoneDescription
	step.Difficulty = req.Difficulty
	step.Rewards = req.Rewards
	step.UpdatedAt = s.now()
	updated, err := s.repo.UpdateStep(ctx, step)
	if err != nil {
		return models.BossStep{}, fmt.Errorf("update step: %w", err)
	}
	return updated, nil
}

func (s *Service) ReorderSteps(ctx context.Context, mjID, dungeonID string, req models.ReorderBossStepsRequest) ([]models.BossStep, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validate reorder steps: %w", apperrors.ErrValidation)
	}
	d, err := s.repo.GetDungeonByID(ctx, dungeonID)
	if err != nil {
		return nil, fmt.Errorf("get dungeon: %w", err)
	}
	if d.CreatedBy != mjID {
		return nil, fmt.Errorf("cannot reorder foreign dungeon: %w", apperrors.ErrForbidden)
	}
	steps, err := s.repo.ListStepsByDungeon(ctx, dungeonID)
	if err != nil {
		return nil, fmt.Errorf("list steps: %w", err)
	}
	if len(steps) != len(req.StepIDs) {
		return nil, fmt.Errorf("step count mismatch: %w", apperrors.ErrValidation)
	}
	existing := make(map[string]struct{}, len(steps))
	for _, st := range steps {
		existing[st.ID] = struct{}{}
	}
	newOrder := make(map[string]int, len(req.StepIDs))
	for idx, id := range req.StepIDs {
		if _, ok := existing[id]; !ok {
			return nil, fmt.Errorf("unknown step %s: %w", id, apperrors.ErrValidation)
		}
		newOrder[id] = idx + 1
	}
	if err := s.repo.ReorderSteps(ctx, dungeonID, newOrder, s.now()); err != nil {
		return nil, fmt.Errorf("reorder steps: %w", err)
	}
	updated, err := s.repo.ListStepsByDungeon(ctx, dungeonID)
	if err != nil {
		return nil, fmt.Errorf("list reordered steps: %w", err)
	}
	return updated, nil
}

func (s *Service) GetStepByID(ctx context.Context, dungeonID, stepID string) (models.BossStep, error) {
	step, err := s.repo.GetStep(ctx, dungeonID, stepID)
	if err != nil {
		return models.BossStep{}, fmt.Errorf("get step by id: %w", err)
	}
	return step, nil
}
