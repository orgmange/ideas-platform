package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/stretchr/testify/suite"
)

type UserIdeasIntegrationTestSuite struct {
	BaseTestSuite
}

func (suite *UserIdeasIntegrationTestSuite) SetupSuite() {
	suite.BaseTestSuite.SetupSuite()
}

func (suite *UserIdeasIntegrationTestSuite) TearDownTest() {
	suite.BaseTestSuite.TearDownTest()
}

func TestUserIdeasIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserIdeasIntegrationTestSuite))
}

func (suite *UserIdeasIntegrationTestSuite) createIdeaPrerequisites() (string, *models.User, *models.CoffeeShop, *models.Category) {
	// Create user and get token
	user := suite.CreateUser("idea-user-test", "999888777")
	token := suite.RegisterUserAndGetToken(user)

	// Create coffee shop
	coffeeShop := &models.CoffeeShop{
		Name:      "User Ideas Shop",
		Address:   "456 User Ideas St",
		CreatorID: user.ID,
	}
	err := suite.DB.Create(coffeeShop).Error
	suite.Require().NoError(err)

	// Create category
	category := &models.Category{
		Title:        "User Ideas Category",
		CoffeeShopID: &coffeeShop.ID,
	}
	err = suite.DB.Create(category).Error
	suite.Require().NoError(err)

	return token, user, coffeeShop, category
}

func (suite *UserIdeasIntegrationTestSuite) createTestIdea(author *models.User, cs *models.CoffeeShop, cat *models.Category, title string) *models.Idea {
	var ideaStatus models.IdeaStatus
	suite.DB.FirstOrCreate(&ideaStatus, "title = ?", "new")

	idea := &models.Idea{
		Title:        title,
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

func (suite *UserIdeasIntegrationTestSuite) TestGetIdeasFromUser() {
	token, user, coffeeShop, category := suite.createIdeaPrerequisites()
	idea1 := suite.createTestIdea(user, coffeeShop, category, "Idea 1")
	idea2 := suite.createTestIdea(user, coffeeShop, category, "Idea 2")

	// Test with empty parameters (Regression test for the fix)
	suite.Run("Get ideas with empty parameters", func() {
		req := TestRequest{
			method: http.MethodGet,
			path:   "/api/v1/users/me/ideas",
			token:  token,
		}
		w := suite.MakeRequest(req)

		suite.Equal(http.StatusOK, w.Code)

		var resp []dto.IdeaResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		suite.NoError(err)
		suite.GreaterOrEqual(len(resp), 2)
		
		ids := []string{resp[0].ID.String(), resp[1].ID.String()}
		suite.Contains(ids, idea1.ID.String())
		suite.Contains(ids, idea2.ID.String())
	})

	// Test with pagination
	suite.Run("Get ideas with pagination", func() {
		req := TestRequest{
			method: http.MethodGet,
			path:   "/api/v1/users/me/ideas?limit=1&page=1",
			token:  token,
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusOK, w.Code)
		var resp []dto.IdeaResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		suite.NoError(err)
		suite.Len(resp, 1)
	})
	
	// Test with sort
	suite.Run("Get ideas with sort", func() {
		req := TestRequest{
			method: http.MethodGet,
			path:   "/api/v1/users/me/ideas?sort=created_at",
			token:  token,
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusOK, w.Code)
	})
}
