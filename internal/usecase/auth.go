package usecase

import (
	"context"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type AuthUsecase interface {

	GetOTP(ctx context.Context, phone string) error

	VerifyOTP(ctx context.Context, req *dto.VerifyOTPRequest) (*dto.AuthResponse, error)

	RegisterAdminAndCoffeeShop(ctx context.Context, req *dto.RegisterAdminRequest) (*dto.AdminAuthResponse, error)
	LoginAdmin(ctx context.Context, req *dto.AdminLoginRequest) (*dto.AdminAuthResponse, error)
	Refresh(ctx context.Context, token string) (*dto.AuthResponse, error)

	Logout(ctx context.Context, token string) error

	LogoutEverywhere(ctx context.Context, userID uuid.UUID) error

	ValidateJWTToken(ctx context.Context, tokenString string) (*dto.JWTClaims, error)

}
