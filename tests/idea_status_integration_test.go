package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type IdeaStatusTestSuite struct {
	BaseTestSuite
	AdminToken string
	AdminUser  *models.User
}

func (suite *IdeaStatusTestSuite) SetupTest() {
	suite.BaseTestSuite.TearDownTest()
	// Create admin user for tests
	name := "Admin"
	phone := "1234567890"
	suite.AdminUser = suite.CreateUser(name, phone)
	
	// Assign admin role in a system shop (needed for admin access)
	// Or use existing helper to create a shop where they are admin
	// The middleware AdminFilter checks IsAdminInAnyShop
	// So we need to create a shop and make them admin
	shop := &models.CoffeeShop{
		ID:        uuid.New(),
		CreatorID: suite.AdminUser.ID,
		Name:      "Admin Shop",
		Address:   "Admin Address",
	}
	suite.DB.Create(shop)
	
	suite.CreateWorkerForShop(suite.AdminUser, shop, suite.AdminRoleID)
	
	suite.AdminToken = suite.GetAuthToken(phone, "123456", name)
	
	// Create "Создана" status for default assignment tests if it doesn't exist
	createdStatus := &models.IdeaStatus{
		Title: "Создана",
	}
	suite.DB.Create(createdStatus)
}

func (suite *IdeaStatusTestSuite) TestCreateUpdateDeleteStatus() {
	// 1. Create Status
	statusTitle := fmt.Sprintf("NewStatus_%d", time.Now().UnixNano())
	createReq := dto.CreateIdeaStatusRequest{
		Title: statusTitle,
	}
	
	req := TestRequest{
		method: http.MethodPost,
		path:   "/api/v1/admin/statuses",
		body:   createReq,
		token:  suite.AdminToken,
	}
	w := suite.MakeRequest(req)
	suite.Require().Equal(http.StatusCreated, w.Code)
	
	var createResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	suite.Require().NoError(err)
	statusIDStr := createResp["id"].(string)
	suite.Require().NotEmpty(statusIDStr)

	// 2. Get Status
	req = TestRequest{
		method: http.MethodGet,
		path:   "/api/v1/statuses/" + statusIDStr,
		token:  "",
	}
	w = suite.MakeRequest(req)
	suite.Require().Equal(http.StatusOK, w.Code)
	
	var getResp dto.IdeaStatusResponse
	err = json.Unmarshal(w.Body.Bytes(), &getResp)
	suite.Require().NoError(err)
	suite.Equal(statusTitle, getResp.Title)
	suite.Equal(statusIDStr, getResp.ID.String())

	// 3. Update Status
	updatedTitle := statusTitle + "_updated"
	updateReq := dto.UpdateIdeaStatusRequest{
		Title: updatedTitle,
	}
	
	req = TestRequest{
		method: http.MethodPut,
		path:   "/api/v1/admin/statuses/" + statusIDStr,
		body:   updateReq,
		token:  suite.AdminToken,
	}
	w = suite.MakeRequest(req)
	suite.Require().Equal(http.StatusOK, w.Code)

	// Verify Update
	req = TestRequest{
		method: http.MethodGet,
		path:   "/api/v1/statuses/" + statusIDStr,
		token:  "",
	}
	w = suite.MakeRequest(req)
	suite.Require().Equal(http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &getResp)
	suite.Require().NoError(err)
	suite.Equal(updatedTitle, getResp.Title)

	// 4. Delete Status
	req = TestRequest{
		method: http.MethodDelete,
		path:   "/api/v1/admin/statuses/" + statusIDStr,
		token:  suite.AdminToken,
	}
	w = suite.MakeRequest(req)
	suite.Require().Equal(http.StatusOK, w.Code)

	// Verify Deletion
	req = TestRequest{
		method: http.MethodGet,
		path:   "/api/v1/statuses/" + statusIDStr,
		token:  "",
	}
	w = suite.MakeRequest(req)
	suite.Require().Equal(http.StatusNotFound, w.Code)
}

func (suite *IdeaStatusTestSuite) TestDefaultStatusAssignment() {
	// User creates an idea
	userName := "User"
	userPhone := "9876543210"
	_, shop := suite.CreateTestUser(userName, userPhone, "User Shop", "Address", suite.UserRoleID)
	userToken := suite.GetAuthToken(userPhone, "123456", userName)
	
	desc := "Desc"
	cat := &models.Category{
		CoffeeShopID: &shop.ID,
		Title:        "Test Category",
		Description:  &desc,
	}
	suite.DB.Create(cat)

	formData := map[string]string{
		"coffee_shop_id": shop.ID.String(),
		"category_id":    cat.ID.String(),
		"title":          "Idea with Default Status",
		"description":    "Testing default status",
	}
	
	req := TestRequest{
		method:      http.MethodPost,
		path:        "/api/v1/ideas",
		formData:    formData,
		contentType: "multipart/form-data",
		token:       userToken,
	}
	w := suite.MakeRequest(req)
	suite.Require().Equal(http.StatusCreated, w.Code) // Idea handler returns 201 Created
	
	var ideaResp dto.IdeaResponse
	err := json.Unmarshal(w.Body.Bytes(), &ideaResp)
	suite.Require().NoError(err)
	
	// Verify StatusID is not nil and corresponds to "Создана"
	suite.Require().NotNil(ideaResp.StatusID)
	
	var status models.IdeaStatus
	suite.DB.First(&status, "title = ?", "Создана")
	suite.Equal(status.ID, *ideaResp.StatusID)
}

func (suite *IdeaStatusTestSuite) TestUpdateIdeaStatus() {
	// Create Statuses
	status1 := &models.IdeaStatus{Title: "В работе"}
	status2 := &models.IdeaStatus{Title: "Done"}
	suite.DB.Create(status1)
	suite.DB.Create(status2)

	// User creates idea
	userName := "Creator"
	userPhone := "5555555555"
	user, shop := suite.CreateTestUser(userName, userPhone, "Creator Shop", "Addr", suite.UserRoleID)
	userToken := suite.GetAuthToken(userPhone, "123456", userName)
	
	desc := "D"
	cat := &models.Category{
		CoffeeShopID: &shop.ID,
		Title:        "Cat",
		Description:  &desc,
	}
	suite.DB.Create(cat)

	// Create idea (will have "Создана" status)
	idea := &models.Idea{
		CreatorID:    &user.ID,
		CoffeeShopID: &shop.ID,
		CategoryID:   &cat.ID,
		Title:        "My Idea",
		Description:  "Desc",
	}
	suite.DB.Create(idea)

	// Update status to "В работе"
	updateReq := dto.UpdateIdeaRequest{
		StatusID: &status1.ID,
	}
	
	req := TestRequest{
		method: http.MethodPut,
		path:   "/api/v1/ideas/" + idea.ID.String(),
		body:   updateReq,
		token:  userToken,
	}
	w := suite.MakeRequest(req)
	suite.Require().Equal(http.StatusNoContent, w.Code) // Expect 204
	
	// Verify update
	var updatedIdea models.Idea
	suite.DB.First(&updatedIdea, "id = ?", idea.ID)
	suite.Equal(status1.ID, *updatedIdea.StatusID)
}

func TestIdeaStatusTestSuite(t *testing.T) {
	suite.Run(t, new(IdeaStatusTestSuite))
}