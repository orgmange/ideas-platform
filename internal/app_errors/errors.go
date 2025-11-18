package apperrors

import (
	"fmt"
)

type ErrNotFound struct {
	recordType string
	recordID   string
}

func NewErrNotFound(recordType string, recordID string) error {
	return &ErrNotFound{recordType: recordType, recordID: recordID}
}

func (err *ErrNotFound) Error() string {
	return fmt.Sprintf("%s with id: %s not found", err.recordType, err.recordID)
}

type ErrNotValid struct {
	message string
}

func NewErrNotValid(message string) error {
	return &ErrNotValid{message: message}
}

func (err *ErrNotValid) Error() string {
	return err.message
}

type AuthErr struct {
	message string
}

func NewAuthErr(message string) error {
	return &AuthErr{message: message}
}

func (err *AuthErr) Error() string {
	return err.message
}

type ErrRateLimit struct {
	message string
}

func NewErrRateLimit(message string) error {
	return &ErrRateLimit{message: message}
}

func (err *ErrRateLimit) Error() string {
	return err.message
}
