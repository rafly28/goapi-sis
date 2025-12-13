package handlers

import (
	"database/sql"
	"encoding/json"
	"go-sis-be/internal/models"
	"go-sis-be/internal/utils"
	"go-sis-be/middleware"
	"log"
	"net/http"
	"strconv"

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

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	userInfo := r.Context().Value(middleware.UserInfoKey).(*utils.JWTClaims)
	log.Printf("Request by: %s", userInfo.Username)

	user, err := models.GetUserByID(uid)
	if err != nil {
		http.Error(w, "Gagal mengambil data user", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User tidak ditemukan", http.StatusNotFound)
		return
	}
	w.Header().Set(utils.ContentHeader, utils.Mime)
	json.NewEncoder(w).Encode(user)
}

func GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pageStr := query.Get("page")
	limitStr := query.Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if limit > 100 {
		limit = 100
	}

	resp, err := models.GetAllUsers(page, limit)
	if err != nil {
		http.Error(w, "Gagal mengambil data users", http.StatusInternalServerError)
		return
	}

	w.Header().Set(utils.ContentHeader, utils.Mime)
	json.NewEncoder(w).Encode(resp)
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
