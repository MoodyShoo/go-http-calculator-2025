package auth

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
)

type TokenStore struct {
	Config *Config
	tokens map[string]int64
	mu     sync.Mutex
}

func NewTokenStore() *TokenStore {
	return &TokenStore{
		Config: configFromEnv(),
		tokens: make(map[string]int64),
	}
}

func (ts *TokenStore) createToken(id int64) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": id,
		"nbf":  now.Unix(),
		"exp":  now.Add(time.Minute * 2).Unix(),
		"iat":  now.Unix(),
	})

	return token.SignedString(ts.Config.Signature)
}

func (ts *TokenStore) AddToken(id int64) (string, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	token, err := ts.createToken(id)
	if err != nil {
		return "", err
	}

	ts.tokens[token] = id

	return token, nil
}

func (ts *TokenStore) ValidateToken(tokenString string) (int64, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return ts.Config.Signature, nil
	})

	if err != nil {
		delete(ts.tokens, tokenString)
		return 0, err
	}

	if !token.Valid {
		delete(ts.tokens, tokenString)
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		delete(ts.tokens, tokenString)
		return 0, fmt.Errorf("invalid claims")
	}

	idFloat, ok := claims["name"].(float64)
	if !ok {
		delete(ts.tokens, tokenString)
		return 0, fmt.Errorf("invalid user id in token")
	}

	return int64(idFloat), nil
}
