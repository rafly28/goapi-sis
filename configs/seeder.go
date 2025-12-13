package configs

import (
	"fmt"
	"log"
	"time"

	"go-sis-be/utils"
)

const (
	ADMIN_ROLE_ID = 1
	GURU_ROLE_ID  = 2
	MURID_ROLE_ID = 3
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
	seedDummyUsers(1000)

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

	query := `INSERT INTO login_users (username, pass, role_id) 
		VALUES ($1, $2, $3)`

	_, err := DB.Exec(query, username, hashedPassword, ADMIN_ROLE_ID)
	if err != nil {
		log.Fatalf("Gagal membuat Admin awal: %v. Pastikan roles.id=1 ada.", err)
	}
	fmt.Printf("   -> Admin awal dibuat (Username: %s, Pass: %s)\n", username, password)
}

func seedDummyUsers(count int) {
	password := "password123"
	hashedPassword, _ := utils.HashPassword(password)

	tx, err := DB.Begin()
	if err != nil {
		log.Fatalf("Gagal memulai transaksi: %v", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO login_users (username, pass, role_id) 
		VALUES ($1, $2, $3)`)
	if err != nil {
		log.Fatalf("Gagal prepare statement: %v", err)
	}
	defer stmt.Close()

	start := time.Now()
	for i := 1; i <= count; i++ {
		username := fmt.Sprintf("test_user_%d", i)
		_, err := stmt.Exec(username, hashedPassword, MURID_ROLE_ID)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Gagal memasukkan user dummy: %v", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("Gagal commit transaksi: %v", err)
	}
	fmt.Printf("   -> %d User Dummy dibuat dalam %v.\n", count, time.Since(start))
}
