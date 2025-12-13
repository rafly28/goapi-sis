package configs

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var DB *sql.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("file env tidak ditemukan")
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=3",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Gagal tersambung ke database: \n", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	err = DB.Ping()
	if err != nil {
		log.Fatal("Gagal tersambung ke database: \n", err)
	}
}

func CloseDB() {
	if DB != nil {
		log.Println("Menutup koneksi database...")
		err := DB.Close()
		if err != nil {
			// Gunakan log.Fatalf jika ini kritis dan harus menghentikan aplikasi,
			// atau log.Printf jika hanya ingin mencatat error
			log.Printf("Gagal saat menutup koneksi database: %v", err)
		} else {
			log.Println("Koneksi database berhasil ditutup.")
		}
	}
}
