package utils

import (
	"strconv"
)

const (
	ContentHeader        = "Content-type"
	Mime                 = "application/json"
	ErrMsgInvalidPayload = "Invalid request payload"
	ErrMsgServerError    = "Internal Server Error"
)

func ParseIntQuery(queryStr string, defaultValue int) int {
	// 1. Cek jika string kosong
	if queryStr == "" {
		return defaultValue
	}

	// 2. Coba konversi string ke integer
	val, err := strconv.Atoi(queryStr)
	if err != nil {
		// Jika gagal konversi (misalnya: "abc"), kembalikan default
		return defaultValue
	}

	// 3. Validasi nilai (pastikan nilainya tidak negatif atau nol)
	if val <= 0 {
		return defaultValue
	}

	return val
}
