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

type ErrUnauthorized struct {
	message string
}

func NewErrUnauthorized(message string) error {
	return &ErrUnauthorized{message: message}
}

func (err *ErrUnauthorized) Error() string {
	return err.message
}

type ErrAccessDenied struct {
	message string
}

func NewErrAccessDenied(message string) error {
	return &ErrAccessDenied{message: message}
}

func (err *ErrAccessDenied) Error() string {
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
