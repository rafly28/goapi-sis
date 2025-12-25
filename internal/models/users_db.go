// models/users_db.go
package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
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
	query := "UPDATE login_users SET refresh_token = $1, updated_at = NOW() WHERE uid = $2::uuid"
	_, err := configs.DB.Exec(query, token, uid)
	if err != nil {
		log.Printf("Error UpdateRefreshToken: %v", err)
		return err
	}
	return nil
}

func GetRefreshToken(uid string) (string, error) {
	var storedToken string
	// Asumsi tabel: refresh_tokens (uid, token_string, expires_at)
	query := "SELECT token_string FROM refresh_tokens WHERE uid = $1"

	err := configs.DB.QueryRow(query, uid).Scan(&storedToken)
	if err == sql.ErrNoRows {
		// Jika token tidak ditemukan (sudah logout atau tidak pernah login), kembalikan string kosong
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("gagal mengambil refresh token: %w", err)
	}
	return storedToken, nil
}

// DeleteRefreshToken menghapus refresh token yang tersimpan, berfungsi sebagai mekanisme logout/revocation.
func DeleteRefreshToken(uid string) error {
	// PASTIKAN:
	// 1. Nama tabel ('refresh_tokens') sudah benar.
	// 2. Nama kolom UID ('uid' atau 'user_id' dll) sudah benar.
	query :=
		`UPDATE login_users 
        SET refresh_token = NULL 
        WHERE uid = $1`

	result, err := configs.DB.Exec(query, uid)
	if err != nil {
		// Log error di sini
		return fmt.Errorf("gagal menghapus refresh token: %w", err)
	}

	// (Opsional) Cek apakah baris benar-benar terpengaruh
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Log ini jika perlu, tapi ini bukan error
		log.Printf("Warning: No refresh token found for UID %s during delete.", uid)
	}

	return nil
}

func BlacklistToken(token string, expiry time.Duration) error {
	// Simpan ke Redis dengan TTL 15 menit
	// Format Key: "blacklist:<token>"
	err := configs.RedisClient.Set(configs.Ctx, "blacklist:"+token, "true", expiry).Err()
	return err
}

func IsTokenBlacklisted(token string) bool {
	val, err := configs.RedisClient.Exists(configs.Ctx, "blacklist:"+token).Result()
	if err != nil {
		log.Printf("Redis error checking blacklist: %v", err)
		return false
	}
	return val > 0
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

func GetUserSessionByUID(uid string) (*UserSession, error) {
	var sess UserSession
	var rt sql.NullString

	// Kita JOIN dengan tabel roles untuk mendapatkan nama role-nya
	query := `
		SELECT u.uid, u.username, r.name as role_name, u.refresh_token 
		FROM login_users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.uid = $1::uuid
	`

	err := configs.DB.QueryRow(query, uid).Scan(
		&sess.UID,
		&sess.Username,
		&sess.Role,
		&rt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user tidak ditemukan")
		}
		return nil, err
	}

	if rt.Valid {
		sess.RefreshToken = rt.String
	}

	return &sess, nil
}
