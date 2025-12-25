package handlers

import (
	"encoding/json" // Tambahkan fmt untuk logging dan error message
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	// Tambahkan strings untuk membuat array ENUM
	"go-sis-be/internal/models"
	"go-sis-be/internal/utils"
	"go-sis-be/middleware"
)

// ==========================================
// STRUCT REQUEST/RESPONSE OTENTIKASI
// ==========================================

type LoginRequest struct {
	Username string `json:"username"`
	Pass     string `json:"password"` // Sebenarnya pass (DB) tapi tag JSON-nya password (Client)
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ==========================================
// 1. LOGIN HANDLER
// ==========================================
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, role, err := models.GetUserForLogin(req.Username)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	hashStart := time.Now()
	if user == nil || !utils.CheckPasswordHash(req.Pass, user.Pass) {
		http.Error(w, "Username atau password salah", http.StatusUnauthorized)
		return
	}
	log.Printf("Checking Credential: %s", time.Since(hashStart))

	// Generate Tokens
	accessToken, _ := utils.GenerateAccessToken(user.UID, user.Username, role)
	refreshToken, _ := utils.GenerateRefreshToken(user.UID)

	// Simpan Refresh Token ke DB (PENTING!)
	if err := models.UpdateRefreshToken(user.UID, refreshToken); err != nil {
		http.Error(w, "Gagal menyimpan session", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set(utils.ContentHeader, utils.Mime)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Login Success!",
		"access_token": accessToken,
	})
}

// ==========================================
// 2. REFRESH TOKEN HANDLER
// ==========================================
func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		fmt.Println("KOK KOSONG? Errornya:", err)
		http.Error(w, "Cookie gak ada Aa", http.StatusUnauthorized)
		return
	}
	fmt.Println("Kuncinya dapet nih:", cookie.Value)
	refreshTokenString := cookie.Value

	claims, err := utils.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		http.Error(w, "Token tidak valid: "+err.Error(), http.StatusUnauthorized)
		return
	}

	userSess, err := models.GetUserSessionByUID(claims.UID)
	if err != nil || userSess.RefreshToken != refreshTokenString {
		http.Error(w, "Sesi kadaluarsa atau sudah logout", http.StatusUnauthorized)
		return
	}

	newAccessToken, _ := utils.GenerateAccessToken(userSess.UID, userSess.Username, userSess.Role)

	w.Header().Set(utils.ContentHeader, utils.Mime)
	json.NewEncoder(w).Encode(map[string]string{
		"message":      "Refresh Success!",
		"access_token": newAccessToken,
	})
}

// ==========================================
// 3. LOGOUT HANDLER
// ==========================================
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	claimsContext := r.Context().Value(middleware.UserInfoKey)
	claims, ok := claimsContext.(*utils.JWTClaims)

	if !ok || claims == nil || claims.UID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	uid := claims.UID

	err := models.DeleteRefreshToken(uid)
	if err != nil {
		log.Printf("Error deleting refresh token: %v", err)
	}

	authHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	errBlacklist := models.BlacklistToken(tokenString, 15*time.Minute)
	if errBlacklist != nil {
		log.Printf("ERROR REDIS BLACKLIST: %v", errBlacklist)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout Success!"})
}
