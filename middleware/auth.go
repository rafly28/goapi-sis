package middleware

import (
	"context"
	"net/http"
	"strings"

	"go-sis-be/internal/models"
	"go-sis-be/internal/utils" // Sesuaikan nama module Anda
)

type contextKey string

const UserInfoKey contextKey = "userInfo"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// 🚨 CEK BLACKLIST REDIS DISINI
		if models.IsTokenBlacklisted(tokenString) {
			http.Error(w, "Token sudah tidak berlaku (Logged Out)", http.StatusUnauthorized)
			return
		}

		// Baru setelah itu validasi JWT seperti biasa
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserInfoKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
