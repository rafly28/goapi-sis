package handlers

import (
	"encoding/json"
	"go-sis-be/internal/models"
	"go-sis-be/internal/utils" // Sesuaikan path package models Anda
	"net/http"

	"github.com/gorilla/mux"
)

func HandleGetUserDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid := vars["uid"]

	if uid == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "UID user diperlukan"})
		return
	}

	// 2. Panggil Logic Model yang Efisien
	// Fungsi ini mengurus LEFT JOIN, filter, dan formatting data berdasarkan role.
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

	// 3. Kirim Response Sukses (Data sudah bersih dan terformat)
	w.Header().Set(utils.ContentHeader, utils.Mime)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(finalData)
}
