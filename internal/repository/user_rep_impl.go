package repository

import (
	"database/sql"
	"fmt"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type UserRepImpl struct {
	db *sql.DB
}

// CreateUser implements IUserRep.

// DeleteUser implements IUserRep.
func (u UserRepImpl) DeleteUser(ID uuid.UUID) error {
	query := `UPDATE users SET is_deleted = true WHERE id = $1`
	_, err := u.db.Exec(query, ID)
	if err == sql.ErrNoRows {
		return apperrors.NewErrNotFound("user", ID.String())
	}
	return err
}

// GetAllUsers implements IUserRep.
func (u UserRepImpl) GetAllUsers(limit int, offset int) ([]models.User, error) {
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
func (u UserRepImpl) GetUser(ID uuid.UUID) (*models.User, error) {
	query := `SELECT id, name, phone, role_id, is_deleted, updated_at, created_at FROM users WHERE id = $1 AND is_deleted = false`
	var user models.User
	err := u.db.QueryRow(query, ID).Scan(&user.ID, &user.Name, &user.Phone, &user.RoleID, &user.IsDeleted, &user.UpdatedAt, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewErrNotFound("user", ID.String())
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser implements IUserRep.
func (u UserRepImpl) UpdateUser(user *models.User) error {
	query := `UPDATE users SET name = $1, phone = $2, role_id = $3, updated_at = NOW() WHERE id = $4`
	_, err := u.db.Exec(query, user.Name, user.Phone, user.RoleID, user.ID)
	if err == sql.ErrNoRows {
		return apperrors.NewErrNotFound("user", user.ID.String())
	}
	return err
}

func (u UserRepImpl) IsUserExist(ID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`

	var exists bool
	err := u.db.QueryRow(query, ID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

func NewUserRepository(db *sql.DB) UserRep {
	return &UserRepImpl{db: db}
}
