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
	var req RefreshRequest // Pastikan RefreshRequest sudah didefinisikan
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 1. Validasi Refresh Token (Signature check & Expiration)
	claims, err := utils.ValidateToken(req.RefreshToken)
	if err != nil {
		// Status 401: Unauthorized (token expired atau signature rusak)
		http.Error(w, "Token tidak valid atau expired", http.StatusUnauthorized)
		return
	}

	// 2. Validasi ke Database (Apakah token ini masih aktif/belum logout?)
	// Asumsi GetRefreshToken mengembalikan string token
	storedToken, err := models.GetRefreshToken(claims.Subject) // Subject berisi UID

	// PERBAIKAN: Menggunakan kondisi OR yang lebih jelas
	if err != nil || storedToken != req.RefreshToken {
		http.Error(w, "Token sudah tidak berlaku (Logged out)", http.StatusUnauthorized)
		return
	}

	// 3. Ambil data user terbaru (untuk role/username jaga-jaga kalau berubah)
	user, err := models.GetUserByID(claims.Subject)

	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "User tidak ditemukan", http.StatusUnauthorized)
			return
		}
		// Error DB lainnya
		http.Error(w, "Gagal mengambil data user", http.StatusInternalServerError)
		return
	}

	// 4. Generate Access Token BARU
	// Catatan: Jika token refresh berlaku untuk sekali pakai
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
