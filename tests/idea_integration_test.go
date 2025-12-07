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

type IdeaIntegrationTestSuite struct {
	BaseTestSuite
}

func (suite *IdeaIntegrationTestSuite) SetupSuite() {
	suite.BaseTestSuite.SetupSuite()
}

func (suite *IdeaIntegrationTestSuite) TearDownTest() {
	suite.BaseTestSuite.TearDownTest()
}

func TestIdeaIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IdeaIntegrationTestSuite))
}

// createIdeaPrerequisites is a helper to set up a user, coffee shop, and category for tests.
func (suite *IdeaIntegrationTestSuite) createIdeaPrerequisites() (string, *models.User, *models.CoffeeShop, *models.Category) {
	// Create user and get token
	user := suite.CreateUser("idea-creator", "111111111")
	token := suite.RegisterUserAndGetToken(user)

	// Create coffee shop
	coffeeShop := &models.CoffeeShop{
		Name:      "Test Coffee Shop for Ideas",
		Address:   "123 Idea St",
		CreatorID: user.ID,
	}
	err := suite.DB.Create(coffeeShop).Error
	suite.Require().NoError(err)

	// Create category
	category := &models.Category{
		Title:        "Idea Category",
		CoffeeShopID: &coffeeShop.ID,
	}
	err = suite.DB.Create(category).Error
	suite.Require().NoError(err)

	return token, user, coffeeShop, category
}

// createTestIdea is a helper to create a single idea for use in tests.
func (suite *IdeaIntegrationTestSuite) createTestIdea(author *models.User, cs *models.CoffeeShop, cat *models.Category) *models.Idea {
	var ideaStatus models.IdeaStatus
	suite.DB.FirstOrCreate(&ideaStatus, "title = ?", "new")

	idea := &models.Idea{
		Title:        "Test Idea",
		Description:  "A brilliant idea.",
		CreatorID:    &author.ID,
		CoffeeShopID: &cs.ID,
		CategoryID:   &cat.ID,
		StatusID:     &ideaStatus.ID,
	}
	err := suite.DB.Create(idea).Error
	suite.Require().NoError(err)
	return idea
}

func (suite *IdeaIntegrationTestSuite) TestCreateIdea() {
	token, _, coffeeShop, category := suite.createIdeaPrerequisites()

	tests := []struct {
		name           string
		token          string
		body           interface{}
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:  "Success - Create Idea",
			token: token,
			body: dto.CreateIdeaRequest{
				Title:        "My New Awesome Idea",
				Description:  "This is the description of my awesome idea.",
				CategoryID:   category.ID,
				CoffeeShopID: coffeeShop.ID,
			},
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
		},
		{
			name:           "Fail - Unauthorized",
			token:          "",
			body:           dto.CreateIdeaRequest{},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := TestRequest{
				method:      http.MethodPost,
				path:        "/api/v1/ideas",
				body:        tt.body,
				token:       tt.token,
				contentType: "application/json",
			}
			w := suite.MakeRequest(req)
			suite.Equal(tt.expectedStatus, w.Code)

			if tt.checkResponse {
				var resp dto.IdeaResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				suite.NoError(err)
				suite.Equal("My New Awesome Idea", resp.Title)
				suite.Equal(coffeeShop.ID, *resp.CoffeeShopID)
				suite.NotEqual(uuid.Nil, resp.ID)
			}
		})
	}
}

func (suite *IdeaIntegrationTestSuite) TestGetAllIdeas() {
	_, user, coffeeShop, category := suite.createIdeaPrerequisites()
	idea := suite.createTestIdea(user, coffeeShop, category)

	req := TestRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/v1/coffee-shops/%s/ideas", coffeeShop.ID),
	}
	w := suite.MakeRequest(req)

	suite.Equal(http.StatusOK, w.Code)

	var resp []dto.IdeaResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)

	found := false
	for _, i := range resp {
		if i.ID == idea.ID {
			found = true
			break
		}
	}
	suite.True(found, "created idea not found in list")
}

func (suite *IdeaIntegrationTestSuite) TestGetIdea() {
	_, user, coffeeShop, category := suite.createIdeaPrerequisites()
	idea := suite.createTestIdea(user, coffeeShop, category)

	req := TestRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/v1/ideas/%s", idea.ID),
	}
	w := suite.MakeRequest(req)

	suite.Equal(http.StatusOK, w.Code)

	var resp models.Idea
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(idea.ID, resp.ID)
	suite.Equal(idea.Title, resp.Title)
}

func (suite *IdeaIntegrationTestSuite) TestUpdateIdea() {
	authorToken, author, coffeeShop, category := suite.createIdeaPrerequisites()
	otherToken := suite.GetRandomAuthToken()

	admin := suite.CreateUser("admin-idea", "987654321")
	adminToken := suite.RegisterUserAndGetToken(admin)
	updTitle := "Updated Title"
	updateReq := dto.UpdateIdeaRequest{Title: &updTitle}

	suite.Run("Author can update their idea", func() {
		idea := suite.createTestIdea(author, coffeeShop, category)
		req := TestRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("/api/v1/ideas/%s", idea.ID),
			body:        updateReq,
			token:       authorToken,
			contentType: "application/json",
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusNoContent, w.Code)
		var updatedIdea models.Idea
		suite.DB.First(&updatedIdea, "id = ?", idea.ID)
		suite.Equal("Updated Title", updatedIdea.Title)
	})

	suite.Run("Other user cannot update idea", func() {
		idea := suite.createTestIdea(author, coffeeShop, category)
		req := TestRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("/api/v1/ideas/%s", idea.ID),
			body:        updateReq,
			token:       otherToken,
			contentType: "application/json",
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusUnauthorized, w.Code)
		var notUpdatedIdea models.Idea
		suite.DB.First(&notUpdatedIdea, "id = ?", idea.ID)
		suite.Equal("Test Idea", notUpdatedIdea.Title)
	})

	suite.Run("Admin cannot update idea", func() {
		idea := suite.createTestIdea(author, coffeeShop, category)
		req := TestRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("/api/v1/ideas/%s", idea.ID),
			body:        updateReq,
			token:       adminToken,
			contentType: "application/json",
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusUnauthorized, w.Code)
		var notUpdatedIdea models.Idea
		suite.DB.First(&notUpdatedIdea, "id = ?", idea.ID)
		suite.Equal("Test Idea", notUpdatedIdea.Title)
	})

	suite.Run("Unauthorized cannot update", func() {
		idea := suite.createTestIdea(author, coffeeShop, category)
		req := TestRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("/api/v1/ideas/%s", idea.ID),
			body:        updateReq,
			token:       "",
			contentType: "application/json",
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusUnauthorized, w.Code)
	})
}

func (suite *IdeaIntegrationTestSuite) TestDeleteIdea() {
	authorToken, author, coffeeShop, category := suite.createIdeaPrerequisites()
	otherToken := suite.GetRandomAuthToken()

	admin := suite.CreateUser("admin-idea-del", "123123123")
	adminToken := suite.RegisterUserAndGetToken(admin)

	// Make the admin a worker in the shop to test admin deletion privileges
	err := suite.DB.Create(&models.WorkerCoffeeShop{
		WorkerID:     &admin.ID,
		CoffeeShopID: &coffeeShop.ID,
		RoleID:       &suite.AdminRoleID,
	}).Error
	suite.Require().NoError(err)

	suite.Run("Author can delete their idea", func() {
		idea := suite.createTestIdea(author, coffeeShop, category)
		req := TestRequest{
			method: http.MethodDelete,
			path:   fmt.Sprintf("/api/v1/ideas/%s", idea.ID),
			token:  authorToken,
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusNoContent, w.Code)
		var count int64
		suite.DB.Model(&models.Idea{}).Where("id = ?", idea.ID).Count(&count)
		suite.Equal(int64(0), count)
	})

	suite.Run("Other user cannot delete idea", func() {
		idea := suite.createTestIdea(author, coffeeShop, category)
		req := TestRequest{
			method: http.MethodDelete,
			path:   fmt.Sprintf("/api/v1/ideas/%s", idea.ID),
			token:  otherToken,
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusForbidden, w.Code)
		var count int64
		suite.DB.Model(&models.Idea{}).Where("id = ?", idea.ID).Count(&count)
		suite.Equal(int64(1), count)
	})

	suite.Run("Admin can delete idea", func() {
		idea := suite.createTestIdea(author, coffeeShop, category)
		req := TestRequest{
			method: http.MethodDelete,
			path:   fmt.Sprintf("/api/v1/ideas/%s", idea.ID),
			token:  adminToken,
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusNoContent, w.Code)
		var count int64
		suite.DB.Model(&models.Idea{}).Where("id = ?", idea.ID).Count(&count)
		suite.Equal(int64(0), count)
	})

	suite.Run("Unauthorized cannot delete", func() {
		idea := suite.createTestIdea(author, coffeeShop, category)
		req := TestRequest{
			method: http.MethodDelete,
			path:   fmt.Sprintf("/api/v1/ideas/%s", idea.ID),
			token:  "",
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusUnauthorized, w.Code)
	})
}

