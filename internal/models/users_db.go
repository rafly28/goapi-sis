// models/users_db.go
package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-sis-be/internal/configs"
	"go-sis-be/internal/utils"
)

// --- BAGIAN AUTH (Login, Refresh, Logout) ---
func GetUserForLogin(username string) (*User, string, error) {
	var user User
	var roleName string

	query := `
		SELECT u.uid, u.username, u.pass, u.role_id, r.name 
		FROM login_users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1`

	row := configs.DB.QueryRow(query, username)
	err := row.Scan(&user.UID, &user.Username, &user.Pass, &user.RoleID, &roleName)

	if err == sql.ErrNoRows {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}

	return &user, roleName, nil
}

func UpdateRefreshToken(uid string, token string) error {
	query := "UPDATE login_users SET refresh_token = $1, updated_at = NOW() WHERE uid = $2"
	_, err := configs.DB.Exec(query, token, uid)
	return err
}

func GetRefreshToken(uid string) (string, error) {
	var token sql.NullString
	query := "SELECT refresh_token FROM login_users WHERE uid = $1"

	err := configs.DB.QueryRow(query, uid).Scan(&token)
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("token kosong")
	}
	return token.String, nil
}

func LogoutUser(uid string) error {
	query := "UPDATE login_users SET refresh_token = NULL WHERE uid = $1"
	_, err := configs.DB.Exec(query, uid)
	return err
}

// --- BAGIAN CRUD USER ---

func CreateUser(req *CreateUserRequest) (*UserResponse, error) {
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	var uid string
	var createdAt, updatedAt time.Time

	query := `
		INSERT INTO login_users (username, pass, role_id) 
		VALUES ($1, $2, $3) 
		RETURNING uid, created_at, updated_at`

	err = configs.DB.QueryRow(query, req.Username, hashedPassword, req.RoleID).
		Scan(&uid, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	var roleName string
	_ = configs.DB.QueryRow("SELECT name FROM roles WHERE id = $1", req.RoleID).Scan(&roleName)

	return &UserResponse{
		UID:      uid,
		Username: req.Username,
		RoleName: roleName,
	}, nil
}

func GetUserByID(uid string) (*UserResponse, error) {
	var user UserResponse

	query := `
        SELECT u.uid, u.username, r.name
        FROM login_users u
        JOIN roles r ON u.role_id = r.id
        WHERE u.uid = $1`

	err := configs.DB.QueryRow(query, uid).
		Scan(&user.UID, &user.Username, &user.RoleName)

	if err == sql.ErrNoRows {
		// PERBAIKAN: Kembalikan error yang jelas jika tidak ditemukan
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user by ID: %w", err)
	}

	return &user, nil
}

func DeleteUser(uid string) error {
	res, err := configs.DB.Exec("DELETE FROM login_users WHERE uid = $1", uid)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func GetRoleIDByUID(uid string) (int, error) {
	var roleID int
	query := `SELECT role_id FROM login_users WHERE uid = $1`

	err := configs.DB.QueryRow(query, uid).Scan(&roleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("UID tidak ditemukan")
		}
		return 0, fmt.Errorf("gagal mendapatkan role ID: %w", err)
	}
	return roleID, nil
}

func GetAllUsers(page int, limit int, search string, roleID int) ([]UserResponse, int, error) {
	// Menghitung OFFSET
	offset := (page - 1) * limit

	// 1. Dynamic WHERE Clause Builder
	var whereClause []string
	var args []interface{}
	argCount := 1

	// Filter berdasarkan Role ID (Jika roleID > 0)
	if roleID > 0 {
		whereClause = append(whereClause, fmt.Sprintf("lu.role_id = $%d", argCount))
		args = append(args, roleID)
		argCount++
	}

	// Filter/Search (berdasarkan username atau full_name)
	if search != "" {
		searchPattern := "%" + search + "%"
		whereClause = append(whereClause, fmt.Sprintf("(lu.username ILIKE $%d OR p.full_name ILIKE $%d)", argCount, argCount))
		args = append(args, searchPattern)
		argCount++
	}

	// Menggabungkan WHERE Clause
	finalWhere := ""
	if len(whereClause) > 0 {
		finalWhere = " WHERE " + strings.Join(whereClause, " AND ")
	}

	// 2. Query untuk Menghitung Total Data (TotalCount)
	countQuery := fmt.Sprintf(`
        SELECT COUNT(lu.uid)
        FROM login_users lu
        LEFT JOIN person p ON lu.uid = p.uid
        %s`, finalWhere)

	var totalCount int
	err := configs.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total data: %w", err)
	}

	// Jika tidak ada data
	if totalCount == 0 {
		return []UserResponse{}, 0, nil
	}

	// 3. Query untuk Mengambil Data Aktual
	dataQuery := fmt.Sprintf(`
		SELECT 
			lu.uid, lu.username, lu.role_id, p.full_name, r.name
		FROM login_users lu                                 -- Start dari tabel 'lu' (jembatan)
		LEFT JOIN person p ON lu.uid = p.uid              -- Dapatkan full_name
		JOIN roles r ON lu.role_id = r.id                 -- Dapatkan role_name
        %s 
        ORDER BY lu.username ASC
        LIMIT $%d OFFSET $%d`, finalWhere, argCount, argCount+1)

	// Menambahkan parameter LIMIT dan OFFSET
	args = append(args, limit, offset)

	rows, err := configs.DB.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil data: %w", err)
	}
	defer rows.Close()

	var users []UserResponse

	for rows.Next() {
		var u UserResponse
		var roleID int
		var fullName sql.NullString
		var roleName sql.NullString

		err := rows.Scan(&u.UID, &u.Username, &roleID, &fullName, &roleName)
		if err != nil {
			return nil, 0, fmt.Errorf("gagal scan baris: %w", err)
		}

		// Mapping roleID dan full_name
		u.RoleID = roleID
		u.FullName = fullName.String // Mengambil nilai string jika tidak NULL
		u.RoleName = roleName.String

		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, totalCount, nil
}
