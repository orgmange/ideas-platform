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

	suite.authRepo = repository.NewAuthRepository(suite.DB)
	authUsecase := usecase.NewAuthUsecase(suite.authRepo, "test-secret")
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	authHandler := handlers.NewAuthHandler(authUsecase, logger)
	err = suite.DB.AutoMigrate(
		&models.User{},
		&models.BannedUser{},
		&models.Role{},
		&models.CoffeeShop{},
		&models.WorkerCoffeeShop{},
		&models.Category{},
		&models.Idea{},
		&models.IdeaLike{},
		&models.IdeaComment{},
		&models.Reward{},
		&models.RewardType{},
		&models.OTP{},
		&models.UserRefreshToken{},
	)
	if err != nil {
		logger.Error("Failed to auto-migrate database:", slog.String("error", err.Error()))
		return
	}
	userRepository := repository.NewUserRepository(suite.DB)
	userUsecase := usecase.NewUserUsecase(userRepository)
	userHandler := handlers.NewUserHandler(userUsecase, logger)

	coffeeShopRepo := repository.NewCoffeeShopRepository(suite.DB)
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

// TestGetOTP - табличный тест для получения OTP
func (suite *AuthIntegrationTestSuite) TestGetOTP() {
	tests := []struct {
		name           string
		phone          string
		expectedStatus int
		checkOTPInDB   bool
	}{
		{
			name:           "успешное получение OTP",
			phone:          "1234567890",
			expectedStatus: http.StatusNoContent,
			checkOTPInDB:   true,
		},
		{
			name:           "получение OTP для другого номера",
			phone:          "9876543210",
			expectedStatus: http.StatusNoContent,
			checkOTPInDB:   true,
		},
		{
			name:           "получение OTP для нового номера",
			phone:          "5555555555",
			expectedStatus: http.StatusNoContent,
			checkOTPInDB:   true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := authTestRequest{
				method: http.MethodGet,
				path:   fmt.Sprintf("/api/v1/auth/%s", tt.phone),
			}
			w := suite.makeRequest(req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.checkOTPInDB {
				var otp models.OTP
				err := suite.DB.First(&otp, "phone = ?", tt.phone).Error
				suite.NoError(err)
				suite.Equal(tt.phone, otp.Phone)
				suite.True(otp.ExpiresAt.After(time.Now()))
			}
		})
	}
}

// TestVerifyOTP - табличный тест для верификации OTP
func (suite *AuthIntegrationTestSuite) TestVerifyOTP() {
	tests := []struct {
		name           string
		phone          string
		otpCode        string
		userName       string
		existingUser   bool
		invalidOTP     bool
		expectedStatus int
		checkTokens    bool
	}{
		{
			name:           "новый пользователь с валидным OTP",
			phone:          "9876543210",
			otpCode:        "123456",
			userName:       "Test User",
			existingUser:   false,
			invalidOTP:     false,
			expectedStatus: http.StatusOK,
			checkTokens:    true,
		},
		{
			name:           "существующий пользователь с валидным OTP",
			phone:          "1122334455",
			otpCode:        "654321",
			userName:       "Existing User",
			existingUser:   true,
			invalidOTP:     false,
			expectedStatus: http.StatusOK,
			checkTokens:    true,
		},
		{
			name:           "невалидный OTP код",
			phone:          "1029384756",
			otpCode:        "111111",
			userName:       "",
			existingUser:   false,
			invalidOTP:     true,
			expectedStatus: http.StatusBadRequest,
			checkTokens:    false,
		},
		{
			name:           "новый пользователь с другим именем",
			phone:          "7778889990",
			otpCode:        "999888",
			userName:       "Another User",
			existingUser:   false,
			invalidOTP:     false,
			expectedStatus: http.StatusOK,
			checkTokens:    true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup: создаем OTP
			hashedCode, _ := bcrypt.GenerateFromPassword([]byte(tt.otpCode), bcrypt.DefaultCost)

			// Если нужен существующий пользователь
			if tt.existingUser {
				user := &models.User{
					Name:  tt.userName,
					Phone: tt.phone,
				}
				_, err := suite.authRepo.CreateUser(user)
				suite.NoError(err)
			}

			// Создаем OTP
			otp := &models.OTP{
				Phone:        tt.phone,
				CodeHash:     string(hashedCode),
				ExpiresAt:    time.Now().Add(5 * time.Minute),
				AttemptsLeft: 3,
			}
			suite.DB.Create(otp)

			// Подготавливаем неправильный код, если нужно
			requestOTP := tt.otpCode
			if tt.invalidOTP {
				requestOTP = "000000"
			}

			// Выполняем запрос
			reqBody := dto.VerifyOTPRequest{
				Phone: tt.phone,
				OTP:   requestOTP,
				Name:  tt.userName,
			}

			req := authTestRequest{
				method:      http.MethodPost,
				path:        "/api/v1/auth",
				body:        reqBody,
				contentType: "application/json",
			}
			w := suite.makeRequest(req)

			// Проверяем статус
			suite.Equal(tt.expectedStatus, w.Code)

			// Проверяем токены, если ожидается успех
			if tt.checkTokens {
				var authResponse dto.AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &authResponse)
				suite.NoError(err)
				suite.NotEmpty(authResponse.AccessToken)
				suite.NotEmpty(authResponse.RefreshToken)

				// Проверяем, что пользователь создан/существует
				var user models.User
				err = suite.DB.First(&user, "phone = ?", tt.phone).Error
				suite.NoError(err)
				if !tt.existingUser {
					suite.Equal(tt.userName, user.Name)
				}
			}
		})
	}
}

// TestRefresh - табличный тест для обновления токенов
func (suite *AuthIntegrationTestSuite) TestRefresh() {
	tests := []struct {
		name              string
		phone             string
		otpCode           string
		userName          string
		waitBeforeRefresh time.Duration
		expectedStatus    int
		checkNewTokens    bool
	}{
		{
			name:              "успешное обновление токена",
			phone:             "5544332211",
			otpCode:           "111222",
			userName:          "Refresh User",
			waitBeforeRefresh: 1 * time.Second,
			expectedStatus:    http.StatusOK,
			checkNewTokens:    true,
		},
		{
			name:              "обновление токена сразу после получения",
			phone:             "6655443322",
			otpCode:           "222333",
			userName:          "Quick Refresh",
			waitBeforeRefresh: 0,
			expectedStatus:    http.StatusOK,
			checkNewTokens:    true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Создаем пользователя и получаем первоначальные токены
			hashedCode, _ := bcrypt.GenerateFromPassword([]byte(tt.otpCode), bcrypt.DefaultCost)

			user := &models.User{
				Name:  tt.userName,
				Phone: tt.phone,
			}
			_, err := suite.authRepo.CreateUser(user)
			suite.NoError(err)

			otp := &models.OTP{
				Phone:        tt.phone,
				CodeHash:     string(hashedCode),
				ExpiresAt:    time.Now().Add(5 * time.Minute),
				AttemptsLeft: 3,
			}
			suite.DB.Create(otp)

			// Верифицируем OTP для получения токенов
			verifyReqBody := dto.VerifyOTPRequest{
				Phone: tt.phone,
				OTP:   tt.otpCode,
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

			// Ждем, если нужно
			if tt.waitBeforeRefresh > 0 {
				time.Sleep(tt.waitBeforeRefresh)
			}

			// Обновляем токен
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

			suite.Equal(tt.expectedStatus, refreshW.Code)

			if tt.checkNewTokens {
				var newAuthResponse dto.AuthResponse
				err = json.Unmarshal(refreshW.Body.Bytes(), &newAuthResponse)
				suite.NoError(err)
				suite.NotEmpty(newAuthResponse.AccessToken)
				suite.NotEmpty(newAuthResponse.RefreshToken)
				suite.NotEqual(authResp.AccessToken, newAuthResponse.AccessToken)
				suite.NotEqual(authResp.RefreshToken, newAuthResponse.RefreshToken)
			}
		})
	}
}

// TestAuthenticatedEndpoints - табличный тест для авторизованных эндпоинтов
func (suite *AuthIntegrationTestSuite) TestAuthenticatedEndpoints() {
	tests := []struct {
		name           string
		phone          string
		otpCode        string
		userName       string
		endpoint       string
		method         string
		withAuth       bool
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "получение данных текущего пользователя с авторизацией",
			phone:          "111222333",
			otpCode:        "123456",
			userName:       "Auth User",
			endpoint:       "/api/v1/users/me",
			method:         http.MethodGet,
			withAuth:       true,
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "получение данных без авторизации",
			phone:          "444555666",
			otpCode:        "654321",
			userName:       "No Auth",
			endpoint:       "/api/v1/users/me",
			method:         http.MethodGet,
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var token string
			if tt.withAuth {
				token = suite.getAuthToken(tt.phone, tt.otpCode, tt.userName)
			}

			req := authTestRequest{
				method: tt.method,
				path:   tt.endpoint,
			}

			var w *httptest.ResponseRecorder
			if tt.withAuth {
				w = suite.makeAuthenticatedRequest(req, token)
			} else {
				w = suite.makeRequest(req)
			}

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.checkResponse && tt.withAuth {
				var userResponse dto.UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &userResponse)
				suite.NoError(err)
				suite.Equal(tt.userName, userResponse.Name)
				suite.Equal(tt.phone, userResponse.Phone)
			}
		})
	}
}

// TestLogout - табличный тест для выхода
func (suite *AuthIntegrationTestSuite) TestLogout() {
	tests := []struct {
		name              string
		phone             string
		otpCode           string
		userName          string
		logoutEndpoint    string
		expectedStatus    int
		checkInvalidation bool
	}{
		{
			name:              "выход с одного устройства",
			phone:             "1234567890",
			otpCode:           "123456",
			userName:          "Logout User",
			logoutEndpoint:    "/api/v1/logout",
			expectedStatus:    http.StatusNoContent,
			checkInvalidation: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			token := suite.getAuthToken(tt.phone, tt.otpCode, tt.userName)

			logoutReq := authTestRequest{
				method: http.MethodPost,
				path:   tt.logoutEndpoint,
			}
			w := suite.makeAuthenticatedRequest(logoutReq, token)
			suite.Equal(tt.expectedStatus, w.Code)

			if tt.checkInvalidation {
				meReq := authTestRequest{
					method: http.MethodGet,
					path:   "/api/v1/users/me",
				}
				w = suite.makeAuthenticatedRequest(meReq, token)
				suite.Equal(http.StatusUnauthorized, w.Code)
			}
		})
	}
}

// TestLogoutEverywhere - табличный тест для выхода со всех устройств
func (suite *AuthIntegrationTestSuite) TestLogoutEverywhere() {
	tests := []struct {
		name               string
		phone              string
		otpCode            string
		userName           string
		numTokens          int
		expectedStatus     int
		checkAllInvalidate bool
	}{
		{
			name:               "выход со всех устройств (2 токена)",
			phone:              "1112223334",
			otpCode:            "123123",
			userName:           "Multi Device User",
			numTokens:          2,
			expectedStatus:     http.StatusNoContent,
			checkAllInvalidate: true,
		},
		{
			name:               "выход со всех устройств (3 токена)",
			phone:              "5556667778",
			otpCode:            "456456",
			userName:           "Many Devices User",
			numTokens:          3,
			expectedStatus:     http.StatusNoContent,
			checkAllInvalidate: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			tokens := make([]string, tt.numTokens)

			// Создаем несколько токенов
			for i := 0; i < tt.numTokens; i++ {
				tokens[i] = suite.getAuthToken(tt.phone, tt.otpCode, tt.userName)
				if i < tt.numTokens-1 {
					time.Sleep(1 * time.Second) // чтобы refresh токены отличались
				}
			}

			// Выходим со всех устройств используя первый токен
			logoutReq := authTestRequest{
				method: http.MethodPost,
				path:   "/api/v1/logout-everywhere",
			}
			w := suite.makeAuthenticatedRequest(logoutReq, tokens[0])
			suite.Equal(tt.expectedStatus, w.Code)

			// Проверяем, что все токены невалидны
			if tt.checkAllInvalidate {
				meReq := authTestRequest{
					method: http.MethodGet,
					path:   "/api/v1/users/me",
				}
				for i, token := range tokens {
					w := suite.makeAuthenticatedRequest(meReq, token)
					suite.Equal(http.StatusUnauthorized, w.Code,
						"токен %d должен быть невалидным", i+1)
				}
			}
		})
	}
}

// Helper functions

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

func (suite *AuthIntegrationTestSuite) getAuthToken(phone, otpCode, name string) string {
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)

	// Проверяем существование пользователя
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
		ExpiresAt:    time.Now().Add(5 * time.Second),
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

