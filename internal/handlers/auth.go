package handlers

import (
	"encoding/json" // Tambahkan fmt untuk logging dan error message
	"log"
	"net/http"
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
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 1. Validasi Refresh Token (Signature check & Expiration)
	claims, err := utils.ValidateToken(req.RefreshToken) // Mengambil claims dari token
	if err != nil {
		http.Error(w, "Token tidak valid atau expired", http.StatusUnauthorized)
		return
	}

	if claims.Subject == "" {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	// 2. Validasi ke Database (Revocation Check)
	storedToken, err := models.GetRefreshToken(claims.Subject) // Subject berisi UID

	if err != nil {
		// Error DB internal
		http.Error(w, "Gagal memverifikasi token", http.StatusInternalServerError)
		return
	}

	if storedToken != req.RefreshToken || storedToken == "" {
		// Token mismatch (sudah di-refresh atau di-revoke) atau token tidak ditemukan di DB
		http.Error(w, "Token sudah tidak berlaku (Revoked or Logged out)", http.StatusUnauthorized)
		return
	}

	// 3. Ambil data user terbaru (Menggunakan GetUserByID yang sudah dikoreksi)
	user, err := models.GetUserByID(claims.Subject)
	if err != nil {
		// User tidak ditemukan (kemungkinan user sudah dihapus)
		http.Error(w, "User tidak ditemukan", http.StatusUnauthorized)
		return
	}

	// 4. Generate Access Token BARU
	newAccessToken, err := utils.GenerateAccessToken(user.UID, user.Username, user.RoleName)
	if err != nil {
		http.Error(w, "Gagal generate access token baru", http.StatusInternalServerError)
		return
	}

	// 5. Kirim Response
	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": newAccessToken,
	})
}

// ==========================================
// 3. LOGOUT HANDLER
// ==========================================
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	claimsContext := r.Context().Value(middleware.UserInfoKey)
	claims, ok := claimsContext.(*utils.JWTClaims)

	// PERBAIKAN DI SINI: Ganti claims.Subject menjadi claims.UID
	if !ok || claims == nil || claims.UID == "" {
		http.Error(w, "Unauthorized: Claims not found in context or UID missing", http.StatusUnauthorized)
		return
	}

	// PERBAIKAN DI SINI: Ambil UID dari claims.UID
	uid := claims.UID

	err := models.DeleteRefreshToken(uid)
	if err != nil {
		log.Printf("Error deleting refresh token for UID %s: %v", uid, err)
	}
	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout berhasil"})
}
