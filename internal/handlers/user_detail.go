package handlers

import (
	"encoding/json"
	"fmt"
	"go-sis-be/internal/models"
	"go-sis-be/internal/utils"
	"net/http"

	"github.com/gorilla/mux"
)

// HandleGetUserDetail menangani permintaan GET /users/{uid}
func HandleGetUserDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	if uid == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "UID user diperlukan"})
		return
	}

	finalData, err := models.GetProfileAndFormat(uid)

	if err != nil {
		if err.Error() == "profil tidak ditemukan" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Error internal server
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Gagal mengambil data profil",
			"details": err.Error(),
		})
		return
	}

	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(finalData)
}

// HandleEditProfile menangani permintaan PUT /users/{uid} untuk Guru dan Murid secara terpadu.
func HandleEditProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	// 1. Dapatkan Role ID (Menggunakan fungsi yang telah disepakati)
	roleID, err := models.GetRoleIDByUID(uid)
	if err != nil {
		if err.Error() == "UID tidak ditemukan" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Profil tidak ditemukan"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Gagal memverifikasi peran"})
		return
	}

	// 2. Switch/Case untuk memanggil logic Update yang sesuai
	var editErr error

	switch roleID {
	case models.TEACHER_ROLE_ID:
		var req models.EditTeacherRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Payload Teacher tidak valid"})
			return
		}
		editErr = models.EditTeacherProfile(uid, &req)

	case models.STUDENT_ROLE_ID:
		var req models.EditStudentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Payload Student tidak valid"})
			return
		}
		editErr = models.EditStudentProfile(uid, &req)

	default:
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Peran ini tidak diizinkan untuk diedit"})
		return
	}

	if editErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Gagal memperbarui data: %s", editErr.Error())})
		return
	}

	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profil berhasil diperbarui"})
}

// HandleDeleteProfile menangani permintaan DELETE /users/{uid} untuk Guru dan Murid secara terpadu.
func HandleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	// 1. Dapatkan Role ID (Menggunakan fungsi yang telah disepakati)
	roleID, err := models.GetRoleIDByUID(uid)
	if err != nil {
		if err.Error() == "UID tidak ditemukan" {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Profil tidak ditemukan"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Gagal memverifikasi peran"})
		return
	}

	// 2. Switch/Case untuk memanggil logic Delete yang sesuai
	var deleteErr error

	switch roleID {
	case models.TEACHER_ROLE_ID:
		deleteErr = models.DeleteTeacherProfile(uid)
	case models.STUDENT_ROLE_ID:
		deleteErr = models.DeleteStudentProfile(uid)
	default:
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Peran ini tidak diizinkan untuk dihapus"})
		return
	}

	// 3. Handle hasil Mutasi
	if deleteErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Gagal menghapus profil: %s", deleteErr.Error())})
		return
	}

	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profil berhasil dihapus"})
}
