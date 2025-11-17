package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/GeorgiiMalishev/ideas-platform/config"
	"github.com/GeorgiiMalishev/ideas-platform/internal/db"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/handlers"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/GeorgiiMalishev/ideas-platform/internal/repository"
	"github.com/GeorgiiMalishev/ideas-platform/internal/router"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthIntegrationTestSuite struct {
	suite.Suite
	DB       *gorm.DB
	cfg      *config.Config
	authRepo repository.AuthRepository
	Router   *gin.Engine
}

func (suite *AuthIntegrationTestSuite) SetupSuite() {
	cfg, err := config.Load()
	if err != nil {
		suite.T().Fatalf("failed to load config: %v", err)
	}
	suite.cfg = cfg

	suite.cfg.DB.Host = "localhost"
	suite.cfg.DB.Port = 5433
	suite.cfg.DB.Name = "ideas_db_test"
	suite.cfg.DB.User = "postgres"
	suite.cfg.DB.Password = "postgres"

	database, err := db.InitDB(suite.cfg)
	if err != nil {
		suite.T().Fatalf("failed to connect to db: %v", err)
	}
	suite.DB = database

	err = db.RunMigrations("file://../migrations", suite.cfg)
	if err != nil {
		suite.T().Fatalf("failed to run migrations: %v", err)
	}

	suite.authRepo = repository.NewAuthRepository(suite.DB)
	authUsecase := usecase.NewAuthUsecase(suite.authRepo, "test-secret")
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	authHandler := handlers.NewAuthHandler(authUsecase, logger)

	sqlDB, err := suite.DB.DB()
	if err != nil {
		suite.T().Fatalf("failed to get sql.DB: %v", err)
	}
	userRepository := repository.NewUserRepository(sqlDB)
	userUsecase := usecase.NewUserUsecase(userRepository)
	userHandler := handlers.NewUserHandler(userUsecase, logger)

	coffeeShopRepo := repository.NewCoffeeShopRepository(sqlDB)
	csUscase := usecase.NewCoffeeShopUsecase(coffeeShopRepo)
	csHandler := handlers.NewCoffeeShopHandler(csUscase, logger)

	appRouter := router.NewRouter(suite.cfg, userHandler, csHandler, authHandler, authUsecase, logger)
	suite.Router = appRouter.SetupRouter()
}

func (suite *AuthIntegrationTestSuite) TearDownSuite() {
	sqlDB, err := suite.DB.DB()
	if err != nil {
		suite.T().Fatalf("failed to get db instance: %v", err)
	}
	sqlDB.Close()
}

func (suite *AuthIntegrationTestSuite) TearDownTest() {
	suite.DB.Exec("DELETE FROM otps")
	suite.DB.Exec("DELETE FROM users")
	suite.DB.Exec("DELETE FROM user_refresh_tokens")
}

func TestAuthIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AuthIntegrationTestSuite))
}

type authTestRequest struct {
	method      string
	path        string
	body        interface{}
	contentType string
}

func (suite *AuthIntegrationTestSuite) makeRequest(req authTestRequest) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var bodyReader *bytes.Buffer

	if req.body != nil {
		bodyBytes, err := json.Marshal(req.body)
		suite.Require().NoError(err)
		bodyReader = bytes.NewBuffer(bodyBytes)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	httpReq, err := http.NewRequest(req.method, req.path, bodyReader)
	suite.Require().NoError(err)

	if req.contentType != "" {
		httpReq.Header.Set("Content-Type", req.contentType)
	}

	suite.Router.ServeHTTP(w, httpReq)
	return w
}

func (suite *AuthIntegrationTestSuite) TestGetOTP() {
	phone := "1234567890"

	req := authTestRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/v1/auth/%s", phone),
	}
	w := suite.makeRequest(req)

	suite.Equal(http.StatusNoContent, w.Code)

	var otp models.OTP
	err := suite.DB.First(&otp, "phone = ?", phone).Error
	suite.NoError(err)
	suite.Equal(phone, otp.Phone)
}

func (suite *AuthIntegrationTestSuite) TestVerifyOTP_NewUser() {
	phone := "9876543210"
	otpCode := "123456"
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)

	otp := &models.OTP{
		Phone:        phone,
		CodeHash:     string(hashedCode),
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		AttemptsLeft: 3,
	}
	suite.DB.Create(otp)

	reqBody := dto.VerifyOTPRequest{
		Phone: phone,
		OTP:   otpCode,
		Name:  "Test User",
	}

	req := authTestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/auth",
		body:        reqBody,
		contentType: "application/json",
	}
	w := suite.makeRequest(req)

	suite.Equal(http.StatusOK, w.Code)

	var authResponse dto.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &authResponse)
	suite.NoError(err)
	suite.NotEmpty(authResponse.AccessToken)
	suite.NotEmpty(authResponse.RefreshToken)

	var user models.User
	err = suite.DB.First(&user, "phone = ?", phone).Error
	suite.NoError(err)
	suite.Equal("Test User", user.Name)
}

func (suite *AuthIntegrationTestSuite) TestVerifyOTP_ExistingUser() {
	phone := "1122334455"
	otpCode := "654321"
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)

	user := &models.User{
		Name:  "Existing User",
		Phone: phone,
	}
	_, err := suite.authRepo.CreateUser(user)
	suite.NoError(err)

	otp := &models.OTP{
		Phone:        phone,
		CodeHash:     string(hashedCode),
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		AttemptsLeft: 3,
	}
	suite.DB.Create(otp)

	reqBody := dto.VerifyOTPRequest{
		Phone: phone,
		OTP:   otpCode,
	}

	req := authTestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/auth",
		body:        reqBody,
		contentType: "application/json",
	}
	w := suite.makeRequest(req)

	suite.Equal(http.StatusOK, w.Code)

	var authResponse dto.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &authResponse)
	suite.NoError(err)
	suite.NotEmpty(authResponse.AccessToken)
	suite.NotEmpty(authResponse.RefreshToken)
}

func (suite *AuthIntegrationTestSuite) TestVerifyOTP_InvalidOTP() {
	phone := "1029384756"
	otpCode := "111111"
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)

	otp := &models.OTP{
		Phone:        phone,
		CodeHash:     string(hashedCode),
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		AttemptsLeft: 3,
	}
	suite.DB.Create(otp)

	reqBody := dto.VerifyOTPRequest{
		Phone: phone,
		OTP:   "000000", // Invalid OTP
	}

	req := authTestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/auth",
		body:        reqBody,
		contentType: "application/json",
	}
	w := suite.makeRequest(req)

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *AuthIntegrationTestSuite) TestRefresh() {
	phone := "5544332211"
	otpCode := "111222"
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)

	user := &models.User{
		Name:  "Refresh User",
		Phone: phone,
	}
	_, err := suite.authRepo.CreateUser(user)
	suite.NoError(err)

	otp := &models.OTP{
		Phone:        phone,
		CodeHash:     string(hashedCode),
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		AttemptsLeft: 3,
	}
	suite.DB.Create(otp)

	// First, verify OTP to get a refresh token
	verifyReqBody := dto.VerifyOTPRequest{
		Phone: phone,
		OTP:   otpCode,
	}
	verifyReq := authTestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/auth",
		body:        verifyReqBody,
		contentType: "application/json",
	}
	verifyW := suite.makeRequest(verifyReq)

	suite.Equal(http.StatusOK, verifyW.Code)
	var authResp dto.AuthResponse
	err = json.Unmarshal(verifyW.Body.Bytes(), &authResp)
	suite.NoError(err)
	suite.NotNil(authResp)

	time.Sleep(1 * time.Second) // Ensure token timestamps are different

	// Now, use the refresh token
	refreshReqBody := dto.RefreshRequest{
		RefreshToken: authResp.RefreshToken,
	}
	refreshReq := authTestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/auth/refresh",
		body:        refreshReqBody,
		contentType: "application/json",
	}
	refreshW := suite.makeRequest(refreshReq)

	suite.Equal(http.StatusOK, refreshW.Code)

	var newAuthResponse dto.AuthResponse
	err = json.Unmarshal(refreshW.Body.Bytes(), &newAuthResponse)
	suite.NoError(err)
	suite.NotNil(newAuthResponse)
	suite.NotEmpty(newAuthResponse.AccessToken)
	suite.NotEmpty(newAuthResponse.RefreshToken)
		suite.NotEqual(authResp.AccessToken, newAuthResponse.AccessToken)
		suite.NotEqual(authResp.RefreshToken, newAuthResponse.RefreshToken)
	}
	
	func (suite *AuthIntegrationTestSuite) TestGetCurrentAuthentificatedUser_Authorized() {
		phone := "111222333"
		otpCode := "123456"
		userName := "Test User"
	
		// 1. Create user and get tokens
			token := suite.getAuthToken(phone, otpCode, userName)
		
			// 2. Make request to /users/me
			req := authTestRequest{
				method: http.MethodGet,
				path:   "/api/v1/users/me",
			}
			w := suite.makeAuthenticatedRequest(req, token)	
		// 3. Assert response
		suite.Equal(http.StatusOK, w.Code)
	
		var userResponse dto.UserResponse
		err := json.Unmarshal(w.Body.Bytes(), &userResponse)
		suite.NoError(err)
		suite.Equal(userName, userResponse.Name)
		suite.Equal(phone, userResponse.Phone)
	}
	
	func (suite *AuthIntegrationTestSuite) TestGetCurrentAuthentificatedUser_Unauthorized() {
		req := authTestRequest{
			method: http.MethodGet,
			path:   "/api/v1/users/me",
		}
		w := suite.makeRequest(req)
		suite.Equal(http.StatusUnauthorized, w.Code)
	}	
	func (suite *AuthIntegrationTestSuite) TestLogout() {
		// 1. Get auth token
			token := suite.getAuthToken("1234567890", "123456", "Test User")
		
			// 2. Logout
			logoutReq := authTestRequest{
				method: http.MethodPost,
				path:   "/api/v1/logout",
			}
			w := suite.makeAuthenticatedRequest(logoutReq, token)
			suite.Equal(http.StatusNoContent, w.Code)
		
			// 3. Verify token is invalidated
			meReq := authTestRequest{
				method: http.MethodGet,
				path:   "/api/v1/users/me",
			}
			w = suite.makeAuthenticatedRequest(meReq, token)
			suite.Equal(http.StatusUnauthorized, w.Code)
		}	
	func (suite *AuthIntegrationTestSuite) TestLogoutEverywhere() {
		phone := "1112223334"
		otpCode := "123123"
		userName := "Logout Everywhere"
	
		// 1. Get first token
		token1 := suite.getAuthToken(phone, otpCode, userName)
		time.Sleep(1 * time.Second) // ensure refresh token is different
		// 2. Get second token (simulating another login)
			token2 := suite.getAuthToken(phone, otpCode, userName)
		
			// 3. Logout from all devices using the first token
			logoutReq := authTestRequest{
				method: http.MethodPost,
				path:   "/api/v1/logout-everywhere",
			}
			w := suite.makeAuthenticatedRequest(logoutReq, token1)
			suite.Equal(http.StatusNoContent, w.Code)
		
			// 4. Verify both tokens are invalidated
			meReq := authTestRequest{
				method: http.MethodGet,
				path:   "/api/v1/users/me",
			}
			w1 := suite.makeAuthenticatedRequest(meReq, token1)
			suite.Equal(http.StatusUnauthorized, w1.Code)
		
			w2 := suite.makeAuthenticatedRequest(meReq, token2)
			suite.Equal(http.StatusUnauthorized, w2.Code)
		}	
	// Helper to make authenticated requests
	func (suite *AuthIntegrationTestSuite) makeAuthenticatedRequest(req authTestRequest, token string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		var bodyReader *bytes.Buffer
	
		if req.body != nil {
			bodyBytes, err := json.Marshal(req.body)
			suite.Require().NoError(err)
			bodyReader = bytes.NewBuffer(bodyBytes)
		} else {
			bodyReader = bytes.NewBuffer(nil)
		}
	
		httpReq, err := http.NewRequest(req.method, req.path, bodyReader)
		suite.Require().NoError(err)
	
		if req.contentType != "" {
			httpReq.Header.Set("Content-Type", req.contentType)
		}
		httpReq.Header.Set("Authorization", "Bearer "+token)
	
		suite.Router.ServeHTTP(w, httpReq)
		return w
	}	
	// Helper to get a valid auth token
	func (suite *AuthIntegrationTestSuite) getAuthToken(phone, otpCode, name string) string {
		hashedCode, _ := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)
	
		// Ensure user exists or is created
		var user models.User
		err := suite.DB.First(&user, "phone = ?", phone).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				_, err = suite.authRepo.CreateUser(&models.User{Name: name, Phone: phone})
				suite.Require().NoError(err)
			} else {
				suite.Require().NoError(err)
			}
		}
	
		otp := &models.OTP{
			Phone:        phone,
			CodeHash:     string(hashedCode),
			ExpiresAt:    time.Now().Add(5 * time.Minute),
			AttemptsLeft: 3,
		}
		suite.DB.Create(otp)
	
			reqBody := dto.VerifyOTPRequest{
				Phone: phone,
				OTP:   otpCode,
				Name:  name,
			}
		
			req := authTestRequest{
				method:      http.MethodPost,
				path:        "/api/v1/auth",
				body:        reqBody,
				contentType: "application/json",
			}
			w := suite.makeRequest(req)	
		suite.Require().Equal(http.StatusOK, w.Code)
	
		var authResponse dto.AuthResponse
		err = json.Unmarshal(w.Body.Bytes(), &authResponse)
		suite.Require().NoError(err)
		return authResponse.AccessToken
	}
	