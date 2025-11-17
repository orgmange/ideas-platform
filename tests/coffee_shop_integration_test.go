package tests

import (
	"bytes"
	"database/sql"
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
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CoffeeShopIntegrationTestSuite struct {
	suite.Suite
	DB       *gorm.DB
	cfg      *config.Config
	Router   *gin.Engine
	authRepo repository.AuthRepository
}

func (suite *CoffeeShopIntegrationTestSuite) SetupSuite() {
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	var sqlDB *sql.DB
	sqlDB, err = suite.DB.DB()
	if err != nil {
		suite.T().Fatalf("failed to get sql.DB: %v", err)
	}

	userRepository := repository.NewUserRepository(sqlDB)
	userUsecase := usecase.NewUserUsecase(userRepository)
	userHandler := handlers.NewUserHandler(userUsecase, logger)

	coffeeShopRepo := repository.NewCoffeeShopRepository(suite.DB)
	csUscase := usecase.NewCoffeeShopUsecase(coffeeShopRepo)
	csHandler := handlers.NewCoffeeShopHandler(csUscase, logger)

	authUsecase := usecase.NewAuthUsecase(suite.authRepo, "test-secret")
	authHandler := handlers.NewAuthHandler(authUsecase, logger)

	appRouter := router.NewRouter(suite.cfg, userHandler, csHandler, authHandler, authUsecase, logger)
	suite.Router = appRouter.SetupRouter()
}

func (suite *CoffeeShopIntegrationTestSuite) TearDownSuite() {
	sqlDB, err := suite.DB.DB()
	if err != nil {
		suite.T().Fatalf("failed to get db instance: %v", err)
	}
	sqlDB.Close()
}

func (suite *CoffeeShopIntegrationTestSuite) TearDownTest() {
	suite.DB.Exec("DELETE FROM coffee_shops")
	suite.DB.Exec("DELETE FROM otps")
	suite.DB.Exec("DELETE FROM users")
	suite.DB.Exec("DELETE FROM user_refresh_tokens")
}

func TestCoffeeShopIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CoffeeShopIntegrationTestSuite))
}

type coffeeShopTestRequest struct {
	method      string
	path        string
	body        interface{}
	contentType string
	token       string
}

func (suite *CoffeeShopIntegrationTestSuite) makeRequest(req coffeeShopTestRequest) *httptest.ResponseRecorder {
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

func (suite *CoffeeShopIntegrationTestSuite) getAuthToken() string {
	phone := "1234567890"
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

	req := coffeeShopTestRequest{
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
	return authResponse.AccessToken
}

func (suite *CoffeeShopIntegrationTestSuite) TestCreateCoffeeShop_Authorized() {
	token := suite.getAuthToken()

	reqBody := dto.CreateCoffeeShopRequest{
		Name:    "Test Coffee Shop",
		Address: "123 Test St",
	}

	req := coffeeShopTestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/coffee-shops",
		body:        reqBody,
		contentType: "application/json",
		token:       token,
	}
	w := suite.makeRequest(req)

	suite.Equal(http.StatusCreated, w.Code)

	var coffeeShop models.CoffeeShop
	err := json.Unmarshal(w.Body.Bytes(), &coffeeShop)
	suite.NoError(err)
	suite.Equal("Test Coffee Shop", coffeeShop.Name)
	suite.Equal("123 Test St", coffeeShop.Address)
}

func (suite *CoffeeShopIntegrationTestSuite) TestCreateCoffeeShop_Unauthorized() {
	reqBody := dto.CreateCoffeeShopRequest{
		Name:    "Test Coffee Shop",
		Address: "123 Test St",
	}

	req := coffeeShopTestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/coffee-shops",
		body:        reqBody,
		contentType: "application/json",
	}
	w := suite.makeRequest(req)

	suite.Equal(http.StatusUnauthorized, w.Code)
}

func (suite *CoffeeShopIntegrationTestSuite) TestGetAllCoffeeShops() {
	// Create a coffee shop to be listed
	newShop := &models.CoffeeShop{ID: uuid.New(), Name: "Shop 1-" + uuid.New().String(), Address: "Addr 1"}
	suite.DB.Create(newShop)

	req := coffeeShopTestRequest{
		method: http.MethodGet,
		path:   "/api/v1/coffee-shops",
	}
	w := suite.makeRequest(req)

	suite.Equal(http.StatusOK, w.Code)
	var resp []models.CoffeeShop
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)

	// Find the created shop in the response
	found := false
	for _, shop := range resp {
		if shop.ID == newShop.ID && shop.Name == newShop.Name {
			found = true
			break
		}
	}
	suite.True(found, "created coffee shop not found in list")
}

func (suite *CoffeeShopIntegrationTestSuite) TestGetCoffeeShop() {
	shopID := uuid.New()
	suite.DB.Create(&models.CoffeeShop{ID: shopID, Name: "Shop 1", Address: "Addr 1"})

	req := coffeeShopTestRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/v1/coffee-shops/%s", shopID),
	}
	w := suite.makeRequest(req)

	suite.Equal(http.StatusOK, w.Code)
	var resp models.CoffeeShop
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(shopID, resp.ID)
	suite.Equal("Shop 1", resp.Name)
}

func (suite *CoffeeShopIntegrationTestSuite) TestUpdateCoffeeShop() {
	token := suite.getAuthToken()
	shopID := uuid.New()
	suite.DB.Create(&models.CoffeeShop{ID: shopID, Name: "Shop 1", Address: "Addr 1"})

	updateReq := dto.UpdateCoffeeShopRequest{
		Name: "Updated Shop Name",
	}

	req := coffeeShopTestRequest{
		method:      http.MethodPut,
		path:        fmt.Sprintf("/api/v1/coffee-shops/%s", shopID),
		body:        updateReq,
		contentType: "application/json",
		token:       token,
	}
	w := suite.makeRequest(req)

	suite.Equal(http.StatusNoContent, w.Code)

	var updatedShop models.CoffeeShop
	suite.DB.First(&updatedShop, "id = ?", shopID)
	suite.Equal("Updated Shop Name", updatedShop.Name)
}

func (suite *CoffeeShopIntegrationTestSuite) TestDeleteCoffeeShop() {
	token := suite.getAuthToken()
	shopID := uuid.New()
	suite.DB.Create(&models.CoffeeShop{ID: shopID, Name: "Shop 1", Address: "Addr 1"})

	req := coffeeShopTestRequest{
		method: http.MethodDelete,
		path:   fmt.Sprintf("/api/v1/coffee-shops/%s", shopID),
		token:  token,
	}
	w := suite.makeRequest(req)

	suite.Equal(http.StatusNoContent, w.Code)

	var count int64
	suite.DB.Model(&models.CoffeeShop{}).Where("id = ?", shopID).Count(&count)
	suite.Equal(int64(0), count)
}