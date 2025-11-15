package usecase

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
)

type AuthUsecase interface {
	GetOTP(phone string) error
	VerifyOTP(req *dto.VerifyOTPRequest) (*string, error)
}
