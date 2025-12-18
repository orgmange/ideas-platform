package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/stretchr/testify/suite"
)

type AdminLoginTestSuite struct {
	BaseTestSuite
}

func (suite *AdminLoginTestSuite) TestLoginAdmin_Success() {
	w := httptest.NewRecorder()

	// 1. Register Admin
	testLogin := "admin_login_test_user"
	testPassword := "securepassword"
	registerReq := dto.RegisterAdminRequest{
		Login:          testLogin,
		Password:       testPassword,
		CoffeeShopName: "Test Coffee Shop",
		Address:        "123 Test St",
	}
	jsonValue, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest("POST", "/api/v1/auth/register/admin", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	suite.Router.ServeHTTP(w, req)
	suite.Equal(http.StatusOK, w.Code)

	// 2. Login Admin
	w = httptest.NewRecorder()
	loginReq := dto.AdminLoginRequest{
		Login:    testLogin,
		Password: testPassword,
	}
	jsonValue, _ = json.Marshal(loginReq)

	req, _ = http.NewRequest("POST", "/api/v1/auth/login/admin", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	suite.Router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var authResp dto.AdminAuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &authResp)
	suite.NoError(err)
	suite.NotEmpty(authResp.AccessToken)
	suite.NotEmpty(authResp.RefreshToken)
	suite.NotEmpty(authResp.CoffeeShopID)
}

func (suite *AdminLoginTestSuite) TestLoginAdmin_InvalidCredentials() {
	w := httptest.NewRecorder()

	// 1. Register Admin
	testLogin := "admin_login_invalid_creds_user"
	testPassword := "securepassword"
	registerReq := dto.RegisterAdminRequest{
		Login:          testLogin,
		Password:       testPassword,
		CoffeeShopName: "Test Coffee Shop",
		Address:        "123 Test St",
	}
	jsonValue, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest("POST", "/api/v1/auth/register/admin", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	suite.Router.ServeHTTP(w, req)
	suite.Equal(http.StatusOK, w.Code)

	// 2. Login Admin with wrong password
	w = httptest.NewRecorder()
	loginReq := dto.AdminLoginRequest{
		Login:    testLogin,
		Password: "wrongpassword",
	}
	jsonValue, _ = json.Marshal(loginReq)

	req, _ = http.NewRequest("POST", "/api/v1/auth/login/admin", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	suite.Router.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)
}

func (suite *AdminLoginTestSuite) TestLoginAdmin_UserNotFound() {
	w := httptest.NewRecorder()

	loginReq := dto.AdminLoginRequest{
		Login:    "nonexistentuser",
		Password: "password",
	}
	jsonValue, _ := json.Marshal(loginReq)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login/admin", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	suite.Router.ServeHTTP(w, req)

	suite.Equal(http.StatusUnauthorized, w.Code)
}

func TestAdminLoginTestSuite(t *testing.T) {
	suite.Run(t, new(AdminLoginTestSuite))
}
