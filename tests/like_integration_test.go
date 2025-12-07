package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/stretchr/testify/suite"
)

type LikeIntegrationTestSuite struct {
	BaseTestSuite
}

func (suite *LikeIntegrationTestSuite) SetupSuite() {
	suite.BaseTestSuite.SetupSuite()
}

func (suite *LikeIntegrationTestSuite) TearDownTest() {
	suite.BaseTestSuite.TearDownTest()
}

func TestLikeIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(LikeIntegrationTestSuite))
}

func (suite *LikeIntegrationTestSuite) createLikePrerequisites() (string, *models.User, *models.Idea) {
	// Create user and get token
	user := suite.CreateUser("like-user", "222222222")
	token := suite.RegisterUserAndGetToken(user)

	// Create coffee shop
	coffeeShop := &models.CoffeeShop{
		Name:      "Test Coffee Shop for Likes",
		Address:   "123 Like St",
		CreatorID: user.ID,
	}
	err := suite.DB.Create(coffeeShop).Error
	suite.Require().NoError(err)

	// Create category
	category := &models.Category{
		Title:        "Like Category",
		CoffeeShopID: &coffeeShop.ID,
	}
	err = suite.DB.Create(category).Error
	suite.Require().NoError(err)

	// Create an idea
	var ideaStatus models.IdeaStatus
	suite.DB.FirstOrCreate(&ideaStatus, "title = ?", "new")
	idea := &models.Idea{
		Title:        "Test Idea for Likes",
		Description:  "A brilliant idea to be liked.",
		CreatorID:    &user.ID,
		CoffeeShopID: &coffeeShop.ID,
		CategoryID:   &category.ID,
		StatusID:     &ideaStatus.ID,
	}
	err = suite.DB.Create(idea).Error
	suite.Require().NoError(err)

	return token, user, idea
}

func (suite *LikeIntegrationTestSuite) TestLikeIdea() {
	token, _, idea := suite.createLikePrerequisites()

	suite.Run("Success - Like Idea", func() {
		req := TestRequest{
			method: http.MethodPost,
			path:   fmt.Sprintf("/api/v1/ideas/%s/like", idea.ID),
			token:  token,
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusCreated, w.Code)

		// Verify the like in the database
		var like models.IdeaLike
		err := suite.DB.Where("idea_id = ?", idea.ID).First(&like).Error
		suite.NoError(err)
	})

	suite.Run("Fail - Unauthorized", func() {
		req := TestRequest{
			method: http.MethodPost,
			path:   fmt.Sprintf("/api/v1/ideas/%s/like", idea.ID),
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusUnauthorized, w.Code)
	})
}

func (suite *LikeIntegrationTestSuite) TestUnlikeIdea() {
	token, user, idea := suite.createLikePrerequisites()

	// Like the idea first
	err := suite.DB.Create(&models.IdeaLike{UserID: &user.ID, IdeaID: &idea.ID}).Error
	suite.Require().NoError(err)

	suite.Run("Success - Unlike Idea", func() {
		req := TestRequest{
			method: http.MethodDelete,
			path:   fmt.Sprintf("/api/v1/ideas/%s/unlike", idea.ID),
			token:  token,
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusNoContent, w.Code)

		// Verify the like is removed from the database
		var count int64
		suite.DB.Model(&models.IdeaLike{}).Where("idea_id = ?", idea.ID).Count(&count)
		suite.Equal(int64(0), count)
	})
}

func (suite *LikeIntegrationTestSuite) TestGetIdeaWithLikes() {
	token, user, idea := suite.createLikePrerequisites()

	// Like the idea
	err := suite.DB.Create(&models.IdeaLike{UserID: &user.ID, IdeaID: &idea.ID}).Error
	suite.Require().NoError(err)

	req := TestRequest{
		method: http.MethodGet,
		path:   fmt.Sprintf("/api/v1/ideas/%s", idea.ID),
		token:  token,
	}
	w := suite.MakeRequest(req)
	suite.Equal(http.StatusOK, w.Code)

	var resp dto.IdeaResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	suite.NoError(err)
	suite.Equal(1, resp.Likes)
}
