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

type WorkerCoffeeShopIntegrationTestSuite struct {
	BaseTestSuite
}

func (suite *WorkerCoffeeShopIntegrationTestSuite) SetupSuite() {
	suite.BaseTestSuite.SetupSuite()
}

func (suite *WorkerCoffeeShopIntegrationTestSuite) TearDownTest() {
	suite.BaseTestSuite.TearDownTest()
}

func TestWorkerCoffeeShopIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerCoffeeShopIntegrationTestSuite))
}

// createWorkerTestPrerequisites is a helper function to set up a standard test environment for worker management tests.
// It creates:
// 1. A Coffee Shop Creator (also a system admin for simplicity, owner of the shop).
// 2. An Admin User (system admin, but not the shop creator).
// 3. A regular User who will be a worker.
// 4. Another regular User for negative test cases (non-admin, non-worker).
// 5. A Coffee Shop created by the creator.
// It returns the created models and their auth tokens.
func (suite *WorkerCoffeeShopIntegrationTestSuite) createWorkerTestPrerequisites() (
	creator *models.User, creatorToken string,
	admin *models.User, adminToken string,
	worker *models.User, workerToken string,
	otherUser *models.User, otherUserToken string,
	shop *models.CoffeeShop) {

	// 1. Create Coffee Shop Creator (Owner)
	creator = suite.CreateUser("shop-creator", "1000000001")
	creatorToken = suite.RegisterUserAndGetToken(creator)

	// 2. Create another Admin User
	admin = suite.CreateUser("shop-admin", "1000000002")
	adminToken = suite.RegisterUserAndGetToken(admin)

	// 3. Create a regular user to be a worker
	worker = suite.CreateUser("worker-user", "1000000003")
	workerToken = suite.RegisterUserAndGetToken(worker)

	// 4. Create another regular user (non-admin)
	otherUser = suite.CreateUser("other-user", "1000000004")
	otherUserToken = suite.RegisterUserAndGetToken(otherUser)

	// 5. Create Coffee Shop
	shop = &models.CoffeeShop{
		Name:      "Test Shop for Workers",
		Address:   "123 Worker St",
		CreatorID: creator.ID,
	}
	err := suite.DB.Create(shop).Error
	suite.Require().NoError(err)

	// Make creator and admin workers in the shop with admin role
	suite.DB.Create(&models.WorkerCoffeeShop{WorkerID: &creator.ID, CoffeeShopID: &shop.ID, RoleID: &suite.AdminRoleID})
	suite.DB.Create(&models.WorkerCoffeeShop{WorkerID: &admin.ID, CoffeeShopID: &shop.ID, RoleID: &suite.AdminRoleID})

	return
}

func (suite *WorkerCoffeeShopIntegrationTestSuite) TestAddWorker() {
	_, creatorToken, _, adminToken, worker, _, otherUser, otherUserToken, shop := suite.createWorkerTestPrerequisites()

	tests := []struct {
		name           string
		token          string
		body           dto.AddWorkerToShopRequest
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "Fail - Regular user cannot add worker",
			token:          otherUserToken,
			body:           dto.AddWorkerToShopRequest{WorkerID: worker.ID, CoffeeShopID: shop.ID},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Success - Creator can add a worker",
			token:          creatorToken,
			body:           dto.AddWorkerToShopRequest{WorkerID: worker.ID, CoffeeShopID: shop.ID},
			expectedStatus: http.StatusCreated,
			checkResponse:  true,
		},
		{
			name:           "Fail - Cannot add an existing worker again",
			token:          creatorToken,
			body:           dto.AddWorkerToShopRequest{WorkerID: worker.ID, CoffeeShopID: shop.ID},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Success - Admin-worker can add another worker",
			token:          adminToken,
			body:           dto.AddWorkerToShopRequest{WorkerID: otherUser.ID, CoffeeShopID: shop.ID},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := TestRequest{
				method:      http.MethodPost,
				path:        "/api/v1/admin/worker-coffee-shops",
				body:        tt.body,
				token:       tt.token,
				contentType: "application/json",
			}

			w := suite.MakeRequest(req)
			suite.Equal(tt.expectedStatus, w.Code, "Response body: %s", w.Body.String())

			if tt.checkResponse {
				var resp dto.WorkerCoffeeShopResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				suite.NoError(err)
				suite.NotEqual(uuid.Nil, resp.ID)
				suite.Equal(tt.body.WorkerID, resp.Worker.ID)
				suite.Equal(tt.body.CoffeeShopID, resp.CoffeeShop.ID)
			}
		})
	}
}

func (suite *WorkerCoffeeShopIntegrationTestSuite) TestListAndRemoveWorker() {
	creator, creatorToken, admin, adminToken, worker, _, _, otherUserToken, shop := suite.createWorkerTestPrerequisites()

	// Add the worker to the shop to test listing and removal
	wcs := &models.WorkerCoffeeShop{WorkerID: &worker.ID, CoffeeShopID: &shop.ID, RoleID: &suite.UserRoleID}
	err := suite.DB.Create(wcs).Error
	suite.Require().NoError(err, "Failed to add worker to shop for testing")

	suite.Run("List Workers - Success by Creator with new worker", func() {
		req := TestRequest{
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/admin/coffee-shops/%s/workers", shop.ID),
			token:  creatorToken,
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusOK, w.Code)
		var resp []dto.UserResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		suite.NoError(err)
		suite.Len(resp, 3) // creator + admin + new worker
	})

	suite.Run("List Workers - Fail by non-admin", func() {
		req := TestRequest{
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/admin/coffee-shops/%s/workers", shop.ID),
			token:  otherUserToken,
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusForbidden, w.Code)
	})

	suite.Run("Remove Worker - Fail by non-admin", func() {
		req := TestRequest{
			method: http.MethodDelete,
			path:   fmt.Sprintf("/api/v1/admin/worker-coffee-shops/%s", wcs.ID),
			token:  otherUserToken,
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusForbidden, w.Code)
	})

	suite.Run("Remove Worker - Success by Admin-Worker", func() {
		req := TestRequest{
			method: http.MethodDelete,
			path:   fmt.Sprintf("/api/v1/admin/worker-coffee-shops/%s", wcs.ID),
			token:  adminToken,
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusNoContent, w.Code)

		// Verify soft delete
		var relation models.WorkerCoffeeShop
		err := suite.DB.Unscoped().First(&relation, "id = ?", wcs.ID).Error
		suite.NoError(err)
		suite.True(relation.IsDeleted, "Expected IsDeleted to be true for soft delete")
	})

	suite.Run("List Workers - Verify worker removed", func() {
		req := TestRequest{
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/admin/coffee-shops/%s/workers", shop.ID),
			token:  creatorToken,
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusOK, w.Code)
		var resp []dto.UserResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		suite.NoError(err)
		suite.Len(resp, 2) // Only creator + admin should remain

		// Verify that the correct users are present
		foundCreator := false
		foundAdmin := false
		for _, u := range resp {
			if u.ID == creator.ID {
				foundCreator = true
			}
			if u.ID == admin.ID {
				foundAdmin = true
			}
			// Fail if the removed worker is still present
			suite.NotEqual(worker.ID, u.ID, "Removed worker should not be in the list")
		}
		suite.True(foundCreator, "Creator should be in the list of workers")
		suite.True(foundAdmin, "Admin should be in the list of workers")
	})
}

func (suite *WorkerCoffeeShopIntegrationTestSuite) TestListShopsForWorker() {
	creator, _, _, _, worker, workerToken, otherUser, otherUserToken, shop1 := suite.createWorkerTestPrerequisites()

	// Create a second shop and add the worker to it as well
	shop2 := &models.CoffeeShop{Name: "Second Shop", Address: "456 Second St", CreatorID: creator.ID}
	suite.DB.Create(shop2)
	suite.DB.Create(&models.WorkerCoffeeShop{WorkerID: &worker.ID, CoffeeShopID: &shop1.ID})
	suite.DB.Create(&models.WorkerCoffeeShop{WorkerID: &worker.ID, CoffeeShopID: &shop2.ID})
	// Add the other user to only one shop
	suite.DB.Create(&models.WorkerCoffeeShop{WorkerID: &otherUser.ID, CoffeeShopID: &shop1.ID})

	suite.Run("Success - User lists their own shops", func() {
		req := TestRequest{
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/users/%s/coffee-shops", worker.ID),
			token:  workerToken,
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusOK, w.Code)
		var resp []dto.CoffeeShopResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		suite.NoError(err)
		suite.Len(resp, 2)
	})

	suite.Run("Fail - User tries to list another user's shops", func() {
		req := TestRequest{
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/users/%s/coffee-shops", worker.ID),
			token:  otherUserToken, // 'otherUser' tries to access 'worker's data
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusForbidden, w.Code)
	})

	suite.Run("Fail - Unauthorized user", func() {
		req := TestRequest{
			method: http.MethodGet,
			path:   fmt.Sprintf("/api/v1/users/%s/coffee-shops", worker.ID),
			token:  "", // No token
		}
		w := suite.MakeRequest(req)
		suite.Equal(http.StatusUnauthorized, w.Code)
	})
}
