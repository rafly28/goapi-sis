// models/users_db.go
package models

import (
	"database/sql"
	"errors"
	"time"

	"go-sis-be/configs"
	"go-sis-be/utils"
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
		UID:       uid,
		Username:  req.Username,
		RoleName:  roleName,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func GetUserByID(uid string) (*UserResponse, error) {
	var user UserResponse

	query := `
		SELECT u.uid, u.username, r.name, u.created_at, u.updated_at
		FROM login_users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.uid = $1`

	err := configs.DB.QueryRow(query, uid).
		Scan(&user.UID, &user.Username, &user.RoleName, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetAllUsers(page int, limit int) (*PaginatedUserResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var totalItems int
	err := configs.DB.QueryRow("SELECT COUNT(*) FROM login_users").Scan(&totalItems)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT u.uid, u.username, r.name, u.created_at, u.updated_at
		FROM login_users u
		JOIN roles r ON u.role_id = r.id
		ORDER BY u.created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := configs.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []UserResponse{}
	for rows.Next() {
		var user UserResponse
		if err := rows.Scan(&user.UID, &user.Username, &user.RoleName, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + limit - 1) / limit
	}

	return &PaginatedUserResponse{
		Meta: PaginationMeta{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalItems:  totalItems,
			Limit:       limit,
		},
		Data: users,
	}, nil
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
