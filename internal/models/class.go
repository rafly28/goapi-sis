package models

import (
	// "database/sql"
	// "fmt"
	"go-sis-be/internal/configs"
	"time"
)

// Struct untuk Response ke Frontend
type ClassResponse struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Level               string `json:"level"`
	Major               string `json:"major"`
	AcademicYear        string `json:"academic_year"`
	HomeroomTeacherName string `json:"homeroom_teacher_name,omitempty"` // Nama Wali Kelas
	HomeroomTeacherUID  string `json:"homeroom_teacher_uid,omitempty"`
}

// Struct untuk Request Input (Create/Update)
type CreateClassRequest struct {
	Name               string `json:"name"`                // Wajib
	Level              string `json:"level"`               // Wajib
	Major              string `json:"major"`               // Opsional
	AcademicYear       string `json:"academic_year"`       // Wajib
	HomeroomTeacherUID string `json:"homeroom_teacher_uid"` // Opsional
}

// Function: Create Kelas Baru
func CreateClass(req *CreateClassRequest) (*ClassResponse, error) {
	var newID int
	var createdAt time.Time

	// Query Insert
	query := `
		INSERT INTO classes (name, level, major, academic_year, homeroom_teacher_uid)
		VALUES ($1, $2, $3, $4, NULLIF($5, '')::uuid)
		RETURNING id, created_at`
	
	// Catatan: NULLIF($5, '')::uuid berguna agar kalau string kosong dikirim, dia jadi NULL di DB
	
	err := configs.DB.QueryRow(query, req.Name, req.Level, req.Major, req.AcademicYear, req.HomeroomTeacherUID).
		Scan(&newID, &createdAt)

	if err != nil {
		return nil, err
	}

	return &ClassResponse{
		ID:           newID,
		Name:         req.Name,
		Level:        req.Level,
		Major:        req.Major,
		AcademicYear: req.AcademicYear,
	}, nil
}

// Function: Get All Classes (Bisa difilter per Tahun Ajaran)
func GetAllClasses(academicYear string) ([]ClassResponse, error) {
	query := `
		SELECT c.id, c.name, c.level, c.major, c.academic_year, 
		       COALESCE(p.full_name, '-') as teacher_name,
		       COALESCE(c.homeroom_teacher_uid::text, '') as teacher_uid
		FROM classes c
		LEFT JOIN person p ON c.homeroom_teacher_uid = p.uid
		WHERE ($1 = '' OR c.academic_year = $1)
		ORDER BY c.level ASC, c.name ASC`

	rows, err := configs.DB.Query(query, academicYear)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []ClassResponse
	for rows.Next() {
		var c ClassResponse
		// Scan harus sesuai urutan SELECT
		if err := rows.Scan(&c.ID, &c.Name, &c.Level, &c.Major, &c.AcademicYear, &c.HomeroomTeacherName, &c.HomeroomTeacherUID); err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, nil
}