package configs

import (
	"fmt"
	"log"

	"go-sis-be/internal/utils"
)

const (
	ADMIN_ROLE_ID = 1
	GURU_ROLE_ID  = 2
	MURID_ROLE_ID = 3
	WALI_ROLE_ID  = 4
)

func SeedDatabase() {
	// Pastikan hanya berjalan saat DEVELOPMENT (misal, cek Environment Variable)
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM login_users").Scan(&count)
	if err != nil {
		log.Fatalf("Gagal cek user count untuk seeding: %v", err)
	}
	if count > 0 {
		fmt.Println("Database sudah memiliki data user. Seeder diabaikan.")
		return
	}

	fmt.Println("Memulai Seeder Database...")

	seedRoles()
	seedInitialAdmin()

	fmt.Println("Seeder Selesai! 1 Admin dan 1000 User Dummy berhasil dibuat.")
}

func seedRoles() {
	roles := map[int]string{
		ADMIN_ROLE_ID: "admin",
		GURU_ROLE_ID:  "guru",
		MURID_ROLE_ID: "murid",
	}

	for id, name := range roles {
		query := `INSERT INTO roles (id, name, created_at, updated_at) 
			VALUES ($1, $2, NOW(), NOW()) ON CONFLICT (id) DO NOTHING`
		_, err := DB.Exec(query, id, name)
		if err != nil {
			log.Printf("Gagal seeding role %s: %v", name, err)
		}
	}
	fmt.Println("   -> Roles (Admin & User) dipastikan ada.")
}

func seedInitialAdmin() {
	password := "admin123"
	hashedPassword, _ := utils.HashPassword(password)
	username := "admin"

	// --- Mulai Transaksi untuk Admin (Wajib 2 INSERT) ---
	tx, err := DB.Begin()
	if err != nil {
		log.Fatalf("Gagal memulai transaksi Admin: %v", err)
	}
	defer tx.Rollback() // Pastikan rollback jika gagal

	var uid string

	// 1. INSERT ke login_users
	queryLogin := `INSERT INTO login_users (username, pass, role_id) 
		VALUES ($1, $2, $3) RETURNING uid`

	err = tx.QueryRow(queryLogin, username, hashedPassword, ADMIN_ROLE_ID).Scan(&uid)
	if err != nil {
		log.Fatalf("Gagal membuat Admin awal (login): %v. Pastikan roles.id=1 ada.", err)
	}

	// 2. INSERT ke person (WAJIB DIBUAT)
	queryPerson := `
		INSERT INTO person (
			uid, full_name, birth_date, nik, gender, religion, 
			marital_status, address, phone_number, email
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	// Data Person Admin (Gunakan data dummy yang valid)
	_, err = tx.Exec(queryPerson,
		uid,
		"Administrator Utama",  // full_name
		"1990-01-01",           // birth_date (Format YYYY-MM-DD)
		"1111111111111111",     // nik (16 digit)
		"Laki-laki",            // gender
		"Islam",                // religion
		"Menikah",              // marital_status
		"Kantor Pusat Sekolah", // address
		"080012345678",         // phone_number
		"admin@sis.id",         // email
	)
	if err != nil {
		log.Fatalf("Gagal membuat Admin awal (person): %v", err)
	}

	// 3. Commit Transaksi
	if err := tx.Commit(); err != nil {
		log.Fatalf("Gagal commit transaksi Admin: %v", err)
	}

	fmt.Printf(" 	-> Admin awal dibuat (Username: %s, Pass: %s). UID: %s\n", username, password, uid)
}
