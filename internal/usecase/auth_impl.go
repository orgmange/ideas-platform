package usecase

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

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
}

func NewAuthUsecase(rep repository.AuthRepository, jwtSecret string) AuthUsecase {
	return &AuthUsecaseImpl{
		rep:       rep,
		jwtSecret: jwtSecret,
	}
}

// GetOTP implements AuthUsecase.
func (a *AuthUsecaseImpl) GetOTP(phone string) error {
	savedOTP, err := a.rep.GetOTP(phone)
	var errNotFound *apperrors.ErrNotFound
	if err != nil && !errors.As(err, &errNotFound) {
		return err
	}
	if savedOTP != nil {
		err = a.checkRateLimit(savedOTP)
		if err != nil {
			return err
		}

		a.updateResendCount(savedOTP)
	}
	code, err := generateCode()
	if err != nil {
		return err
	}

	hashedCode, err := hashCode(code)
	if err != nil {
		return err
	}

	err = a.saveOTP(savedOTP, hashedCode, phone)
	if err != nil {
		return err
	}
	err = sendOTPToPhone(phone, code)
	if err != nil {
		return err
	}

	return nil
}

func (*AuthUsecaseImpl) checkRateLimit(savedOTP *models.OTP) error {
	untilNextCode := time.Until(savedOTP.NextAllowedAt).Seconds()
	if untilNextCode > 0 {
		return apperrors.NewAuthErr(
			fmt.Sprintf("wait %d seconds before requesting new code",
				int(untilNextCode)))
	}
	return nil
}

func (a *AuthUsecaseImpl) saveOTP(savedOTP *models.OTP, hashedCode string, phone string) error {
	if savedOTP != nil {
		savedOTP.ExpiresAt = time.Now().Add(time.Minute * 5)
		savedOTP.AttemptsLeft = 3
		savedOTP.CodeHash = hashedCode

		err := a.rep.UpdateOTP(savedOTP)
		if err != nil {
			return err
		}
	} else {
		err := a.createOTP(phone, hashedCode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (*AuthUsecaseImpl) updateResendCount(savedOTP *models.OTP) {
	if time.Now().After(savedOTP.ExpiresAt.Add(24 * time.Hour)) {
		savedOTP.ResendCount = 0
	} else {
		savedOTP.ResendCount++
		if savedOTP.ResendCount < 3 {
			savedOTP.NextAllowedAt = time.Now().Add(time.Minute * 1)
		} else if savedOTP.ResendCount >= 3 && savedOTP.ResendCount < 5 {
			savedOTP.NextAllowedAt = time.Now().Add(time.Minute * 30)
		} else {
			savedOTP.NextAllowedAt = time.Now().Add(time.Hour * 24)
		}
	}
}

// VerifyOTP implements AuthUsecase.
func (a *AuthUsecaseImpl) VerifyOTP(req *dto.VerifyOTPRequest) (*dto.AuthResponse, error) {
	savedOTP, err := a.rep.GetOTP(req.Phone)
	if err != nil {
		return nil, err
	}
	if savedOTP.AttemptsLeft <= 0 {
		return nil, apperrors.NewAuthErr("too much attempts")
	}
	savedOTP.AttemptsLeft--
	defer a.rep.UpdateOTP(savedOTP)
	err = bcrypt.CompareHashAndPassword([]byte(savedOTP.CodeHash), []byte(req.OTP))
	if err != nil {
		return nil, apperrors.NewAuthErr("invalid credentials")
	}

	id, err := a.rep.GetUserIDByPhone(req.Phone)
	if err != nil {
		var errNotFound *apperrors.ErrNotFound
		if errors.As(err, &errNotFound) {
			id, err = a.createUser(req)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return a.makeAuthResponse(*id, "")
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

func (a *AuthUsecaseImpl) createUser(req *dto.VerifyOTPRequest) (*uuid.UUID, error) {
	if req.Name == "" {
		return nil, apperrors.NewErrNotValid("name can't be empty")
	}
	if req.Phone == "" {
		return nil, apperrors.NewErrNotValid("phone can't be empty")
	}

	user := &models.User{
		Phone: req.Phone,
		Name:  req.Name,
	}
	id, err := a.rep.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return id, nil
}

func (a *AuthUsecaseImpl) createOTP(phone, hashedCode string) error {
	otp := models.OTP{
		Phone:         phone,
		CodeHash:      hashedCode,
		ExpiresAt:     time.Now().Add(5 * time.Minute),
		AttemptsLeft:  3,
		NextAllowedAt: time.Now().Add(time.Minute * 1),
		ResendCount:   0,
		CreatedAt:     time.Now(),
	}
	err := a.rep.CreateOTP(&otp)
	if err != nil {
		return err
	}
	return nil
}

// Logout implements AuthUsecase.
func (a *AuthUsecaseImpl) Logout(tokenString string) error {
	return a.rep.DeleteRefreshToken(tokenString)
}

func (a *AuthUsecaseImpl) LogoutEverywhere(userID uuid.UUID) error {
	return a.rep.DeleteRefreshTokensByUserID(userID)
}

func (a *AuthUsecaseImpl) Refresh(oldTokenString string) (*dto.AuthResponse, error) {
	oldToken, err := a.validateAndGetRefreshToken(oldTokenString)
	if err != nil {
		return nil, err
	}

	return a.makeAuthResponse(oldToken.UserID, oldToken.RefreshToken)
}

func (a *AuthUsecaseImpl) makeAuthResponse(userID uuid.UUID, oldToken string) (*dto.AuthResponse, error) {
	jwtToken, err := a.createJWTToken(userID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := a.createRefreshToken(userID, oldToken)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  *jwtToken,
		RefreshToken: *refreshToken,
	}, nil
}

func (a *AuthUsecaseImpl) createRefreshToken(userID uuid.UUID, oldTokenString string) (*string, error) {
	if oldTokenString != "" {
		var errNotFound *apperrors.ErrNotFound
		err := a.rep.DeleteRefreshToken(oldTokenString)
		if err != nil && !errors.As(err, &errNotFound) {
			return nil, err
		}
	}

	newTokenString, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	hashedToken := hashToken(newTokenString)
	err = a.rep.CreateRefreshToken(&models.UserRefreshToken{
		UserID:       userID,
		RefreshToken: hashedToken,
		ExpiresAt:    time.Now().Add(time.Hour * 24 * 30),
	})
	if err != nil {
		return nil, err
	}

	return &newTokenString, nil
}

func (a *AuthUsecaseImpl) ValidateJWTToken(tokenString string) (*dto.JWTClaims, error) {
	var claims dto.JWTClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.NewAuthErr(fmt.Sprintf("unexpexted singing method: %v", token.Header["alg"]))
		}

		return []byte(a.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*dto.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, apperrors.NewAuthErr("invalid token")
}

func (a *AuthUsecaseImpl) validateAndGetRefreshToken(token string) (*models.UserRefreshToken, error) {
	hashedToken := hashToken(token)
	savedToken, err := a.rep.GetRefreshToken(hashedToken)
	if err != nil {
		return nil, err
	}

	if !savedToken.ExpiresAt.After(time.Now()) {
		return nil, apperrors.NewAuthErr("token expired")
	}
	return savedToken, nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (a *AuthUsecaseImpl) createJWTToken(userID uuid.UUID) (*string, error) {
	JWTClaims := dto.JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
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
