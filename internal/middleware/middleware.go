package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/MoodyShoo/go-http-calculator/internal/auth"
	"github.com/MoodyShoo/go-http-calculator/internal/util"
)

const ID = "userID"

func GetUserID(r *http.Request) (int64, bool) {
	val := r.Context().Value(ID)
	userId, ok := val.(int64)
	return userId, ok
}

// Перед выполнением запроса проверяет авторизацию пользователя по токену
func AuthMiddleware(store *auth.TokenStore, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			util.SendError(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		userId, err := store.ValidateToken(tokenString)
		if err != nil {
			util.SendError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ID, userId)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
