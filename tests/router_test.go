package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
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
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type RouterTestSuite struct {
	suite.Suite
	DB             *gorm.DB
	Router         *gin.Engine
	cfg            *config.Config
	userRepo       repository.UserRep
	userRepository repository.UserRep
	authRepo       repository.AuthRepository
}

func (suite *RouterTestSuite) SetupSuite() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		suite.T().Fatalf("failed to load config: %v", err)
	}
	suite.cfg = cfg

	// Override with test DB config
	suite.cfg.DB.Host = "localhost"
	suite.cfg.DB.Port = 5433
	suite.cfg.DB.Name = "ideas_db_test"
	suite.cfg.DB.User = "postgres"
	suite.cfg.DB.Password = "postgres"

	// Connect to DB
	database, err := db.InitDB(suite.cfg)
	if err != nil {
		suite.T().Fatalf("failed to connect to db: %v", err)
	}
	suite.DB = database

	// Run migrations
	err = db.RunMigrations("file://../migrations", suite.cfg)
	if err != nil {
		suite.T().Fatalf("failed to run migrations: %v", err)
	}

	// Setup router
	sqlDB, err := suite.DB.DB()
	if err != nil {
		suite.T().Fatalf("failed to get sql.DB: %v", err)
	}
	suite.userRepository = repository.NewUserRepository(sqlDB)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	userUsecase := usecase.NewUserUsecase(suite.userRepository)
	userHandler := handlers.NewUserHandler(userUsecase, logger)

	coffeeShopRepo := repository.NewCoffeeShopRepository(suite.DB)
	csUscase := usecase.NewCoffeeShopUsecase(coffeeShopRepo)
	csHandler := handlers.NewCoffeeShopHandler(csUscase, logger)

	authRepo := repository.NewAuthRepository(suite.DB)
	authUsecase := usecase.NewAuthUsecase(authRepo, "test-secret")
	authHandler := handlers.NewAuthHandler(authUsecase, logger)
	suite.authRepo = authRepo

	appRouter := router.NewRouter(suite.cfg, userHandler, csHandler, authHandler, authUsecase, logger)
	suite.Router = appRouter.SetupRouter()
}

func (suite *RouterTestSuite) TearDownSuite() {
	// Close DB connection
	sqlDB, err := suite.DB.DB()
	if err != nil {
		suite.T().Fatalf("failed to get db instance: %v", err)
	}
	sqlDB.Close()
}

func (suite *RouterTestSuite) TearDownTest() {
	// Clean up the database after each test
	suite.DB.Exec("DELETE FROM users")
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}

func (suite *RouterTestSuite) makeAuthenticatedRequest(req routerTestRequest, token string) *httptest.ResponseRecorder {
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

func (suite *RouterTestSuite) getAuthToken() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	phone := fmt.Sprintf("7%09d", r.Intn(1000000000)) // unique 10-digit phone
	otpCode := "123456"
	name := "Test User"

	hashedCode, err := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)
	suite.Require().NoError(err)

	otp := &models.OTP{
		Phone:        phone,
		CodeHash:     string(hashedCode),
		ExpiresAt:    time.Now().Add(5 * time.Minute),
		AttemptsLeft: 3,
	}
	err = suite.DB.Create(otp).Error
	suite.Require().NoError(err)

	reqBody := dto.VerifyOTPRequest{
		Phone: phone,
		OTP:   otpCode,
		Name:  name,
	}

	w := suite.makeRequest(routerTestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/auth",
		body:        reqBody,
		contentType: "application/json",
	})
	suite.Require().Equal(http.StatusOK, w.Code, "Failed to get auth token. Body: %s", w.Body.String())

	var authResponse dto.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &authResponse)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(authResponse.AccessToken)

	return authResponse.AccessToken
}

func (suite *RouterTestSuite) TestHealthCheck() {
	tests := []struct {
		name           string
		expectedStatus int
		checkResponse  func(body []byte)
	}{
		{
			name:           "health check",
			expectedStatus: http.StatusOK,
			checkResponse: func(body []byte) {
				suite.Contains(string(body), "ok")
			},
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			testRequest := routerTestRequest{
				method: "GET",
				path:   "/health",
			}
			w := suite.makeRequest(testRequest)
			suite.Equal(test.expectedStatus, w.Code)
			test.checkResponse(w.Body.Bytes())
		})
	}
}

func (suite *RouterTestSuite) TestGetAllUsers() {
	type testCase struct {
		name           string
		setup          func()
		expectedStatus int
		checkResponse  func(body []byte)
	}

	tests := []testCase{
		{
			name: "get all with one user",
			setup: func() {
				_, err := suite.authRepo.CreateUser(&models.User{Name: "testuser", Phone: "12345"})
				suite.Require().NoError(err)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(body []byte) {
				var response []dto.UserResponse
				err := json.Unmarshal(body, &response)
				suite.NoError(err)
				suite.Len(response, 1)
				suite.Equal("testuser", response[0].Name)
			},
		},
		{
			name:           "get all with no users",
			setup:          func() { /* Do nothing, DB is clean */ },
			expectedStatus: http.StatusOK,
			checkResponse: func(body []byte) {
				var response []dto.UserResponse
				err := json.Unmarshal(body, &response)
				suite.NoError(err)
				suite.Len(response, 0)
			},
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			suite.TearDownTest() // Clean DB before each sub-test
			test.setup()
			testRequest := routerTestRequest{
				method: "GET",
				path:   "/api/v1/users",
			}
			w := suite.makeRequest(testRequest)
			suite.Equal(test.expectedStatus, w.Code)
			test.checkResponse(w.Body.Bytes())
		})
	}
}

func (suite *RouterTestSuite) TestGetUser() {
	type testCase struct {
		name           string
		setup          func() string // returns userID
		expectedStatus int
		checkResponse  func(body []byte, userID string)
	}

	tests := []testCase{
		{
			name: "get existing user",
			setup: func() string {
				userID, err := suite.authRepo.CreateUser(&models.User{Name: "testuser", Phone: "12345"})
				suite.Require().NoError(err)
				return userID.String()
			}, expectedStatus: http.StatusOK,
			checkResponse: func(body []byte, userID string) {
				var response dto.UserResponse
				err := json.Unmarshal(body, &response)
				suite.NoError(err)
				suite.Equal("testuser", response.Name)
				uid, err := uuid.Parse(userID)
				suite.NoError(err)
				suite.Equal(uid, response.ID)
			},
		},
		{
			name: "get non-existing user",
			setup: func() string {
				return uuid.New().String()
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(body []byte, userID string) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				suite.NoError(err)
				suite.Contains(response["error"], "not found")
			},
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			userID := test.setup()
			testRequest := routerTestRequest{
				method: "GET",
				path:   fmt.Sprintf("/api/v1/users/%s", userID),
			}
			w := suite.makeRequest(testRequest)
			suite.Equal(test.expectedStatus, w.Code)
			test.checkResponse(w.Body.Bytes(), userID)
		})
	}
}

func (suite *RouterTestSuite) TestUpdateUser() {
	type testCase struct {
		name           string
		setup          func() string // returns userID
		input          dto.UpdateUserRequest
		expectedStatus int
		checkResponse  func(w *httptest.ResponseRecorder, userID string)
	}
	tests := []testCase{
		{
			name: "update existing user",
			setup: func() string {
				userID, err := suite.authRepo.CreateUser(&models.User{Name: "testuser", Phone: "12345"})
				suite.Require().NoError(err)
				return userID.String()
			}, input: dto.UpdateUserRequest{
				Name: "updateduser",
			},
			expectedStatus: http.StatusNoContent,
			checkResponse: func(w *httptest.ResponseRecorder, userID string) {
				suite.Empty(w.Body.Bytes())

				getResp := suite.makeRequest(routerTestRequest{
					method: "GET",
					path:   fmt.Sprintf("/api/v1/users/%s", userID),
				})
				suite.Equal(http.StatusOK, getResp.Code)

				var response dto.UserResponse
				err := json.Unmarshal(getResp.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Equal("updateduser", response.Name)
			},
		},
		{
			name: "update non-existing user",
			setup: func() string {
				return uuid.New().String()
			},
			input: dto.UpdateUserRequest{
				Name: "updateduser",
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(w *httptest.ResponseRecorder, userID string) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], "not found")
			},
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			userID := test.setup()
			token := suite.getAuthToken()
			testRequest := routerTestRequest{
				method:      "PUT",
				path:        fmt.Sprintf("/api/v1/users/%s", userID),
				body:        test.input,
				contentType: "application/json",
			}
			w := suite.makeAuthenticatedRequest(testRequest, token)
			suite.Equal(test.expectedStatus, w.Code)
			test.checkResponse(w, userID)
		})
	}
}

func (suite *RouterTestSuite) TestDeleteUser() {
	tests := []struct {
		name           string
		setup          func() string // Returns userID for the test
		expectedStatus int
		checkResponse  func(w *httptest.ResponseRecorder)
	}{
		{
			name: "delete existing user",
			setup: func() string {
				userID, err := suite.authRepo.CreateUser(&models.User{Name: "testuser", Phone: "12345"})
				suite.Require().NoError(err)
				return userID.String()
			},
			expectedStatus: http.StatusNoContent,
			checkResponse: func(w *httptest.ResponseRecorder) {
				// No body
			},
		},
		{
			name: "delete non-existing user",
			setup: func() string {
				return uuid.New().String()
			},
			expectedStatus: http.StatusNoContent,
			checkResponse: func(w *httptest.ResponseRecorder) {
			},
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			userID := test.setup()
			token := suite.getAuthToken()
			testRequest := routerTestRequest{
				method: "DELETE",
				path:   fmt.Sprintf("/api/v1/users/%s", userID),
			}
			w := suite.makeAuthenticatedRequest(testRequest, token)
			suite.Equal(test.expectedStatus, w.Code)
			test.checkResponse(w)
		})
	}
}

type routerTestRequest struct {
	method      string
	path        string
	body        interface{}
	contentType string
}

func (suite *RouterTestSuite) makeRequest(req routerTestRequest) *httptest.ResponseRecorder {
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
