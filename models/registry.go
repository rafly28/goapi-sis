// models/registry_db.go
package models

import (
	"database/sql"
	"fmt"

	"go-sis-be/configs"
	"go-sis-be/utils"
)

// RegisterStudent melakukan insert ke 3 tabel (login_users, person, student_details) dalam satu transaksi
func RegisterStudent(req *RegisterStudentRequest) (*UserProfileResponse, error) {
	// 1. Mulai Transaksi
	tx, err := configs.DB.Begin()
	if err != nil {
		return nil, err
	}
	// Defer Rollback: Jika fungsi return sebelum Commit, otomatis rollback
	defer tx.Rollback()

	// ==========================================
	// STEP 1: Insert ke login_users
	// ==========================================
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var uid string
	// Pastikan kolom 'pass' sesuai database Anda (pass atau password_hash)
	queryLogin := `
		INSERT INTO login_users (username, pass, role_id) 
		VALUES ($1, $2, $3) 
		RETURNING uid`

	err = tx.QueryRow(queryLogin, req.Username, hashedPassword, req.RoleID).Scan(&uid)
	if err != nil {
		return nil, fmt.Errorf("gagal insert login: %w", err)
	}

	// ==========================================
	// STEP 2: Insert ke person
	// ==========================================
	queryPerson := `
		INSERT INTO person (
			uid, full_name, birth_date, nik, gender, religion, 
			marital_status, address, phone_number, email
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = tx.Exec(queryPerson,
		uid, req.FullName, req.BirthDate, req.NIK, req.Gender, req.Religion,
		req.MaritalStatus, req.Address, req.PhoneNumber, req.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal insert person: %w", err)
	}

	// ==========================================
	// STEP 3: Generate NIS & Insert student_details
	// ==========================================

	// A. Generate NIS
	var nisSeq int64
	err = tx.QueryRow("SELECT nextval('nis_seq')").Scan(&nisSeq)
	if err != nil {
		return nil, fmt.Errorf("gagal generate nis sequence: %w", err)
	}
	// Format NIS (misal: 0000000001)
	generatedNIS := fmt.Sprintf("%010d", nisSeq)

	// B. Insert Detail
	queryStudent := `
		INSERT INTO student_details (
			uid, nis, nisn, family_status, child_order, 
			origin_school, received_class, received_date,
			father_name, mother_name, parent_address, father_job, mother_job,
			guardian_name, guardian_address, guardian_phone, guardian_job
		) VALUES (
			$1, $2, $3, $4, $5, 
			$6, $7, $8,
			$9, $10, $11, $12, $13,
			$14, $15, $16, $17
		)`

	// Handle Nullable Fields untuk Wali
	// Jika string kosong, kita kirim NULL ke DB menggunakan sql.NullString
	gName := sql.NullString{String: req.GuardianName, Valid: req.GuardianName != ""}
	gAddr := sql.NullString{String: req.GuardianAddress, Valid: req.GuardianAddress != ""}
	gPhone := sql.NullString{String: req.GuardianPhone, Valid: req.GuardianPhone != ""}
	gJob := sql.NullString{String: req.GuardianJob, Valid: req.GuardianJob != ""}

	_, err = tx.Exec(queryStudent,
		uid, generatedNIS, req.NISN, req.FamilyStatus, req.ChildOrder,
		req.OriginSchool, req.ReceivedClass, req.ReceivedDate,
		req.FatherName, req.MotherName, req.ParentAddress, req.FatherJob, req.MotherJob,
		gName, gAddr, gPhone, gJob,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal insert student details: %w", err)
	}

	// ==========================================
	// FINAL: Commit Transaksi
	// ==========================================
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Return response sukses
	return &UserProfileResponse{
		UID:      uid,
		Username: req.Username,
		RoleName: "Murid",
		PersonData: Person{
			UID:      uid,
			FullName: req.FullName,
			NIK:      req.NIK,
		},
	}, nil
}
func RegisterTeacher(req *RegisterTeacherRequest) (*UserProfileResponse, error) {
	// 1. Mulai Transaksi
	tx, err := configs.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Rollback jika ada error sebelum Commit

	// ==========================================
	// STEP 1: Insert ke login_users
	// ==========================================
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var uid string
	queryLogin := `
		INSERT INTO login_users (username, pass, role_id) 
		VALUES ($1, $2, $3) 
		RETURNING uid`

	err = tx.QueryRow(queryLogin, req.Username, hashedPassword, req.RoleID).Scan(&uid)
	if err != nil {
		return nil, fmt.Errorf("gagal insert login: %w", err)
	}

	// ==========================================
	// STEP 2: Insert ke person
	// ==========================================
	queryPerson := `
		INSERT INTO person (
			uid, full_name, birth_date, nik, gender, religion, 
			marital_status, address, phone_number, email
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = tx.Exec(queryPerson,
		uid, req.FullName, req.BirthDate, req.NIK, req.Gender, req.Religion,
		req.MaritalStatus, req.Address, req.PhoneNumber, req.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal insert person: %w", err)
	}

	// ==========================================
	// STEP 3: Insert teacher_details
	// ==========================================
	queryTeacher := `
		INSERT INTO teacher_details (
			uid, nip, nuptk, nrg, functional_position, employment_status, 
			rank_class, years_of_service_y, years_of_service_m, 
			sk_appointment_number, educator_cert_number, 
			last_education, university, major, graduation_year, diploma_number
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)`

	// Handle Nullable Fields
	nNip := sql.NullString{String: req.NIP, Valid: req.NIP != ""}
	nNuptk := sql.NullString{String: req.NUPTK, Valid: req.NUPTK != ""}
	nNrg := sql.NullString{String: req.NRG, Valid: req.NRG != ""}
	nRank := sql.NullString{String: req.RankClass, Valid: req.RankClass != ""}
	nDiploma := sql.NullString{String: req.DiplomaNumber, Valid: req.DiplomaNumber != ""}

	_, err = tx.Exec(queryTeacher,
		uid, nNip, nNuptk, nNrg, req.FunctionalPosition, req.EmploymentStatus,
		nRank, req.YearsOfServiceY, req.YearsOfServiceM,
		req.LastEducation, req.University, req.Major, req.GraduationYear, nDiploma,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal insert teacher details: %w", err)
	}

	// ==========================================
	// FINAL: Commit Transaksi
	// ==========================================
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Return response sukses
	return &UserProfileResponse{
		UID:      uid,
		Username: req.Username,
		RoleName: "Guru", // Hardcode karena ini fungsi register Guru
		PersonData: Person{
			UID:      uid,
			FullName: req.FullName,
			NIK:      req.NIK,
		},
	}, nil
}
