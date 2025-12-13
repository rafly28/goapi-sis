package handlers

import (
	"encoding/json"
	"fmt" // Tambahkan fmt untuk logging dan error message
	"net/http"

	// Tambahkan strings untuk membuat array ENUM
	"go-sis-be/internal/models"
	"go-sis-be/internal/utils"
)

const (
	ADMIN_ROLE_ID   = 1
	TEACHER_ROLE_ID = 2
	STUDENT_ROLE_ID = 3
	PARENT_ROLE_ID  = 4
)

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
// 4. REGISTRASI MURID HANDLER
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
	// if req.Username == "" || req.Password == "" || req.NIK == "" || req.NISN == "" || req.FullName == "" {
	// 	respondWithError(w, http.StatusBadRequest, "Username, Password, Nama Lengkap, NIK, dan NISN wajib diisi.")
	// 	return
	// }

	if req.Username == "" || req.Password == "" || req.NIK == "" || req.NISN == "" || req.FullName == "" || req.ReceivedDate == "" {
		respondWithError(w, http.StatusBadRequest, "Username, Password, Nama Lengkap, NIK, NISN, dan Tanggal Pendaftaran wajib diisi.")
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
// 5. REGISTRASI GURU HANDLER
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

// ==========================================
// 6. REGISTRASI ADMIN
// ==========================================
func HandleAdminRegistration(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterBaseRequest

	// 1. Decode Request Body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, utils.ErrMsgInvalidPayload)
		return
	}

	// Set RoleID secara eksplisit untuk Admin
	req.RoleID = ADMIN_ROLE_ID

	// 2. Validasi Dasar (Wajib Diisi)
	if req.Username == "" || req.Password == "" || req.NIK == "" || req.FullName == "" {
		respondWithError(w, http.StatusBadRequest, "Username, Password, Nama Lengkap, dan NIK wajib diisi.")
		return
	}

	// 3. Validasi ENUMs (Sama seperti user base lainnya)
	if !isEnumValid(req.Gender, []string{models.GenderMale, models.GenderFemale}) ||
		!isEnumValid(req.Religion, []string{models.ReligionIslam, models.ReligionKristen, models.ReligionKatolik, models.ReligionHindu, models.ReligionBuddha, models.ReligionKonghucu}) ||
		!isEnumValid(req.MaritalStatus, []string{models.MaritalMarried, models.MaritalSingle, models.MaritalSingleParent}) {
		respondWithError(w, http.StatusBadRequest, "Data Person (Gender/Agama/Status Nikah) tidak valid.")
		return
	}

	// 4. Panggil Fungsi Database Transaksi (RegisterBaseUser)
	resp, err := models.RegisterBaseUser(&req)

	if err != nil {
		fmt.Printf("Error registering Admin: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Gagal mendaftarkan Admin. Data duplikat (NIK) atau error server.")
		return
	}

	// 5. Kirim Respons Sukses
	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// ==========================================
// 7. REGISTRASI ORANG TUA
// ==========================================
func HandleParentRegistration(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterBaseRequest

	// 1. Decode Request Body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, utils.ErrMsgInvalidPayload)
		return
	}

	// Set RoleID secara eksplisit untuk Wali Murid
	req.RoleID = PARENT_ROLE_ID

	// 2. Validasi Dasar (Wajib Diisi)
	if req.Username == "" || req.Password == "" || req.NIK == "" || req.FullName == "" {
		respondWithError(w, http.StatusBadRequest, "Username, Password, Nama Lengkap, dan NIK wajib diisi.")
		return
	}

	// 3. Validasi ENUMs
	if !isEnumValid(req.Gender, []string{models.GenderMale, models.GenderFemale}) ||
		!isEnumValid(req.Religion, []string{models.ReligionIslam, models.ReligionKristen, models.ReligionKatolik, models.ReligionHindu, models.ReligionBuddha, models.ReligionKonghucu}) ||
		!isEnumValid(req.MaritalStatus, []string{models.MaritalMarried, models.MaritalSingle, models.MaritalSingleParent}) {
		respondWithError(w, http.StatusBadRequest, "Data Person (Gender/Agama/Status Nikah) tidak valid.")
		return
	}

	// 4. Panggil Fungsi Database Transaksi (RegisterBaseUser)
	resp, err := models.RegisterBaseUser(&req)

	if err != nil {
		fmt.Printf("Error registering Parent: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Gagal mendaftarkan Wali Murid. Data duplikat (NIK) atau error server.")
		return
	}

	// 5. Kirim Respons Sukses
	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
