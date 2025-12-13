package handlers

import (
	"encoding/json" // Tambahkan fmt untuk logging dan error message
	"log"
	"net/http"
	"time"

	// Tambahkan strings untuk membuat array ENUM
	"go-sis-be/internal/models"
	"go-sis-be/internal/utils"
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

	// Cek User
	// *Asumsi* models.GetUserForLogin mengambil pass hash dan role
	user, role, err := models.GetUserForLogin(req.Username)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Periksa apakah user ditemukan dan password match
	hashStart := time.Now()
	if user == nil || !utils.CheckPasswordHash(req.Pass, user.Pass) {
		http.Error(w, "Username atau password salah", http.StatusUnauthorized)
		return
	}
	log.Printf("Hashing Time: %s", time.Since(hashStart))

	// Generate Tokens
	accessToken, _ := utils.GenerateAccessToken(user.UID, user.Username, role)
	refreshToken, _ := utils.GenerateRefreshToken(user.UID)

	// Simpan Refresh Token ke DB (PENTING!)
	if err := models.UpdateRefreshToken(user.UID, refreshToken); err != nil {
		http.Error(w, "Gagal menyimpan session", http.StatusInternalServerError)
		return
	}

	w.Header().Set(utils.ContentHeader, utils.Mime)
	json.NewEncoder(w).Encode(TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// ==========================================
// 2. REFRESH TOKEN HANDLER
// ==========================================
func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validasi Refresh Token (Signature check)
	claims, err := utils.ValidateToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "Token tidak valid atau expired", http.StatusUnauthorized)
		return
	}

	// Validasi ke Database (Apakah token ini masih aktif/belum logout?)
	storedToken, err := models.GetRefreshToken(claims.Subject) // Subject berisi UID
	if err != nil || storedToken != req.RefreshToken {
		http.Error(w, "Token sudah tidak berlaku (Logged out)", http.StatusUnauthorized)
		return
	}

	// Ambil data user terbaru (untuk role jaga-jaga kalau berubah)
	user, err := models.GetUserByID(claims.Subject)
	if err != nil {
		http.Error(w, "User tidak ditemukan", http.StatusUnauthorized)
		return
	}

	// Generate Access Token BARU
	newAccessToken, _ := utils.GenerateAccessToken(user.UID, user.Username, user.RoleName)

	w.Header().Set(utils.ContentHeader, utils.Mime)
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": newAccessToken,
	})
}

// ==========================================
// 3. LOGOUT HANDLER
// ==========================================
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	type LogoutRequest struct {
		UID string `json:"uid"`
	}
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.UID == "" {
		http.Error(w, "UID diperlukan", http.StatusBadRequest)
		return
	}

	if err := models.LogoutUser(req.UID); err != nil {
		http.Error(w, "Gagal logout", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Berhasil logout"))
}
