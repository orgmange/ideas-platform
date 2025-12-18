package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type AuthNormalizationTestSuite struct {
	BaseTestSuite
}

func TestAuthNormalizationTestSuite(t *testing.T) {
	suite.Run(t, new(AuthNormalizationTestSuite))
}

func (suite *AuthNormalizationTestSuite) TestPhoneNormalization() {
	phonePlus7 := "+79991112233"
	phone8 := "89991112233"
	phoneNormalized := "9991112233"
	otpCode := "123456"

	// 1. Get OTP with +7
	req := TestRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/v1/auth/%s", phonePlus7),
	}
	w := suite.MakeRequest(req)
	suite.Equal(http.StatusNoContent, w.Code)

	// Check DB for OTP - should be normalized
	var otp models.OTP
	err := suite.DB.First(&otp, "phone = ?", phoneNormalized).Error
	suite.NoError(err, "OTP should be stored with normalized phone")
	suite.Equal(phoneNormalized, otp.Phone)

	// Manually set code hash for verification
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)
	otp.CodeHash = string(hashedCode)
	suite.DB.Save(&otp)

	// 2. Verify OTP with +7 (Registration)
	verifyReqBody := dto.VerifyOTPRequest{
		Phone: phonePlus7,
		OTP:   otpCode,
		Name:  "Test User",
	}
	verifyReq := TestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/auth",
		body:        verifyReqBody,
		contentType: "application/json",
	}
	w = suite.MakeRequest(verifyReq)
	suite.Equal(http.StatusOK, w.Code)

	var authResp1 dto.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &authResp1)
	suite.NoError(err)

	// 3. Check User in DB - should be normalized
	var user models.User
	err = suite.DB.First(&user, "phone = ?", phoneNormalized).Error
	suite.NoError(err, "User should be stored with normalized phone")
	suite.Equal(phoneNormalized, *user.Phone)

	firstUserID := user.ID

	// 4. Get OTP with 8 (Login attempt)
	req = TestRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/v1/auth/%s", phone8),
	}
	w = suite.MakeRequest(req)
	suite.Equal(http.StatusNoContent, w.Code)

	// Check DB - OTP should be created/found using normalized phone
	otp = models.OTP{}
	err = suite.DB.First(&otp, "phone = ?", phoneNormalized).Error
	suite.NoError(err)
	
	// Manually set code hash again
	otp.CodeHash = string(hashedCode)
	suite.DB.Save(&otp)

	// 5. Verify OTP with 8
	verifyReqBody2 := dto.VerifyOTPRequest{
		Phone: phone8,
		OTP:   otpCode,
		Name:  "Test User",
	}
	verifyReq2 := TestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/auth",
		body:        verifyReqBody2,
		contentType: "application/json",
	}
	w = suite.MakeRequest(verifyReq2)
	suite.Equal(http.StatusOK, w.Code)

	// 6. Check that we logged in the SAME user (no new user created)
	var userCount int64
	suite.DB.Model(&models.User{}).Where("phone = ?", phoneNormalized).Count(&userCount)
	suite.Equal(int64(1), userCount, "Should be only 1 user with this phone")

	var user2 models.User
	err = suite.DB.First(&user2, "phone = ?", phoneNormalized).Error
	suite.NoError(err)
	suite.Equal(firstUserID, user2.ID, "UserID should match")
}
