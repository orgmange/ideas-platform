package dto

import "github.com/google/uuid"

type CreateUserRequest struct {
	Name  string
	Phone string
}

type UpdateUserRequest struct {
	Name string
}

type UserResponse struct {
	ID    uuid.UUID
	Name  string
	Phone string
}
