package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/stretchr/testify/suite"
)

type AdminRegistrationTestSuite struct {
	BaseTestSuite
}

func (suite *AdminRegistrationTestSuite) TestRegisterAdminAndCoffeeShop() {
	w := httptest.NewRecorder()

	testLogin := "admin_test_user_123"
	testPassword := "securepassword"
	testCoffeeShopName := "Test Coffee Shop"
	testAddress := "123 Test St"

	registerReq := dto.RegisterAdminRequest{
		Login:          testLogin,
		Password:       testPassword,
		CoffeeShopName: testCoffeeShopName,
		Address:        testAddress,
	}
	jsonValue, _ := json.Marshal(registerReq)

	req, _ := http.NewRequest("POST", "/api/v1/auth/register/admin", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	suite.Router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var authResp dto.AdminAuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &authResp)
	suite.NoError(err)
	suite.NotEmpty(authResp.AccessToken)
	suite.NotEmpty(authResp.RefreshToken)
	suite.NotEmpty(authResp.CoffeeShopID)

	// Verify that the user and coffee shop are created in the database
	var user models.User
	err = suite.DB.First(&user, "login = ?", testLogin).Error
	suite.NoError(err)
	suite.NotEmpty(user.ID)
	suite.NotNil(user.PasswordHash) // PasswordHash should be set

	var coffeeShop models.CoffeeShop
	err = suite.DB.First(&coffeeShop, "name = ?", testCoffeeShopName).Error
	suite.NoError(err)
	suite.NotEmpty(coffeeShop.ID)
	suite.Equal(user.ID, coffeeShop.CreatorID)
	suite.Equal(testAddress, coffeeShop.Address)

	// Verify worker_coffee_shop entry
	var workerCoffeeShop models.WorkerCoffeeShop
	err = suite.DB.Preload("Role").First(&workerCoffeeShop, "worker_id = ? AND coffee_shop_id = ?", user.ID, coffeeShop.ID).Error
	suite.NoError(err)
	suite.Equal("admin", workerCoffeeShop.Role.Name)
}

func (suite *AdminRegistrationTestSuite) TestRegisterAdminAndCoffeeShop_LoginConflict() {
	// First registration (should succeed)
	testLogin := "admin_test_user_conflict"
	testPassword := "securepassword"
	testCoffeeShopName := "Test Coffee Shop Conflict"
	testAddress := "456 Conflict Ave"

	registerReq1 := dto.RegisterAdminRequest{
		Login:          testLogin,
		Password:       testPassword,
		CoffeeShopName: testCoffeeShopName,
		Address:        testAddress,
	}
	jsonValue1, _ := json.Marshal(registerReq1)

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/api/v1/auth/register/admin", bytes.NewBuffer(jsonValue1))
	req1.Header.Set("Content-Type", "application/json")
	suite.Router.ServeHTTP(w1, req1)
	suite.Equal(http.StatusOK, w1.Code)

	// Second registration with the same login (should fail with conflict)
	registerReq2 := dto.RegisterAdminRequest{
		Login:          testLogin,
		Password:       testPassword,
		CoffeeShopName: "Another Coffee Shop",
		Address:        "789 Another Rd",
	}
	jsonValue2, _ := json.Marshal(registerReq2)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/v1/auth/register/admin", bytes.NewBuffer(jsonValue2))
	req2.Header.Set("Content-Type", "application/json")
	suite.Router.ServeHTTP(w2, req2)

	suite.Equal(http.StatusConflict, w2.Code)
	var errResp dto.ErrorResponse
	err := json.Unmarshal(w2.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Contains(errResp.Message, "user with this login already exists")
}

func (suite *AdminRegistrationTestSuite) TestRegisterAdminAndCoffeeShop_InvalidInput() {
	// Missing login
	registerReq := dto.RegisterAdminRequest{
		Password:       "securepassword",
		CoffeeShopName: "Test Coffee Shop",
		Address:        "123 Test St",
	}
	jsonValue, _ := json.Marshal(registerReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register/admin", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	suite.Router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	var errResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Contains(errResp.Message, "bad request")

	// Missing password
	registerReq = dto.RegisterAdminRequest{
		Login:          "test_admin_invalid",
		CoffeeShopName: "Test Coffee Shop",
		Address:        "123 Test St",
	}
	jsonValue, _ = json.Marshal(registerReq)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/register/admin", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	suite.Router.ServeHTTP(w, req)

	suite.Equal(http.StatusBadRequest, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &errResp)
	suite.NoError(err)
	suite.Contains(errResp.Message, "bad request")
}

func TestAdminRegistrationTestSuite(t *testing.T) {
	suite.Run(t, new(AdminRegistrationTestSuite))
}