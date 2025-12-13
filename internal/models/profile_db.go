package models

import (
	"database/sql"
	"errors"
	"fmt"
	"go-sis-be/internal/configs"
	"strconv"
)

// Pastikan import struct TeacherProfileResponse dan StudentProfileResponse dari user.go

const (
	ADMIN_ROLE_ID   = 1
	TEACHER_ROLE_ID = 2
	STUDENT_ROLE_ID = 3
	PARENT_ROLE_ID  = 4
)

// InternalUnifiedProfile: Struct besar untuk menampung hasil scan dari query LEFT JOIN (Nullable fields)
type InternalUnifiedProfile struct {
	UID           string
	RoleID        int
	Username      string
	FullName      string
	BirthDate     string
	NIK           string
	Gender        string
	Religion      string
	MaritalStatus string
	Address       string
	PhoneNumber   sql.NullString // Harus Nullable jika di DB nullable
	Email         sql.NullString // Harus Nullable jika di DB nullable

	// Teacher Fields
	NIP                 sql.NullString
	NUPTK               sql.NullString
	NRG                 sql.NullString
	FunctionalPosition  sql.NullString
	EmploymentStatus    sql.NullString
	RankClass           sql.NullString
	HireDate            sql.NullString
	SKAppointmentNumber sql.NullString
	EducatorCertNumber  sql.NullString
	LastEducation       sql.NullString
	University          sql.NullString
	Major               sql.NullString
	GraduationYear      sql.NullString
	DiplomaNumber       sql.NullString
	YosY                sql.NullInt32
	YosM                sql.NullInt32

	// Student Fields
	NISN sql.NullString
	NIS  sql.NullString

	// Field yang diambil dari student_details untuk JOIN/Parsing
	ReceivedDate sql.NullString
}

// ... (EditTeacherRequest, GetTeacherProfile, EditTeacherProfile, DeleteTeacherProfile - code lama) ...
// Saya menghilangkan fungsi CRUD lama untuk fokus pada GetProfileAndFormat

func GetProfileAndFormat(uid string) (interface{}, error) {
	// 1. Definisikan Query LEFT JOIN Besar (Menggunakan LEFT JOIN LATERAL untuk function)
	query := `
        SELECT
            lu.role_id, lu.username, 
            p.full_name, p.birth_date, p.nik, p.gender, p.religion, 
            p.marital_status, p.address, p.phone_number, p.email,
            
            -- Teacher Fields (td)
            td.nip, td.nuptk, td.nrg, td.functional_position, td.employment_status, 
            td.rank_class, td.hire_date, td.sk_appointment_number, td.educator_cert_number,
            td.last_education, td.university, td.major, td.graduation_year, td.diploma_number,
            
            -- Hasil JOIN LATERAL (yos) - SOLUSI ERROR pq: set-returning functions are not allowed in CASE
            yos.years AS yos_y, 
            yos.months AS yos_m, 
            
            -- Student Fields (sd)
            sd.nisn, sd.nis, sd.received_date

        FROM 
            login_users lu
        JOIN roles r ON lu.role_id = r.id
        JOIN person p ON lu.uid = p.uid
        LEFT JOIN teacher_details td ON lu.uid = td.uid

        -- Panggil function calculate_service hanya jika hire_date ada
        LEFT JOIN LATERAL calculate_service(td.hire_date) yos ON td.hire_date IS NOT NULL 

        LEFT JOIN student_details sd ON lu.uid = sd.uid
        WHERE lu.uid = $1`
	//

	var raw InternalUnifiedProfile

	// Variabel Nullable untuk field person (jika Nullable di DB)
	var nPhone, nEmail sql.NullString

	// 2. Eksekusi Query dan Scan Hasil
	err := configs.DB.QueryRow(query, uid).Scan(
		// Base fields (1-12)
		&raw.RoleID, &raw.Username, // <<< Ditambahkan kembali RoleName
		&raw.FullName, &raw.BirthDate, &raw.NIK, &raw.Gender, &raw.Religion,
		&raw.MaritalStatus, &raw.Address, &nPhone, &nEmail,

		// Teacher fields (td) (13-26)
		&raw.NIP, &raw.NUPTK, &raw.NRG, &raw.FunctionalPosition, &raw.EmploymentStatus,
		&raw.RankClass, &raw.HireDate, &raw.SKAppointmentNumber, &raw.EducatorCertNumber,
		&raw.LastEducation, &raw.University, &raw.Major, &raw.GraduationYear, &raw.DiplomaNumber,
		&raw.YosY, &raw.YosM, // <<< Hasil dari LEFT JOIN LATERAL

		// Student fields (sd) & Joined fields (c, m) (27-34)
		&raw.NISN, &raw.NIS, &raw.ReceivedDate,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("profil tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil profil terpadu: %w", err)
	}

	// Assign nullable person fields dan UID
	raw.PhoneNumber = nPhone
	raw.Email = nEmail
	raw.UID = uid

	// 3. Switch Case dan Mapping Output ke Struct Bersih
	switch raw.RoleID {
	case TEACHER_ROLE_ID:
		// Mapping/Transformasi ke TeacherProfileResponse
		return TeacherProfileResponse{
			// Base Mapping
			UID: raw.UID, Username: raw.Username,
			FullName: raw.FullName, BirthDate: raw.BirthDate, NIK: raw.NIK, Gender: raw.Gender,
			Religion: raw.Religion, MaritalStatus: raw.MaritalStatus, Address: raw.Address,
			PhoneNumber: raw.PhoneNumber.String, Email: raw.Email.String,

			// Teacher Mapping (menggunakan .String atau .Int32)
			NIP: raw.NIP.String, NUPTK: raw.NUPTK.String, NRG: raw.NRG.String,
			FunctionalPosition: raw.FunctionalPosition.String, EmploymentStatus: raw.EmploymentStatus.String,
			RankClass: raw.RankClass.String, HireDate: raw.HireDate.String, SKAppointmentNumber: raw.SKAppointmentNumber.String,
			EducatorCertNumber: raw.EducatorCertNumber.String, LastEducation: raw.LastEducation.String,
			University: raw.University.String, Major: raw.Major.String, GraduationYear: raw.GraduationYear.String,
			DiplomaNumber: raw.DiplomaNumber.String,

			// YOS Mapping
			YearsOfServiceY: int(raw.YosY.Int32),
			YearsOfServiceM: int(raw.YosM.Int32),
		}, nil

	case STUDENT_ROLE_ID:
		// Mapping/Transformasi ke StudentProfileResponse
		entryYear := 0
		if raw.ReceivedDate.Valid {
			// Ekstrak Tahun Masuk dari received_date (YYYY-MM-DD)
			yearStr := raw.ReceivedDate.String[:4]
			entryYear, _ = strconv.Atoi(yearStr)
		}

		return StudentProfileResponse{
			// Base Mapping
			UID: raw.UID, Username: raw.Username,
			FullName: raw.FullName, BirthDate: raw.BirthDate, NIK: raw.NIK, Gender: raw.Gender,
			Religion: raw.Religion, MaritalStatus: raw.MaritalStatus, Address: raw.Address,
			PhoneNumber: raw.PhoneNumber.String, Email: raw.Email.String,

			// Student Mapping
			NISN: raw.NISN.String, NIS: raw.NIS.String,
			EntryYear: entryYear, // Hasil Parsing
		}, nil

	default:
		// Default role (misalnya Admin) - hanya mengembalikan data base
		return struct {
			UID       string `json:"uid"`
			Username  string `json:"username"`
			RoleName  string `json:"role_name"`
			FullName  string `json:"full_name"`
			BirthDate string `json:"birth_date"`
			NIK       string `json:"nik"`
		}{
			UID: raw.UID, Username: raw.Username,
			FullName: raw.FullName, BirthDate: raw.BirthDate, NIK: raw.NIK,
		}, nil
	}
}
