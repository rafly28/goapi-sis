package handlers

import (
	"database/sql"
	"encoding/json"
	"go-sis-be/middleware"
	"go-sis-be/models"
	"go-sis-be/utils"
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
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	log.Printf("Data yang diterima: Username: %s, RoleID: %d\n", req.Username, req.RoleID)

	userResponse, err := models.CreateUser(&req)
	if err != nil {
		log.Printf("Error creating user: %v\n", err)
		http.Error(w, "Gagal membuat user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log sukses
	log.Printf("User berhasil dibuat: %s\n", userResponse.Username)

	// Kirim Response Sukses
	w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
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

	w.Header().Set("Content-Type", "application/json")
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
