package models

import (
	"database/sql"
	"errors"
	"fmt"
	"go-sis-be/internal/configs"
	"strconv"
)

const (
	ADMIN_ROLE_ID   = 1
	TEACHER_ROLE_ID = 2
	STUDENT_ROLE_ID = 3
	PARENT_ROLE_ID  = 4
)

type InternalUnifiedProfile struct {
	UID           string
	RoleID        int
	Username      string
	FullName      string
	RoleName      string
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
	ReceivedDate sql.NullTime
}

func GetProfileAndFormat(uid string) (interface{}, error) {
	// 1. Definisikan Query LEFT JOIN Besar (Menggunakan LEFT JOIN LATERAL untuk function)
	query := `
        SELECT
            lu.role_id, lu.username, r.name,
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
		&raw.RoleID, &raw.Username, &raw.RoleName, // <<< Ditambahkan kembali RoleName
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
			UID: raw.UID, Username: raw.Username, RoleName: raw.RoleName,
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
		receivedDateOutput := ""
		if raw.ReceivedDate.Valid {
			receivedDateOutput = raw.ReceivedDate.Time.Format("2006-01-02")

			// 2. Ambil Tahun Masuk untuk EntryYear
			yearStr := raw.ReceivedDate.Time.Format("2006")
			entryYear, _ = strconv.Atoi(yearStr)
		}

		return StudentProfileResponse{
			// Base Mapping
			UID: raw.UID, Username: raw.Username, RoleName: raw.RoleName,
			FullName: raw.FullName, BirthDate: raw.BirthDate, NIK: raw.NIK, Gender: raw.Gender,
			Religion: raw.Religion, MaritalStatus: raw.MaritalStatus, Address: raw.Address,
			PhoneNumber: raw.PhoneNumber.String, Email: raw.Email.String,

			// Student Mapping
			NISN: raw.NISN.String, NIS: raw.NIS.String,
			ReceivedDate: receivedDateOutput,
			EntryYear:    entryYear,
		}, nil

	default:
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

func EditStudentProfile(uid string, req *EditStudentRequest) error {
	tx, err := configs.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryPerson := `
    UPDATE person SET 
        full_name = $2, birth_date = $3, 
        religion = $4, marital_status = $5, address = $6, 
        phone_number = $7, email = $8
    WHERE uid = $1`

	// Persiapan data Nullable untuk person
	nPhone := sql.NullString{String: req.PhoneNumber, Valid: req.PhoneNumber != ""}
	nEmail := sql.NullString{String: req.Email, Valid: req.Email != ""}

	resPerson, err := tx.Exec(queryPerson,
		uid, req.FullName, req.BirthDate,
		req.Religion, req.MaritalStatus, req.Address, nPhone, nEmail,
	)
	if err != nil {
		return fmt.Errorf("gagal update person (student): %w", err)
	}

	if rowsAffected, _ := resPerson.RowsAffected(); rowsAffected == 0 {
		return errors.New("data murid tidak ditemukan (person)")
	}

	// 2. UPDATE student_details
	queryStudent := `
    UPDATE student_details SET 
        nisn = $2, nis = $3, received_date = $4
    WHERE uid = $1`

	nReceivedDate := sql.NullString{String: req.ReceivedDate, Valid: req.ReceivedDate != ""}

	_, err = tx.Exec(queryStudent,
		uid, req.NISN, req.NIS, nReceivedDate,
	)
	if err != nil {
		return fmt.Errorf("gagal update data murid: %w", err)
	}

	// 3. Commit Transaction (Memastikan kedua update berhasil)
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func DeleteStudentProfile(uid string) error {
	tx, err := configs.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM student_details WHERE uid = $1", uid)
	if err != nil {
		return fmt.Errorf("gagal delete data murid: %w", err)
	}

	_, err = tx.Exec("DELETE FROM person WHERE uid = $1", uid)
	if err != nil {
		return fmt.Errorf("gagal delete person (student): %w", err)
	}

	res, err := tx.Exec("DELETE FROM login_users WHERE uid = $1", uid)
	if err != nil {
		return fmt.Errorf("gagal delete login user (student): %w", err)
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return errors.New("data murid tidak ditemukan")
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func EditTeacherProfile(uid string, req *EditTeacherRequest) error {
	tx, err := configs.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. UPDATE person
	// Menggunakan 10 parameter ($1-$10)
	queryPerson := `
    UPDATE person SET 
        full_name = $2,
        religion = $3, marital_status = $4, address = $5, 
        phone_number = $6, email = $7
    WHERE uid = $1`

	// Persiapan data Nullable untuk person (PhoneNumber, Email)
	nPhone := sql.NullString{String: req.PhoneNumber, Valid: req.PhoneNumber != ""}
	nEmail := sql.NullString{String: req.Email, Valid: req.Email != ""}

	resPerson, err := tx.Exec(queryPerson,
		uid, req.FullName,
		req.Religion, req.MaritalStatus, req.Address, nPhone, nEmail,
	)
	if err != nil {
		return fmt.Errorf("gagal update person (teacher): %w", err)
	}

	if rowsAffected, _ := resPerson.RowsAffected(); rowsAffected == 0 {
		return errors.New("data guru tidak ditemukan (person)")
	}

	// 2. UPDATE teacher_details
	// Menggunakan 15 parameter ($1-$15)
	queryTeacher := `
    UPDATE teacher_details SET 
        nip = $2, nuptk = $3, nrg = $4, functional_position = $5, 
        employment_status = $6, rank_class = $7, hire_date = $8,
        sk_appointment_number = $9, educator_cert_number = $10, 
        last_education = $11, university = $12, major = $13, 
        graduation_year = $14, diploma_number = $15
    WHERE uid = $1`

	// Persiapan data Nullable untuk teacher_details (NUPTK, NRG, RankClass, SKAppointmentNumber, EducatorCertNumber, DiplomaNumber)
	nNuptk := sql.NullString{String: req.NUPTK, Valid: req.NUPTK != ""}
	nNrg := sql.NullString{String: req.NRG, Valid: req.NRG != ""}
	nRank := sql.NullString{String: req.RankClass, Valid: req.RankClass != ""}
	nSK := sql.NullString{String: req.SKAppointmentNumber, Valid: req.SKAppointmentNumber != ""}
	nCert := sql.NullString{String: req.EducatorCertNumber, Valid: req.EducatorCertNumber != ""}
	nDiploma := sql.NullString{String: req.DiplomaNumber, Valid: req.DiplomaNumber != ""}

	_, err = tx.Exec(queryTeacher,
		uid, req.NIP, nNuptk, nNrg, req.FunctionalPosition, req.EmploymentStatus,
		nRank, req.HireDate, nSK, nCert,
		req.LastEducation, req.University, req.Major, req.GraduationYear, nDiploma,
	)
	if err != nil {
		return fmt.Errorf("gagal update data guru: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func DeleteTeacherProfile(uid string) error {
	tx, err := configs.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM teacher_details WHERE uid = $1", uid)
	if err != nil {
		return fmt.Errorf("gagal delete data guru: %w", err)
	}

	_, err = tx.Exec("DELETE FROM person WHERE uid = $1", uid)
	if err != nil {
		return fmt.Errorf("gagal delete person (teacher): %w", err)
	}

	res, err := tx.Exec("DELETE FROM login_users WHERE uid = $1", uid)
	if err != nil {
		return fmt.Errorf("gagal delete login user (teacher): %w", err)
	}

	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		return errors.New("data guru tidak ditemukan")
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
