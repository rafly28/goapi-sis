package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"go-sis-be/internal/utils" // Sesuaikan nama module Anda
)

type contextKey string

const UserInfoKey contextKey = "userInfo"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Ambil Header Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Token diperlukan", http.StatusUnauthorized)
			return
		}

		// 2. Format harus "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Format token salah", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// 3. Validasi Token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Token tidak valid atau expired", http.StatusUnauthorized)
			return
		}

		log.Printf("User Access: %s (Role: %s) -> %s %s",
			claims.Username,
			claims.Role,
			r.Method,
			r.URL.Path,
		)

		ctx := context.WithValue(r.Context(), UserInfoKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
