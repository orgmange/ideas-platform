package repository

import (
	"database/sql"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type UserRep struct {
	db *sql.DB
}

// CreateUser implements IUserRep.
func (u UserRep) CreateUser(user *models.User) (*models.User, error) {
	query := `INSERT INTO users (name, phone, role_id) VALUES ($1, $2, $3) RETURNING id, name, phone, role_id, is_deleted, updated_at, created_at`
	var createdUser models.User
	err := u.db.QueryRow(query, user.Name, user.Phone, user.RoleID).Scan(&createdUser.ID, &createdUser.Name, &createdUser.Phone, &createdUser.RoleID, &createdUser.IsDeleted, &createdUser.UpdatedAt, &createdUser.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &createdUser, nil
}

// DeleteUser implements IUserRep.
func (u UserRep) DeleteUser(ID uuid.UUID) error {
	query := `UPDATE users SET is_deleted = true WHERE id = $1`
	_, err := u.db.Exec(query, ID)
	if err == sql.ErrNoRows {
		return apperrors.NewErrNotFound("user", ID)
	}
	return err
}

// GetAllUsers implements IUserRep.
func (u UserRep) GetAllUsers(limit int, offset int) ([]models.User, error) {
	query := `SELECT id, name, phone, role_id, is_deleted, updated_at, created_at FROM users WHERE is_deleted = false LIMIT $1 OFFSET $2`
	rows, err := u.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Phone, &user.RoleID, &user.IsDeleted, &user.UpdatedAt, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// GetUser implements IUserRep.
func (u UserRep) GetUser(ID uuid.UUID) (*models.User, error) {
	query := `SELECT id, name, phone, role_id, is_deleted, updated_at, created_at FROM users WHERE id = $1 AND is_deleted = false`
	var user models.User
	err := u.db.QueryRow(query, ID).Scan(&user.ID, &user.Name, &user.Phone, &user.RoleID, &user.IsDeleted, &user.UpdatedAt, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewErrNotFound("user", ID)
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser implements IUserRep.
func (u UserRep) UpdateUser(user *models.User) error {
	query := `UPDATE users SET name = $1, phone = $2, role_id = $3, updated_at = NOW() WHERE id = $4`
	_, err := u.db.Exec(query, user.Name, user.Phone, user.RoleID, user.ID)
	if err == sql.ErrNoRows {
		return apperrors.NewErrNotFound("user", user.ID)
	}
	return err
}

func NewUserRepository(db *sql.DB) IUserRep {
	return &UserRep{db: db}
}
