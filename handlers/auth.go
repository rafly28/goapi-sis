package handlers

import (
	"encoding/json"
	"fmt" // Tambahkan fmt untuk logging dan error message
	"net/http"

	// Tambahkan strings untuk membuat array ENUM
	"go-sis-be/models"
	"go-sis-be/utils"
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
// HELPER ENUM VALIDATION
// ==========================================

// isEnumValid: Memvalidasi apakah nilai ada dalam daftar ENUM yang valid
func isEnumValid(value string, validOptions []string) bool {
	if value == "" {
		return false // Nilai wajib tidak boleh kosong
	}
	for _, option := range validOptions {
		if value == option {
			return true
		}
	}
	return false
}

// Fungsi helper untuk mengirim respons error JSON
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
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
	if user == nil || !utils.CheckPasswordHash(req.Pass, user.Pass) {
		http.Error(w, "Username atau password salah", http.StatusUnauthorized)
		return
	}

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
	// IDEALNYA: UID diambil dari Middleware Access Token
	// Untuk sementara kita ambil dari body request logout
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

// ==========================================
// 4. REGISTRASI MURID HANDLER (BARU!)
// ==========================================

func HandleStudentRegistration(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterStudentRequest

	// 1. Decode Request Body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Set RoleID secara eksplisit untuk Murid (Role ID 3)
	req.RoleID = 3

	// 2. Validasi Dasar (Wajib Diisi)
	if req.Username == "" || req.Password == "" || req.NIK == "" || req.NISN == "" || req.FullName == "" {
		respondWithError(w, http.StatusBadRequest, "Username, Password, Nama Lengkap, NIK, dan NISN wajib diisi.")
		return
	}

	// 3. Validasi ENUMs (Person & Student Details)

	// A. Person ENUMs
	if !isEnumValid(req.Gender, []string{models.GenderMale, models.GenderFemale}) {
		respondWithError(w, http.StatusBadRequest, "Gender tidak valid.")
		return
	}
	validReligions := []string{models.ReligionIslam, models.ReligionKristen, models.ReligionKatolik, models.ReligionHindu, models.ReligionBuddha, models.ReligionKonghucu}
	if !isEnumValid(req.Religion, validReligions) {
		respondWithError(w, http.StatusBadRequest, "Agama tidak valid.")
		return
	}
	validMarital := []string{models.MaritalMarried, models.MaritalSingle, models.MaritalSingleParent}
	if !isEnumValid(req.MaritalStatus, validMarital) {
		respondWithError(w, http.StatusBadRequest, "Status perkawinan tidak valid.")
		return
	}

	// B. StudentDetails ENUMs
	validFamilyStatus := []string{models.FamilyKandung, models.FamilyTiri, models.FamilyAngkat, models.FamilyLainnya}
	if !isEnumValid(req.FamilyStatus, validFamilyStatus) {
		respondWithError(w, http.StatusBadRequest, "Status dalam keluarga tidak valid.")
		return
	}
	validJobs := []string{models.JobPNS, models.JobTNI_Polri, models.JobSwasta, models.JobWiraswasta, models.JobTidakBekerja, models.JobLainnya, models.JobBUMN}

	// Pekerjaan Ayah (Tidak boleh IRT)
	if !isEnumValid(req.FatherJob, validJobs) {
		respondWithError(w, http.StatusBadRequest, "Pekerjaan Ayah tidak valid.")
		return
	}
	// Pekerjaan Ibu (Boleh IRT)
	validMotherJobs := append(validJobs, models.JobIRT)
	if !isEnumValid(req.MotherJob, validMotherJobs) {
		respondWithError(w, http.StatusBadRequest, "Pekerjaan Ibu tidak valid.")
		return
	}

	// 4. Panggil Fungsi Database Transaksi
	resp, err := models.RegisterStudent(&req)

	if err != nil {
		fmt.Printf("Error registering student: %v\n", err)
		// Pesan error umum untuk transaksi yang gagal
		respondWithError(w, http.StatusInternalServerError, "Gagal mendaftarkan Murid. Data duplikat (NIK/NISN) atau error server.")
		return
	}

	// 5. Kirim Respons Sukses
	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// ==========================================
// 5. REGISTRASI GURU HANDLER (BARU!)
// ==========================================

func HandleTeacherRegistration(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterTeacherRequest

	// 1. Decode Request Body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Set RoleID secara eksplisit untuk Guru (Role ID 2)
	req.RoleID = 2

	// 2. Validasi Dasar (Wajib Diisi)
	if req.Username == "" || req.Password == "" || req.NIK == "" || req.FullName == "" || req.University == "" {
		respondWithError(w, http.StatusBadRequest, "Data dasar, NIK, dan Universitas wajib diisi.")
		return
	}

	// 3. Validasi ENUMs (Person & Teacher Details)

	// A. Person ENUMs (Sama seperti Murid)
	if !isEnumValid(req.Gender, []string{models.GenderMale, models.GenderFemale}) ||
		!isEnumValid(req.Religion, []string{models.ReligionIslam, models.ReligionKristen, models.ReligionKatolik, models.ReligionHindu, models.ReligionBuddha, models.ReligionKonghucu}) ||
		!isEnumValid(req.MaritalStatus, []string{models.MaritalMarried, models.MaritalSingle, models.MaritalSingleParent}) {
		respondWithError(w, http.StatusBadRequest, "Data Person (Gender/Agama/Status Nikah) tidak valid.")
		return
	}

	// B. TeacherDetails ENUMs
	validEmployment := []string{models.StatusPNS, models.StatusPPPK, models.StatusKontrak, models.StatusGuruTamu, models.StatusHonorer, models.StatusLainnya}
	if !isEnumValid(req.EmploymentStatus, validEmployment) {
		respondWithError(w, http.StatusBadRequest, "Status Kepegawaian tidak valid.")
		return
	}
	validPosition := []string{models.EmploymentGuruKelas, models.EmploymentGuruMatPel, models.EmploymentKepsek, models.EmploymentLainnya}
	if !isEnumValid(req.FunctionalPosition, validPosition) {
		respondWithError(w, http.StatusBadRequest, "Jabatan Fungsional tidak valid.")
		return
	}
	validEducation := []string{models.EduSMA, models.EduD1, models.EduD2, models.EduD3, models.EduS1, models.EduS2}
	if !isEnumValid(req.LastEducation, validEducation) {
		respondWithError(w, http.StatusBadRequest, "Pendidikan Terakhir tidak valid.")
		return
	}

	// 4. Panggil Fungsi Database Transaksi
	resp, err := models.RegisterTeacher(&req)

	if err != nil {
		fmt.Printf("Error registering teacher: %v\n", err)
		// Pesan error umum untuk transaksi yang gagal
		respondWithError(w, http.StatusInternalServerError, "Gagal mendaftarkan Guru. Data duplikat (NIK) atau error server.")
		return
	}

	// 5. Kirim Respons Sukses
	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
