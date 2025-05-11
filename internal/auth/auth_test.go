package auth_test

import (
	"testing"
	"time"

	"github.com/MoodyShoo/go-http-calculator/internal/auth"
	"github.com/golang-jwt/jwt"
)

func createCustomToken(store *auth.TokenStore, id int64, exp time.Time) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": id,
		"nbf":  time.Now().Unix(),
		"exp":  exp.Unix(),
		"iat":  time.Now().Unix(),
	})
	tokenStr, _ := token.SignedString(store.Config.Signature)
	return tokenStr
}

func TestTokenStore(t *testing.T) {
	store := auth.NewTokenStore()

	cases := []struct {
		name       string
		prepare    func() (string, error)
		expectErr  bool
		expectedID int64
	}{
		{
			name: "valid token",
			prepare: func() (string, error) {
				return store.AddToken(42)
			},
			expectErr:  false,
			expectedID: 42,
		},
		{
			name: "invalid token string",
			prepare: func() (string, error) {
				return "invalid.token.string", nil
			},
			expectErr:  true,
			expectedID: 0,
		},
		{
			name: "expired token",
			prepare: func() (string, error) {
				storeShort := auth.NewTokenStore()
				token := createCustomToken(storeShort, 99, time.Now().Add(-2*time.Minute))
				storeShort.AddToken(99)
				return token, nil
			},
			expectErr:  true,
			expectedID: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := tc.prepare()
			if err != nil {
				t.Fatalf("prepare() error: %v", err)
			}

			id, err := store.ValidateToken(token)
			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error but got nil, id = %d", id)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if id != tc.expectedID {
					t.Errorf("expected id = %d, got %d", tc.expectedID, id)
				}
			}
		})
	}
}
