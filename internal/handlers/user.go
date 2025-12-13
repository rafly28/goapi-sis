package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-sis-be/internal/models"
	"go-sis-be/internal/utils"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest

	log.Println("Menerima request POST /users")

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Error decoding request body: %v\n", err)
		http.Error(w, utils.ErrMsgInvalidPayload, http.StatusBadRequest)
		return
	}

	userResponse, err := models.CreateUser(&req)
	if err != nil {
		log.Printf("Error creating user: %v\n", err)
		http.Error(w, "Gagal membuat user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log sukses
	log.Printf("User berhasil dibuat: %s\n", userResponse.Username)

	// Kirim Response Sukses
	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userResponse)
}

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	// Default Pagination
	page := utils.ParseIntQuery(q.Get("page"), 1)
	limit := utils.ParseIntQuery(q.Get("limit"), 10)

	// Filtering Parameters (Akan dikembangkan nanti, fokus ke base query dulu)
	search := q.Get("search")
	roleID := utils.ParseIntQuery(q.Get("role_id"), 0) // role_id=2 untuk Guru, role_id=3 untuk Murid

	// 2. Hit Model Logic
	results, totalCount, err := models.GetAllUsers(page, limit, search, roleID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Gagal mengambil daftar pengguna: %s", err.Error()),
		})
		return
	}

	// 3. Buat Struktur Response Pagination
	response := map[string]interface{}{
		"total_data": totalCount,
		"page":       page,
		"limit":      limit,
		"data":       results,
	}

	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	err := models.DeleteUser(uid)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User tidak ditemukan", http.StatusNotFound)
			return
		}
		http.Error(w, "Gagal menghapus user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User berhasil dihapus"))
}

type Person struct {
	UID           string `json:"uid"`
	FullName      string `json:"full_name"`
	BirthDate     string `json:"birth_date"`
	NIK           string `json:"nik"`
	Gender        string `json:"gender"`
	Religion      string `json:"religion"`
	MaritalStatus string `json:"marital_status"`
	Address       string `json:"address"`
	PhoneNumber   string `json:"phone_number"`
	Email         string `json:"email"`
}

type RegisterBaseRequest struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	RoleID        int    // Ditetapkan di Handler (1=Admin, 4=Wali Murid)
	FullName      string `json:"full_name"`
	BirthDate     string `json:"birth_date"`
	NIK           string `json:"nik"`
	Gender        string `json:"gender"`
	Religion      string `json:"religion"`
	MaritalStatus string `json:"marital_status"`
	Address       string `json:"address"`
	PhoneNumber   string `json:"phone_number"`
	Email         string `json:"email"`
}

// RegisterTeacherRequest: Untuk Guru (Inherit BaseRequest + Teacher Details)
type RegisterTeacherRequest struct {
	RegisterBaseRequest
	FunctionalPosition string `json:"functional_position"`
	EmploymentStatus   string `json:"employment_status"`
	LastEducation      string `json:"last_education"`
	University         string `json:"university"`
}

// RegisterStudentRequest: Untuk Siswa (Inherit BaseRequest + Student Details)
type RegisterStudentRequest struct {
	RegisterBaseRequest
	NISN            string  `json:"nisn"`
	FamilyStatus    string  `json:"family_status"`
	FatherJob       string  `json:"father_job"`
	MotherJob       string  `json:"mother_job"`
	ParentAddress   string  `json:"parent_address"`
	ReceivedDate    string  `json:"received_date"`
	GuardianName    *string `json:"guardian_name,omitempty"`
	GuardianAddress *string `json:"guardian_address,omitempty"`
	GuardianPhone   *string `json:"guardian_phone,omitempty"`
	GuardianJob     *string `json:"guardian_job,omitempty"`
}
