package handlers

import (
	"encoding/json"
	"go-sis-be/internal/models"
	"go-sis-be/internal/utils"
	"net/http"
)

func CreateClassHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload tidak valid", http.StatusBadRequest)
		return
	}

	// Validasi Sederhana
	if req.Name == "" || req.Level == "" || req.AcademicYear == "" {
		http.Error(w, "Nama Kelas, Tingkat, dan Tahun Ajaran wajib diisi", http.StatusBadRequest)
		return
	}

	resp, err := models.CreateClass(&req)
	if err != nil {
		http.Error(w, "Gagal membuat kelas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func GetAllClassesHandler(w http.ResponseWriter, r *http.Request) {
	// Ambil query param ?academic_year=2025/2026
	academicYear := r.URL.Query().Get("academic_year")

	classes, err := models.GetAllClasses(academicYear)
	if err != nil {
		http.Error(w, "Gagal mengambil data kelas: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(utils.ContentHeader, utils.Mime)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": classes,
	})
}