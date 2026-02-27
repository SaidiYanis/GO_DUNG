package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Claims struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
}

func Sign(secret, playerID, role string, ttl time.Duration) (string, error) {
	claims := Claims{Sub: playerID, Role: role, Exp: time.Now().Add(ttl).Unix()}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadRaw := base64.RawURLEncoding.EncodeToString(payload)
	sig := signBytes([]byte(payloadRaw), []byte(secret))
	return payloadRaw + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func Parse(secret, token string) (Claims, error) {
	var claims Claims
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return claims, fmt.Errorf("invalid token format")
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return claims, fmt.Errorf("decode signature: %w", err)
	}

	expected := signBytes([]byte(parts[0]), []byte(secret))
	if !hmac.Equal(sig, expected) {
		return claims, fmt.Errorf("invalid signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return claims, fmt.Errorf("decode payload: %w", err)
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return claims, fmt.Errorf("unmarshal claims: %w", err)
	}

	if time.Now().Unix() > claims.Exp {
		return claims, fmt.Errorf("token expired")
	}

	return claims, nil
}

func signBytes(payload, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(payload)
	return h.Sum(nil)
}
