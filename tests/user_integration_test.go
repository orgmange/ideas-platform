package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/suite"
)

type RouterTestSuite struct {
	BaseTestSuite
}

func (suite *RouterTestSuite) SetupSuite() {
	suite.BaseTestSuite.SetupSuite()
	// You can add suite-specific setup here if needed
}

func (suite *RouterTestSuite) TearDownTest() {
	suite.BaseTestSuite.TearDownTest()
	// You can add suite-specific teardown here if needed
}

func TestRouterTestSuite(t *testing.T) {
	suite.Run(t, new(RouterTestSuite))
}

func (suite *RouterTestSuite) TestGetAllUsers() {
	    admin := suite.CreateUser("admin", "33333")
		adminToken := suite.GetAuthToken(admin.Phone, "333", admin.Name)
		// Create a coffee shop and make the user an admin
		coffeeShop := &models.CoffeeShop{Name: "Admin's Test Shop", CreatorID: admin.ID, Address: "123 Admin Lane"}
		suite.DB.Create(coffeeShop)
		suite.DB.Create(&models.WorkerCoffeeShop{WorkerID: &admin.ID, CoffeeShopID: &coffeeShop.ID, RoleID: &suite.AdminRoleID})
	
	userToken := suite.GetRandomAuthToken()

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "admin gets all users",
			token:          adminToken,
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "user fails to get all users",
			token:          userToken,
			expectedStatus: http.StatusForbidden,
			checkResponse:  false,
		},
		{
			name:           "unauthorized get fails",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  false,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			req := TestRequest{
				method: "GET",
				path:   "/api/v1/users",
				token:  tc.token,
			}
			w := suite.MakeRequest(req)
			suite.Equal(tc.expectedStatus, w.Code)

			if tc.checkResponse {
				var resp []dto.UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				suite.NoError(err)
				// 2 users created in setup + admin + random user
				suite.GreaterOrEqual(len(resp), 2)
			}
		})
	}
}

func (suite *RouterTestSuite) TestGetUser() {
	targetUser := suite.CreateUser("target", "11111")
	targetToken := suite.GetAuthToken(targetUser.Phone, "111", targetUser.Name)

	otherUser := suite.CreateUser("other", "22222")
	otherToken := suite.GetAuthToken(otherUser.Phone, "222", otherUser.Name)

	admin := suite.CreateUser("admin", "33333")
	adminToken := suite.GetAuthToken(admin.Phone, "333", admin.Name)
	// Create a coffee shop and make the user an admin
	coffeeShop := &models.CoffeeShop{Name: "Admin's Test Shop", CreatorID: admin.ID, Address: "123 Admin Lane"}
	suite.DB.Create(coffeeShop)
	suite.DB.Create(&models.WorkerCoffeeShop{WorkerID: &admin.ID, CoffeeShopID: &coffeeShop.ID, RoleID: &suite.AdminRoleID})


	tests := []struct {
		name           string
		token          string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "user gets own profile",
			token:          targetToken,
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "user fails to get other's profile",
			token:          otherToken,
			expectedStatus: http.StatusForbidden,
			checkResponse:  false,
		},
		{
			name:           "admin gets other's profile",
			token:          adminToken,
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "unauthorized get fails",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  false,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			req := TestRequest{
				method: http.MethodGet,
				path:   fmt.Sprintf("/api/v1/users/%s", targetUser.ID.String()),
				token:  tc.token,
			}
			w := suite.MakeRequest(req)
			suite.Equal(tc.expectedStatus, w.Code)

			if tc.checkResponse {
				var resp dto.UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				suite.NoError(err)
				suite.Equal(targetUser.ID, resp.ID)
				suite.Equal(targetUser.Name, resp.Name)
			}
		})
	}
}

func (suite *RouterTestSuite) TestUpdateUser() {
	userToUpdate := suite.CreateUser("testuser", "12345")
	userToken := suite.GetAuthToken(userToUpdate.Phone, "123456", userToUpdate.Name)

	otherUser := suite.CreateUser("otheruser", "54321")
	otherToken := suite.GetAuthToken(otherUser.Phone, "123456", otherUser.Name)

	updateReq := dto.UpdateUserRequest{Name: "updated-name"}

	tests := []struct {
		name           string
		userID         string
		token          string
		expectedStatus int
	}{
		{
			name:           "user updates own profile",
			userID:         userToUpdate.ID.String(),
			token:          userToken,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "user fails to update other's profile",
			userID:         userToUpdate.ID.String(),
			token:          otherToken,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "unauthorized update fails",
			userID:         userToUpdate.ID.String(),
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			req := TestRequest{
				method:      http.MethodPut,
				path:        fmt.Sprintf("/api/v1/users/%s", tc.userID),
				body:        updateReq,
				contentType: "application/json",
				token:       tc.token,
			}
			w := suite.MakeRequest(req)
			suite.Equal(tc.expectedStatus, w.Code)
		})
	}

	// Verify the name was actually updated
	var finalUser models.User
	suite.DB.First(&finalUser, "id = ?", userToUpdate.ID)
	suite.Equal("updated-name", finalUser.Name)
}

func (suite *RouterTestSuite) TestDeleteUser() {
	userToDelete := suite.CreateUser("todelete", "111111")
	userToken := suite.GetAuthToken(userToDelete.Phone, "111", userToDelete.Name)

	otherUser := suite.CreateUser("other", "22222")
	otherToken := suite.GetAuthToken(otherUser.Phone, "222", otherUser.Name)

	tests := []struct {
		name           string
		userID         string
		token          string
		expectedStatus int
	}{
		{
			name:           "user fails to delete other's profile",
			userID:         userToDelete.ID.String(),
			token:          otherToken,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "unauthorized delete fails",
			userID:         userToDelete.ID.String(),
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "user deletes own profile",
			userID:         userToDelete.ID.String(),
			token:          userToken,
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			// Re-create user if it was deleted in a previous sub-test
			if tc.name == "user deletes own profile" {
				suite.DB.Model(&models.User{}).Where("id = ?", userToDelete.ID).Update("is_deleted", false)
			}

			req := TestRequest{
				method: http.MethodDelete,
				path:   fmt.Sprintf("/api/v1/users/%s", tc.userID),
				token:  tc.token,
			}
			w := suite.MakeRequest(req)
			suite.Equal(tc.expectedStatus, w.Code)
		})
	}

	// Verify the user was actually deleted
	var deletedUser models.User
	err := suite.DB.First(&deletedUser, "id = ?", userToDelete.ID).Error
	suite.NoError(err) // GORM soft-delete means the record is still there
	suite.True(deletedUser.IsDeleted)
}
