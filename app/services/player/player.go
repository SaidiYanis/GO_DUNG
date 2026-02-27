package player

import (
	"context"
	apperrors "dungeons/app/errors"
	"dungeons/app/functions"
	"dungeons/app/models"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	EnsureIndexes(ctx context.Context) error
	Create(ctx context.Context, p models.Player) error
	GetByID(ctx context.Context, id string) (models.Player, error)
	GetByEmail(ctx context.Context, email string) (models.Player, error)
	List(ctx context.Context, params models.QueryParams) ([]models.Player, error)
	UpdateDisplayName(ctx context.Context, id, displayName string, updatedAt time.Time) (models.Player, error)
}

type TokenSigner interface {
	Sign(playerID, role string, ttl time.Duration) (string, error)
}

type Service struct {
	repo     Repository
	validate *validator.Validate
	token    TokenSigner
	tokenTTL time.Duration
	now      func() time.Time
}

func New(repo Repository, validate *validator.Validate, token TokenSigner, tokenTTL time.Duration) *Service {
	return &Service{
		repo:     repo,
		validate: validate,
		token:    token,
		tokenTTL: tokenTTL,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) EnsureIndexes(ctx context.Context) error {
	if err := s.repo.EnsureIndexes(ctx); err != nil {
		return fmt.Errorf("player ensure indexes: %w", err)
	}
	return nil
}

func (s *Service) Register(ctx context.Context, req models.RegisterRequest) (models.AuthResponse, error) {
	var out models.AuthResponse
	if err := s.validate.Struct(req); err != nil {
		return out, fmt.Errorf("validate register request: %w", apperrors.ErrValidation)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return out, fmt.Errorf("hash password: %w", err)
	}

	now := s.now()
	player := models.Player{
		ID:           functions.NewUUID(),
		DisplayName:  req.DisplayName,
		Gold:         0,
		CreatedAt:    now,
		UpdatedAt:    now,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         req.Role,
	}

	if err := s.repo.Create(ctx, player); err != nil {
		return out, fmt.Errorf("create player: %w", err)
	}

	token, err := s.token.Sign(player.ID, string(player.Role), s.tokenTTL)
	if err != nil {
		return out, fmt.Errorf("sign token: %w", err)
	}

	out = models.AuthResponse{Token: token, Player: player.ToResponse()}
	return out, nil
}

func (s *Service) Login(ctx context.Context, req models.LoginRequest) (models.AuthResponse, error) {
	var out models.AuthResponse
	if err := s.validate.Struct(req); err != nil {
		return out, fmt.Errorf("validate login request: %w", apperrors.ErrValidation)
	}

	player, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return out, fmt.Errorf("load player by email: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(player.PasswordHash), []byte(req.Password)); err != nil {
		return out, fmt.Errorf("invalid credentials: %w", apperrors.ErrUnauthorized)
	}

	token, err := s.token.Sign(player.ID, string(player.Role), s.tokenTTL)
	if err != nil {
		return out, fmt.Errorf("sign token: %w", err)
	}

	out = models.AuthResponse{Token: token, Player: player.ToResponse()}
	return out, nil
}

func (s *Service) Me(ctx context.Context, playerID string) (models.PlayerResponse, error) {
	player, err := s.repo.GetByID(ctx, playerID)
	if err != nil {
		return models.PlayerResponse{}, fmt.Errorf("get me player: %w", err)
	}
	return player.ToResponse(), nil
}

func (s *Service) List(ctx context.Context, params models.QueryParams) ([]models.PlayerResponse, error) {
	players, err := s.repo.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list players: %w", err)
	}
	out := make([]models.PlayerResponse, 0, len(players))
	for _, p := range players {
		out = append(out, p.ToResponse())
	}
	return out, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (models.PlayerResponse, error) {
	player, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.PlayerResponse{}, fmt.Errorf("get player by id: %w", err)
	}
	return player.ToResponse(), nil
}

func (s *Service) UpdateDisplayName(ctx context.Context, id string, req models.UpdatePlayerRequest) (models.PlayerResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return models.PlayerResponse{}, fmt.Errorf("validate update player request: %w", apperrors.ErrValidation)
	}
	updated, err := s.repo.UpdateDisplayName(ctx, id, req.DisplayName, s.now())
	if err != nil {
		return models.PlayerResponse{}, fmt.Errorf("update display name: %w", err)
	}
	return updated.ToResponse(), nil
}
