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

type RewardIntegrationTestSuite struct {
	BaseTestSuite
}

func (suite *RewardIntegrationTestSuite) SetupSuite() {
	suite.BaseTestSuite.SetupSuite()
}

func (suite *RewardIntegrationTestSuite) TearDownTest() {
	suite.BaseTestSuite.TearDownTest()
}

func TestRewardIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(RewardIntegrationTestSuite))
}

// Helper to create a standard test setup: an admin, a regular user, a coffee shop, an idea, and a reward type
func (suite *RewardIntegrationTestSuite) createAdminTestPrerequisites() (adminToken string, coffeeShop models.CoffeeShop, idea models.Idea, rewardType models.RewardType, regularUserToken string) {
	// 1. Create Admin User
	admin := suite.CreateUser("test-admin", "9999999999")
	admin.RoleID = suite.AdminRoleID
	err := suite.DB.Save(admin).Error
	suite.Require().NoError(err)
	adminToken = suite.RegisterUserAndGetToken(admin)

	// 2. Create Coffee Shop (owned by admin for simplicity)
	coffeeShop = models.CoffeeShop{
		Name:      "Admin's Coffee Shop",
		Address:   "1 Admin St",
		CreatorID: admin.ID,
	}
	err = suite.DB.Create(&coffeeShop).Error
	suite.Require().NoError(err)

	// 3. Create Regular User (Idea Author)
	author := suite.CreateUser("idea-author", "2222222222")
	regularUserToken = suite.RegisterUserAndGetToken(author)

	// 4. Create Category for Idea
	category := models.Category{Title: "Test Category", CoffeeShopID: &coffeeShop.ID}
	err = suite.DB.Create(&category).Error
	suite.Require().NoError(err)

	// 5. Create Idea Status
	ideaStatus := models.IdeaStatus{Title: "new"}
	suite.DB.FirstOrCreate(&ideaStatus, "title = ?", "new")

	// 6. Create Idea
	idea = models.Idea{
		CreatorID:    &author.ID,
		CoffeeShopID: &coffeeShop.ID,
		CategoryID:   &category.ID,
		StatusID:     &ideaStatus.ID,
		Title:        "A Great Idea by a User",
		Description:  "This idea is truly great.",
	}
	err = suite.DB.Create(&idea).Error
	suite.Require().NoError(err)

	// 7. Create Reward Type
	rewardType = models.RewardType{
		CoffeeShopID: &coffeeShop.ID,
		Description:  "Free Admin-Approved Coffee",
	}
	err = suite.DB.Create(&rewardType).Error
	suite.Require().NoError(err)

	return adminToken, coffeeShop, idea, rewardType, regularUserToken
}

func (suite *RewardIntegrationTestSuite) TestGiveReward_Admin() {
	adminToken, _, idea, rewardType, regularUserToken := suite.createAdminTestPrerequisites()
	var createdRewardID uuid.UUID

	tests := []struct {
		name           string
		token          string
		body           interface{}
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "Fail - No Token",
			token:          "",
			body:           dto.GiveRewardRequest{IdeaID: idea.ID, RewardTypeID: rewardType.ID},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Fail - Regular User cannot give reward",
			token:          regularUserToken,
			body:           dto.GiveRewardRequest{IdeaID: idea.ID, RewardTypeID: rewardType.ID},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Success - Admin gives a reward",
			token:          adminToken,
			body:           dto.GiveRewardRequest{IdeaID: idea.ID, RewardTypeID: rewardType.ID},
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := TestRequest{
				method:      http.MethodPost,
				path:        "/api/v1/admin/rewards",
				body:        tt.body,
				token:       tt.token,
				contentType: "application/json",
			}
			w := suite.MakeRequest(req)
			suite.Equal(tt.expectedStatus, w.Code)
			if tt.checkResponse {
				var resp dto.RewardResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				suite.NoError(err)
				suite.Equal(idea.ID, *resp.IdeaID)
				suite.Equal(*idea.CreatorID, *resp.ReceiverID)
				createdRewardID = resp.ID
			}
		})
	}
	// Final check: ensure the created reward can be fetched via the public GET endpoint
	suite.NotEqual(uuid.Nil, createdRewardID)
	req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/rewards/%s", createdRewardID)}
	w := suite.MakeRequest(req)
	suite.Equal(http.StatusOK, w.Code)
}

func (suite *RewardIntegrationTestSuite) TestRevokeReward_Admin() {
	adminToken, _, idea, rewardType, regularUserToken := suite.createAdminTestPrerequisites()

	// Admin gives a reward first
	createReq := TestRequest{
		method: http.MethodPost, path: "/api/v1/admin/rewards", token: adminToken, contentType: "application/json",
		body: dto.GiveRewardRequest{IdeaID: idea.ID, RewardTypeID: rewardType.ID},
	}
	w := suite.MakeRequest(createReq)
	suite.Require().Equal(http.StatusCreated, w.Code)
	var createdReward dto.RewardResponse
	err := json.Unmarshal(w.Body.Bytes(), &createdReward)
	suite.Require().NoError(err)

	tests := []struct {
		name           string
		path           string
		token          string
		expectedStatus int
	}{
		{
			name:           "Fail - Regular User cannot revoke reward",
			path:           fmt.Sprintf("/api/v1/admin/rewards/%s", createdReward.ID),
			token:          regularUserToken,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Fail - No token cannot revoke",
			path:           fmt.Sprintf("/api/v1/admin/rewards/%s", createdReward.ID),
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Success - Admin revokes reward",
			path:           fmt.Sprintf("/api/v1/admin/rewards/%s", createdReward.ID),
			token:          adminToken,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Fail - Admin revokes already revoked reward",
			path:           fmt.Sprintf("/api/v1/admin/rewards/%s", createdReward.ID),
			token:          adminToken,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := TestRequest{method: http.MethodDelete, path: tt.path, token: tt.token}
			w := suite.MakeRequest(req)
			suite.Equal(tt.expectedStatus, w.Code)
		})
	}
}

func (suite *RewardIntegrationTestSuite) TestListRewards_Public() {
	adminToken, coffeeShop, idea, rewardType, authorToken := suite.createAdminTestPrerequisites()

	// Admin gives a reward to the author
	suite.MakeRequest(TestRequest{
		method: http.MethodPost, path: "/api/v1/admin/rewards", token: adminToken, contentType: "application/json",
		body: dto.GiveRewardRequest{IdeaID: idea.ID, RewardTypeID: rewardType.ID},
	})

	// --- GetMyRewards (Authenticated User) ---
	suite.Run("Author lists their own rewards", func() {
		req := TestRequest{method: http.MethodGet, path: "/api/v1/users/me/rewards", token: authorToken}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusOK, w.Code)
		var resp []dto.RewardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		suite.NoError(err)
		suite.Len(resp, 1)
		suite.Equal(*idea.CreatorID, *resp[0].ReceiverID)
	})

	// --- GetRewardsForCoffeeShop (Authenticated User) ---
	suite.Run("Admin lists rewards for a shop", func() {
		req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/coffee-shops/%s/rewards", coffeeShop.ID), token: adminToken}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusOK, w.Code)
		var resp []dto.RewardResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		suite.NoError(err)
		suite.Len(resp, 1)
	})
}
