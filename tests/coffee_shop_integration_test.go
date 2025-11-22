package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type CoffeeShopIntegrationTestSuite struct {
	BaseTestSuite
}

func (suite *CoffeeShopIntegrationTestSuite) SetupSuite() {
	suite.BaseTestSuite.SetupSuite()
	// You can add suite-specific setup here if needed
}

func (suite *CoffeeShopIntegrationTestSuite) TearDownTest() {
	suite.BaseTestSuite.TearDownTest()
	// You can add suite-specific teardown here if needed
}

func TestCoffeeShopIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CoffeeShopIntegrationTestSuite))
}

func (suite *CoffeeShopIntegrationTestSuite) TestCreateCoffeeShop_Authorized() {
	token := suite.GetRandomAuthToken()

	reqBody := dto.CreateCoffeeShopRequest{
		Name:    "Test Coffee Shop",
		Address: "123 Test St",
	}

	req := TestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/coffee-shops",
		body:        reqBody,
		contentType: "application/json",
		token:       token,
	}
	w := suite.MakeRequest(req)

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

	req := TestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/coffee-shops",
		body:        reqBody,
		contentType: "application/json",
	}
	w := suite.MakeRequest(req)

	suite.Equal(http.StatusUnauthorized, w.Code)
}

func (suite *CoffeeShopIntegrationTestSuite) TestGetAllCoffeeShops() {
	// Create a user first
	user := suite.CreateUser("test-user", "1234567890")

	// Create a coffee shop to be listed
	newShop := &models.CoffeeShop{ID: uuid.New(), Name: "Shop 1-" + uuid.New().String(), Address: "Addr 1", CreatorID: user.ID}
	err := suite.DB.Create(newShop).Error
	suite.Require().NoError(err)

	req := TestRequest{
		method: http.MethodGet,
		path:   "/api/v1/coffee-shops",
	}
	w := suite.MakeRequest(req)

	suite.Equal(http.StatusOK, w.Code)
	var resp []models.CoffeeShop
	err = json.Unmarshal(w.Body.Bytes(), &resp)
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
	user := suite.CreateUser("test-user", "1234567890")
	shopID := uuid.New()
	suite.DB.Create(&models.CoffeeShop{ID: shopID, Name: "Shop 1", Address: "Addr 1", CreatorID: user.ID})

	req := TestRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/v1/coffee-shops/%s", shopID),
	}
	w := suite.MakeRequest(req)

	suite.Equal(http.StatusOK, w.Code)
	var resp models.CoffeeShop
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(shopID, resp.ID)
	suite.Equal("Shop 1", resp.Name)
}

func (suite *CoffeeShopIntegrationTestSuite) TestUpdateCoffeeShop_Authorized() {
	token := suite.GetRandomAuthToken()
	reqBody := dto.CreateCoffeeShopRequest{
		Name:    "Test Coffee Shop",
		Address: "123 Test St",
	}

	createReq := TestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/coffee-shops",
		body:        reqBody,
		contentType: "application/json",
		token:       token,
	}
	w := suite.MakeRequest(createReq)

	var coffeeShop models.CoffeeShop
	err := json.Unmarshal(w.Body.Bytes(), &coffeeShop)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, w.Code)

	updateReq := dto.UpdateCoffeeShopRequest{
		Name: "Updated Shop Name",
	}

	req := TestRequest{
		method:      http.MethodPut,
		path:        fmt.Sprintf("/api/v1/coffee-shops/%s", coffeeShop.ID),
		body:        updateReq,
		contentType: "application/json",
		token:       token,
	}
	w = suite.MakeRequest(req)

	suite.Equal(http.StatusNoContent, w.Code)

	var updatedShop models.CoffeeShop
	suite.DB.First(&updatedShop, "id = ?", coffeeShop.ID)
	suite.Equal("Updated Shop Name", updatedShop.Name)
}

func (suite *CoffeeShopIntegrationTestSuite) TestUpdateCoffeeShop_Unauthorized() {
	token := suite.GetRandomAuthToken()
	reqBody := dto.CreateCoffeeShopRequest{
		Name:    "Test Coffee Shop",
		Address: "123 Test St",
	}

	createReq := TestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/coffee-shops",
		body:        reqBody,
		contentType: "application/json",
		token:       token,
	}
	w := suite.MakeRequest(createReq)

	var coffeeShop models.CoffeeShop
	err := json.Unmarshal(w.Body.Bytes(), &coffeeShop)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, w.Code)

	updateReq := dto.UpdateCoffeeShopRequest{
		Name: "Updated Shop Name",
	}

	req := TestRequest{
		method:      http.MethodPut,
		path:        fmt.Sprintf("/api/v1/coffee-shops/%s", coffeeShop.ID),
		body:        updateReq,
		contentType: "application/json",
	}
	w = suite.MakeRequest(req)

	suite.Equal(http.StatusUnauthorized, w.Code)

	var updatedShop models.CoffeeShop
	suite.DB.First(&updatedShop, "id = ?", coffeeShop.ID)
	suite.Equal("Test Coffee Shop", updatedShop.Name)
}

func (suite *CoffeeShopIntegrationTestSuite) TestDeleteCoffeeShop_Authorized() {
	token := suite.GetRandomAuthToken()
	reqBody := dto.CreateCoffeeShopRequest{
		Name:    "Test Coffee Shop",
		Address: "123 Test St",
	}

	createReq := TestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/coffee-shops",
		body:        reqBody,
		contentType: "application/json",
		token:       token,
	}
	w := suite.MakeRequest(createReq)

	var coffeeShop models.CoffeeShop
	err := json.Unmarshal(w.Body.Bytes(), &coffeeShop)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, w.Code)

	delReq := TestRequest{
		method: http.MethodDelete,
		path:   fmt.Sprintf("/api/v1/coffee-shops/%s", coffeeShop.ID),
		token:  token,
	}

	w = suite.MakeRequest(delReq)

	suite.Equal(http.StatusNoContent, w.Code)

	var count int64
	suite.DB.Model(&models.CoffeeShop{}).Where("id = ?", coffeeShop.ID).Count(&count)
	suite.Equal(int64(0), count)
}

func (suite *CoffeeShopIntegrationTestSuite) TestDeleteCoffeeShop_Unauthorized() {
	token := suite.GetRandomAuthToken()
	reqBody := dto.CreateCoffeeShopRequest{
		Name:    "Test Coffee Shop",
		Address: "123 Test St",
	}

	createReq := TestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/coffee-shops",
		body:        reqBody,
		contentType: "application/json",
		token:       token,
	}
	w := suite.MakeRequest(createReq)

	var coffeeShop models.CoffeeShop
	err := json.Unmarshal(w.Body.Bytes(), &coffeeShop)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, w.Code)

	delReq := TestRequest{
		method: http.MethodDelete,
		path:   fmt.Sprintf("/api/v1/coffee-shops/%s", coffeeShop.ID),
	}

	w = suite.MakeRequest(delReq)

	suite.Equal(http.StatusUnauthorized, w.Code)

	var count int64
	suite.DB.Model(&models.CoffeeShop{}).Where("id = ?", coffeeShop.ID).Count(&count)
	suite.Equal(int64(1), count)
}
