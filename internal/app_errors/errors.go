package apperrors

import (
	"fmt"

	"github.com/google/uuid"
)

type ErrNotFound struct {
	recordType string
	recordID   uuid.UUID
}

func NewErrNotFound(recordType string, recordID uuid.UUID) ErrNotFound {
	return ErrNotFound{recordType: recordType, recordID: recordID}
}

func (err ErrNotFound) Error() string {
	return fmt.Sprintf("%s with id: %s not found", err.recordType, err.recordID)
}

type ErrNotValid struct {
	message string
}

func NewErrNotValid(message string) ErrNotValid {
	return ErrNotValid{message: message}
}

func (err ErrNotValid) Error() string {
	return err.message
}
