package models

import (
	"fmt"
	"go-sis-be/internal/configs"
	"strings"
)

type AddStudentsRequest struct {
	ClassID     int      `json:"class_id"`
	StudentUIDs []string `json:"student_uids"` // Array UID Siswa
}

type ClassMemberResponse struct {
	ClassID      int    `json:"class_id"`
	StudentUID   string `json:"student_uid"`
	StudentName  string `json:"student_name"`
	NIS          string `json:"nis"`
	AcademicYear string `json:"academic_year"`
}

// AddStudentsToClass: Memasukkan banyak siswa sekaligus
func AddStudentsToClass(classID int, studentUIDs []string) error {
	// 1. Ambil info Tahun Ajaran/Batch dari tabel Classes
	var academicYear string
	err := configs.DB.QueryRow("SELECT academic_year FROM classes WHERE id = $1", classID).Scan(&academicYear)
	if err != nil {
		return fmt.Errorf("kelas tidak ditemukan")
	}

	if len(studentUIDs) == 0 {
		return nil
	}

	tx, err := configs.DB.Begin()
	if err != nil { return err }
	defer tx.Rollback()

	// 2. Bangun Query Bulk Insert
	query := "INSERT INTO class_members (class_id, student_uid, academic_year) VALUES "
	values := []interface{}{}
	placeholders := []string{}

	paramCount := 1
	for _, uid := range studentUIDs {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d)", paramCount, paramCount+1, paramCount+2))
		values = append(values, classID, uid, academicYear)
		paramCount += 3
	}

	fullQuery := query + strings.Join(placeholders, ",") + " ON CONFLICT DO NOTHING"

	_, err = tx.Exec(fullQuery, values...)
	if err != nil { return err }

	return tx.Commit()
}

// GetStudentsByClass: Melihat daftar murid di suatu kelas
func GetStudentsByClass(classID int) ([]ClassMemberResponse, error) {
	query := `
		SELECT cm.class_id, cm.student_uid, cm.academic_year,
		       p.full_name, COALESCE(sd.nis, '-')
		FROM class_members cm
		JOIN person p ON cm.student_uid = p.uid
		LEFT JOIN student_details sd ON cm.student_uid = sd.uid
		WHERE cm.class_id = $1
		ORDER BY p.full_name ASC`

	rows, err := configs.DB.Query(query, classID)
	if err != nil { return nil, err }
	defer rows.Close()

	var members []ClassMemberResponse
	for rows.Next() {
		var m ClassMemberResponse
		if err := rows.Scan(&m.ClassID, &m.StudentUID, &m.AcademicYear, &m.StudentName, &m.NIS); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}