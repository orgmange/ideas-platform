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
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// BaseTestSuite is a base suite for integration tests
type BaseTestSuite struct {
	suite.Suite
	DB             *gorm.DB
	Router         *gin.Engine
	cfg            *config.Config
	AuthRepo       repository.AuthRepository
	UserRepo       repository.UserRep
	CoffeeShopRepo repository.CoffeeShopRep
}

// TestRequest is a helper struct for making requests
type TestRequest struct {
	method      string
	path        string
	body        interface{}
	contentType string
	token       string
}

// SetupSuite sets up the test suite
func (suite *BaseTestSuite) SetupSuite() {
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

	// Use a short JWT token timer for tests
	suite.cfg.AuthConfig.JWTConfig.JWTTokenTimer = time.Second * 2

	database, err := db.InitDB(suite.cfg)
	if err != nil {
		suite.T().Fatalf("failed to connect to db: %v", err)
	}
	suite.DB = database

	// Using AutoMigrate for tests to ensure schema is up-to-date
	err = suite.DB.AutoMigrate(
		&models.User{}, &models.BannedUser{}, &models.Role{},
		&models.CoffeeShop{}, &models.WorkerCoffeeShop{}, &models.Category{},
		&models.Idea{}, &models.IdeaLike{}, &models.IdeaComment{},
		&models.Reward{}, &models.RewardType{}, &models.OTP{},
		&models.UserRefreshToken{},
	)
	if err != nil {
		suite.T().Fatalf("failed to auto-migrate database: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Repositories
	suite.AuthRepo = repository.NewAuthRepository(suite.DB)
	suite.UserRepo = repository.NewUserRepository(suite.DB)
	suite.CoffeeShopRepo = repository.NewCoffeeShopRepository(suite.DB)

	// Usecases
	authUsecase := usecase.NewAuthUsecase(suite.AuthRepo, "test-secret", &suite.cfg.AuthConfig, logger)
	userUsecase := usecase.NewUserUsecase(suite.UserRepo, logger)
	csUscase := usecase.NewCoffeeShopUsecase(suite.CoffeeShopRepo, logger)

	// Handlers
	authHandler := handlers.NewAuthHandler(authUsecase, logger)
	userHandler := handlers.NewUserHandler(userUsecase, logger)
	csHandler := handlers.NewCoffeeShopHandler(csUscase, logger)

	// Router
	appRouter := router.NewRouter(suite.cfg, userHandler, csHandler, authHandler, nil, authUsecase, logger)
	suite.Router = appRouter.SetupRouter()
}

// TearDownSuite tears down the test suite
func (suite *BaseTestSuite) TearDownSuite() {
	sqlDB, err := suite.DB.DB()
	if err != nil {
		suite.T().Fatalf("failed to get db instance: %v", err)
	}
	sqlDB.Close()
}

// TearDownTest cleans up the database after each test
func (suite *BaseTestSuite) TearDownTest() {
	// The order is important to avoid foreign key violations
	suite.DB.Exec("DELETE FROM user_refresh_tokens")
	suite.DB.Exec("DELETE FROM idea_like")
	suite.DB.Exec("DELETE FROM idea_comment")
	suite.DB.Exec("DELETE FROM idea")
	suite.DB.Exec("DELETE FROM worker_coffee_shop")
	suite.DB.Exec("DELETE FROM coffee_shop")
	suite.DB.Exec("DELETE FROM otps")
	suite.DB.Exec("DELETE FROM users")
}

// MakeRequest is a helper to make an HTTP request
func (suite *BaseTestSuite) MakeRequest(req TestRequest) *httptest.ResponseRecorder {
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
	if req.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.token)
	}

	suite.Router.ServeHTTP(w, httpReq)
	return w
}

// GetAuthResponse is a helper to get auth tokens for a test user
func (suite *BaseTestSuite) GetAuthResponse(phone, otpCode, name string) dto.AuthResponse {
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)

	var user models.User
	err := suite.DB.First(&user, "phone = ?", phone).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			_, err = suite.AuthRepo.CreateUser(&models.User{Name: name, Phone: phone})
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

	req := TestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/auth",
		body:        reqBody,
		contentType: "application/json",
	}
	w := suite.MakeRequest(req)
	suite.Require().Equal(http.StatusOK, w.Code, "Failed to get auth response. Body: %s", w.Body.String())

	var authResponse dto.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &authResponse)
	suite.Require().NoError(err)
	return authResponse
}

// GetAuthToken is a helper to get an auth token for a test user
func (suite *BaseTestSuite) GetAuthToken(phone, otpCode, name string) string {
	return suite.GetAuthResponse(phone, otpCode, name).AccessToken
}

// GetRandomAuthToken creates a user with random phone and returns an auth token
func (suite *BaseTestSuite) GetRandomAuthToken() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	phone := fmt.Sprintf("7%09d", r.Intn(1000000000))
	return suite.GetAuthToken(phone, "123456", "Test User")
}

// CreateUser is a helper to create a user
func (suite *BaseTestSuite) CreateUser(name, phone string) (*models.User, error) {
	user := &models.User{
		ID:    uuid.New(),
		Name:  name,
		Phone: phone,
	}
	err := suite.DB.Create(user).Error
	return user, err
}
