package models

type Role string

const (
	RolePlayer Role = "player"
	RoleMJ     Role = "mj"
)

type RegisterRequest struct {
	Email       string `json:"email" validate:"required,email,max=254"`
	DisplayName string `json:"displayName" validate:"required,min=3,max=64"`
	Password    string `json:"password" validate:"required,min=8,max=128"`
	Role        Role   `json:"role" validate:"required,oneof=player mj"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

type UpdatePlayerRequest struct {
	DisplayName string `json:"displayName" validate:"required,min=3,max=64"`
}

type AuthResponse struct {
	Token  string         `json:"token"`
	Player PlayerResponse `json:"player"`
}
