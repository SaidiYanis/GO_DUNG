package models

import "time"

type Wallet struct {
	Gold int64 `bson:"gold" json:"gold"`
}

type Player struct {
	ID           string    `bson:"customID" json:"id"`
	DisplayName  string    `bson:"display_name" json:"display_name"`
	Gold         int64     `bson:"gold" json:"gold"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updated_at"`
	Email        string    `bson:"email" json:"email,omitempty"`
	PasswordHash string    `bson:"password_hash" json:"-"`
	Role         Role      `bson:"role" json:"role,omitempty"`
}

type PlayerResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        Role      `json:"role"`
	Wallet      Wallet    `json:"wallet"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (p Player) ToResponse() PlayerResponse {
	return PlayerResponse{
		ID:          p.ID,
		Email:       p.Email,
		DisplayName: p.DisplayName,
		Role:        p.Role,
		Wallet:      Wallet{Gold: p.Gold},
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
