package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/GeorgiiMalishev/ideas-platform/config"
	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecaseImpl struct {
	rep       repository.AuthRepository
	jwtSecret string
	authCfg   *config.AuthConfig
	logger    *slog.Logger
}

func NewAuthUsecase(rep repository.AuthRepository, jwtSecret string, authCfg *config.AuthConfig, logger *slog.Logger) AuthUsecase {
	return &AuthUsecaseImpl{
		rep:       rep,
		jwtSecret: jwtSecret,
		authCfg:   authCfg,
		logger:    logger,
	}
}

// GetOTP implements AuthUsecase.
func (a *AuthUsecaseImpl) GetOTP(ctx context.Context, phone string) error {
	logger := a.logger.With(
		"method", "GetOTP",
		"phone", phone,
	)

	logger.Debug("starting OTP generation")

	savedOTP, err := a.rep.GetOTP(ctx, phone)
	var errNotFound *apperrors.ErrNotFound
	if err != nil && !errors.As(err, &errNotFound) {
		logger.Error("failed to get OTP from repository", "error", err.Error())
		return err
	}

	if savedOTP != nil {
		logger.Info("existing OTP found",
			"otp_id", savedOTP.ID,
			"resend_count", savedOTP.ResendCount)

		err = a.checkRateLimit(savedOTP)
		if err != nil {
			logger.Info("rate limit exceeded", "error", err.Error())
			return err
		}

		a.updateResendCount(savedOTP)
		logger.Debug("resend count updated", "new_resend_count", savedOTP.ResendCount)
	}

	code, err := generateCode()
	if err != nil {
		logger.Error("failed to generate OTP code", "error", err.Error())
		return err
	}

	hashedCode, err := hashCode(code)
	if err != nil {
		logger.Error("failed to hash OTP code", "error", err.Error())
		return err
	}

	err = a.saveOTP(ctx, savedOTP, hashedCode, phone)
	if err != nil {
		logger.Error("failed to save OTP", "error", err.Error())
		return err
	}

	logger.Info("sending OTP to phone")

	err = sendOTPToPhone(phone, code)
	if err != nil {
		logger.Error("failed to send OTP", "error", err.Error())
		return err
	}

	logger.Info("OTP sent successfully")
	return nil
}

func (*AuthUsecaseImpl) checkRateLimit(savedOTP *models.OTP) error {
	untilNextCode := time.Until(savedOTP.NextAllowedAt).Seconds()
	if untilNextCode > 0 {
		return apperrors.NewErrRateLimit(
			fmt.Sprintf("wait %d seconds before requesting new code",
				int(untilNextCode)))
	}
	return nil
}

func (a *AuthUsecaseImpl) saveOTP(ctx context.Context, savedOTP *models.OTP, hashedCode string, phone string) error {
	if savedOTP != nil {
		savedOTP.ExpiresAt = time.Now().Add(a.authCfg.OTPConfig.ExpiresAtTimer)
		savedOTP.AttemptsLeft = a.authCfg.OTPConfig.AttemptsLeft
		savedOTP.CodeHash = hashedCode

		err := a.rep.UpdateOTP(ctx, savedOTP)
		if err != nil {
			return err
		}
	} else {
		err := a.createOTP(ctx, phone, hashedCode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *AuthUsecaseImpl) updateResendCount(savedOTP *models.OTP) {
	if time.Now().After(savedOTP.ExpiresAt.Add(a.authCfg.OTPConfig.ResetResendCountTimer)) {
		savedOTP.ResendCount = 0
	} else {
		savedOTP.ResendCount++
		if savedOTP.ResendCount < a.authCfg.OTPConfig.SoftAttemptsCount {
			savedOTP.NextAllowedAt = time.Now().Add(a.authCfg.OTPConfig.SubSoftAttemptsTimer)
		} else if savedOTP.ResendCount >= a.authCfg.OTPConfig.SoftAttemptsCount && savedOTP.ResendCount < a.authCfg.OTPConfig.HardAttemptsCount {
			savedOTP.NextAllowedAt = time.Now().Add(a.authCfg.OTPConfig.SubHardAttemptsTimer)
		} else {
			savedOTP.NextAllowedAt = time.Now().Add(a.authCfg.OTPConfig.PostHardAttemptsCount)
		}
	}
}

// VerifyOTP implements AuthUsecase.
func (a *AuthUsecaseImpl) VerifyOTP(ctx context.Context, req *dto.VerifyOTPRequest) (*dto.AuthResponse, error) {
	logger := a.logger.With(
		"method", "GetOTP",
		"phone", req.Phone,
	)

	logger.Debug("Starting virify OTP")

	savedOTP, err := a.rep.GetOTP(ctx, req.Phone)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("Saved OTP not found for this phone")
		} else {
			logger.Error("failed to get OTP from repository: ", "error", err.Error())
		}
		return nil, err
	}
	if savedOTP.AttemptsLeft <= 0 {
		logger.Info("no attempts left to virify OTP")
		return nil, apperrors.NewErrUnauthorized("too much attempts")
	}
	savedOTP.AttemptsLeft--
	err = bcrypt.CompareHashAndPassword([]byte(savedOTP.CodeHash), []byte(req.OTP))
	if err != nil {
		logger.Info("sended OTP code not match with saved")
		a.rep.UpdateOTP(ctx, savedOTP)
		return nil, apperrors.NewErrUnauthorized("invalid credentials")
	}

	user, err := a.rep.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("user with this phone not found, creating")
			user, err = a.createUser(ctx, req)
			if err != nil {
				return nil, err
			}
		} else {
			logger.Error("failed to get user id by phone", "error", err.Error())
			return nil, err
		}
	}
	err = a.rep.DeleteOTP(ctx, req.Phone)
	if err != nil {
		logger.Error("failed to delete OTP", "error", err.Error())
		return nil, err
	}

	return a.makeAuthResponse(ctx, user, "")
}

func generateCode() (string, error) {
	digits := "0123456789"
	result := make([]byte, 6)

	for i := range 6 {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		result[i] = digits[num.Int64()]
	}

	return string(result), nil
}

func sendOTPToPhone(phone string, code string) error {
	fmt.Printf("code %s for phone %s\n", code, phone)
	return nil
}

func hashCode(code string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	return string(hash), err
}

func (a *AuthUsecaseImpl) createUser(ctx context.Context, req *dto.VerifyOTPRequest) (*models.User, error) {
	logger := a.logger.With(
		"method", "createUser",
		"phone", req.Phone,
	)

	logger.Debug("starting create generation")
	if req.Name == "" {
		logger.Info("name not valid")
		return nil, apperrors.NewErrNotValid("name can't be empty")
	}
	if req.Phone == "" {
		logger.Info("phone not valid")
		return nil, apperrors.NewErrNotValid("phone can't be empty")
	}

	user := &models.User{
		Phone: req.Phone,
		Name:  req.Name,
	}
	savedUser, err := a.rep.CreateUser(ctx, user)
	if err != nil {
		logger.Error("failed create user", "error", err.Error())
		return nil, err
	}

	logger.Info("user created successfully")
	return savedUser, nil
}

func (a *AuthUsecaseImpl) createOTP(ctx context.Context, phone, hashedCode string) error {
	otp := models.OTP{
		Phone:         phone,
		CodeHash:      hashedCode,
		ExpiresAt:     time.Now().Add(a.authCfg.OTPConfig.ExpiresAtTimer),
		AttemptsLeft:  3,
		NextAllowedAt: time.Now().Add(a.authCfg.OTPConfig.SubSoftAttemptsTimer),
		ResendCount:   0,
		CreatedAt:     time.Now(),
	}
	err := a.rep.CreateOTP(ctx, &otp)
	if err != nil {
		return err
	}
	return nil
}

// Logout implements AuthUsecase.
func (a *AuthUsecaseImpl) Logout(ctx context.Context, tokenString string) error {
	hashedToken := hashToken(tokenString)
	return a.rep.DeleteRefreshToken(ctx, hashedToken)
}

func (a *AuthUsecaseImpl) LogoutEverywhere(ctx context.Context, userID uuid.UUID) error {
	return a.rep.DeleteRefreshTokensByUserID(ctx, userID)
}

func (a *AuthUsecaseImpl) Refresh(ctx context.Context, oldTokenString string) (*dto.AuthResponse, error) {
	oldToken, err := a.validateAndGetRefreshToken(ctx, oldTokenString)
	if err != nil {
		return nil, err
	}

	if oldToken.User == nil {
		a.logger.Warn("refresh token exists but user not found", "token_hash", oldTokenString)
		_ = a.rep.DeleteRefreshToken(ctx, oldTokenString)
		return nil, apperrors.NewErrUnauthorized("user not found")
	}

	return a.makeAuthResponse(ctx, oldToken.User, oldToken.RefreshToken)
}

func (a *AuthUsecaseImpl) makeAuthResponse(ctx context.Context, user *models.User, oldToken string) (*dto.AuthResponse, error) {
	logger := a.logger.With("method", "makeAuthResponse", "userID", user.ID.String())

	logger.Debug("starting make auth response")
	jwtToken, err := a.createJWTToken(user)
	if err != nil {
		logger.Error("failed to create JWT token", "error", err.Error())
		return nil, err
	}

	refreshToken, err := a.createRefreshToken(ctx, user.ID, oldToken)
	if err != nil {
		logger.Error("failed to create refresh token token", "error", err.Error())
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  *jwtToken,
		RefreshToken: *refreshToken,
	}, nil
}

func (a *AuthUsecaseImpl) createRefreshToken(ctx context.Context, userID uuid.UUID, oldTokenString string) (*string, error) {
	logger := a.logger.With(
		"method", "createRefreshToken",
		"userID", userID.String(),
	)

	logger.Debug("starting create refresh token")
	if oldTokenString != "" {
		var errNotFound *apperrors.ErrNotFound
		err := a.rep.DeleteRefreshToken(ctx, oldTokenString)
		if err != nil && !errors.As(err, &errNotFound) {
			logger.Error("failed to delete old token", "error", err.Error())
			return nil, err
		}
	}

	newTokenString, err := generateRefreshToken()
	if err != nil {
		logger.Error("failed to generate refresh token", "error", err.Error())
		return nil, err
	}
	hashedToken := hashToken(newTokenString)
	err = a.rep.CreateRefreshToken(ctx, &models.UserRefreshToken{
		UserID:       userID,
		RefreshToken: hashedToken,
		ExpiresAt:    time.Now().Add(a.authCfg.JWTConfig.RefreshTokenTimer),
	})
	if err != nil {
		logger.Error("failed to create refresh token", "error", err.Error())
		return nil, err
	}

	return &newTokenString, nil
}

func (a *AuthUsecaseImpl) ValidateJWTToken(ctx context.Context, tokenString string) (*dto.JWTClaims, error) {
	logger := a.logger.With(
		"method", "ValidateJWTToken",
	)

	logger.Debug("starting validate JWT token")
	var claims dto.JWTClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Info(fmt.Sprintf("unexpexted singing method: %v", token.Header["alg"]))
			return nil, apperrors.NewErrUnauthorized(fmt.Sprintf("unexpexted singing method: %v", token.Header["alg"]))
		}

		return []byte(a.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		logger.Info("invalid token")
		return nil, apperrors.NewErrUnauthorized("invalid token")
	}
	logger.Info("JWT claims successfully getted", "userID", claims.UserID.String())
	return &claims, nil
}

func (a *AuthUsecaseImpl) validateAndGetRefreshToken(ctx context.Context, token string) (*models.UserRefreshToken, error) {
	logger := a.logger.With(
		"method", "validateAndGetRefreshToken",
		"token", token,
	)

	logger.Debug("starting validate and refresh token")
	hashedToken := hashToken(token)
	savedToken, err := a.rep.GetRefreshToken(ctx, hashedToken)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			logger.Info("refresh token not found")
			return nil, apperrors.NewErrUnauthorized("refresh token not found")
		}

		logger.Error("failed to get refresh token", "error", err.Error())
		return nil, err
	}

	if !savedToken.ExpiresAt.After(time.Now()) {
		logger.Info("token expired")
		return nil, apperrors.NewErrUnauthorized("token expired")
	}
	return savedToken, nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (a *AuthUsecaseImpl) createJWTToken(user *models.User) (*string, error) {
	JWTClaims := dto.JWTClaims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.authCfg.JWTConfig.JWTTokenTimer)),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims)
	tokenString, err := jwtToken.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return nil, err
	}
	return &tokenString, nil
}

func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}
