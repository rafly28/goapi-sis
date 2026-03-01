package handlers

import (
	"encoding/json"
	"go-sis-be/internal/models"
	"go-sis-be/internal/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// POST /api/v1/classes/{id}/students
func AddStudentsToClassHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	classID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID Kelas tidak valid", http.StatusBadRequest)
		return
	}

	var req models.AddStudentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload tidak valid", http.StatusBadRequest)
		return
	}

	err = models.AddStudentsToClass(classID, req.StudentUIDs)
	if err != nil {
		http.Error(w, "Gagal menambahkan siswa: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Murid berhasil ditambahkan ke kelas",
	})
}

// GET /api/v1/classes/{id}/students
func GetClassStudentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	classID, _ := strconv.Atoi(vars["id"])

	students, err := models.GetStudentsByClass(classID)
	if err != nil {
		http.Error(w, "Gagal mengambil data murid: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(utils.ContentHeader, utils.Mime)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"class_id": classID,
		"total":    len(students),
		"data":     students,
	})
}