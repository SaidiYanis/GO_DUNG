package player

import (
	"dungeons/app/auth"
	"time"
)

type HMACTokenSigner struct {
	secret string
}

func NewHMACTokenSigner(secret string) *HMACTokenSigner {
	return &HMACTokenSigner{secret: secret}
}

func (s *HMACTokenSigner) Sign(playerID, role string, ttl time.Duration) (string, error) {
	return auth.Sign(s.secret, playerID, role, ttl)
}
