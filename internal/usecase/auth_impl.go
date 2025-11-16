package usecase

import (
	"crypto/rand"
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
func (a *AuthUsecaseImpl) VerifyOTP(req *dto.VerifyOTPRequest) (*string, error) {
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

	JWTClaims := models.JWTClaims{
		UserID:       id.String(),
		RefreshToken: "",
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
