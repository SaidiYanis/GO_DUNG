package player

import (
	"context"
	apperrors "dungeons/app/errors"
	"dungeons/app/models"
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
)

type playerRepoStub struct {
	createCalls int
	created     models.Player
}

func (s *playerRepoStub) EnsureIndexes(context.Context) error { return nil }
func (s *playerRepoStub) Create(_ context.Context, p models.Player) error {
	s.createCalls++
	s.created = p
	return nil
}
func (s *playerRepoStub) GetByID(context.Context, string) (models.Player, error) {
	return models.Player{}, errors.New("not implemented")
}
func (s *playerRepoStub) GetByEmail(context.Context, string) (models.Player, error) {
	return models.Player{}, errors.New("not implemented")
}
func (s *playerRepoStub) List(context.Context, models.QueryParams) ([]models.Player, error) {
	return nil, errors.New("not implemented")
}
func (s *playerRepoStub) UpdateDisplayName(context.Context, string, string, time.Time) (models.Player, error) {
	return models.Player{}, errors.New("not implemented")
}

type tokenStub struct{}

func (tokenStub) Sign(playerID, role string, ttl time.Duration) (string, error) {
	return playerID + ":" + role + ":" + ttl.String(), nil
}

func TestRegisterValidation(t *testing.T) {
	svc := New(&playerRepoStub{}, validator.New(), tokenStub{}, time.Hour)
	_, err := svc.Register(context.Background(), models.RegisterRequest{Email: "bad", DisplayName: "x", Password: "123", Role: models.RolePlayer})
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestRegisterSuccess(t *testing.T) {
	repo := &playerRepoStub{}
	svc := New(repo, validator.New(), tokenStub{}, time.Hour)

	resp, err := svc.Register(context.Background(), models.RegisterRequest{
		Email:       "ok@example.com",
		DisplayName: "PlayerOne",
		Password:    "Password123!",
		Role:        models.RolePlayer,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected create call once, got %d", repo.createCalls)
	}
	if resp.Player.Email != "ok@example.com" {
		t.Fatalf("unexpected player email: %s", resp.Player.Email)
	}
	if resp.Token == "" {
		t.Fatalf("expected non-empty token")
	}
}
