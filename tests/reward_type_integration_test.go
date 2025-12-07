package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/stretchr/testify/suite"
)

type RewardTypeIntegrationTestSuite struct {
	BaseTestSuite
}

func (suite *RewardTypeIntegrationTestSuite) SetupSuite() {
	suite.BaseTestSuite.SetupSuite()
}

func (suite *RewardTypeIntegrationTestSuite) TearDownTest() {
	suite.BaseTestSuite.TearDownTest()
}

func TestRewardTypeIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(RewardTypeIntegrationTestSuite))
}

func (suite *RewardTypeIntegrationTestSuite) TestCreateRewardType() {
	tests := []struct {
		name           string
		expectedStatus int
		needCheckResp  bool
		setup          func() (string, dto.CreateRewardTypeRequest)
	}{
		{
			name:           "create reward type with coffee shop admin",
			expectedStatus: http.StatusCreated,
			needCheckResp:  true,
			setup: func() (string, dto.CreateRewardTypeRequest) {
				admin := models.User{
					Name:  "admin name",
					Phone: "7777",
				}
				token := suite.RegisterUserAndGetToken(&admin)
				shop, err := suite.CoffeeShopRepo.CreateCoffeeShop(suite.Ctx, &models.CoffeeShop{
					Name:      "test coffee shop",
					CreatorID: admin.ID,
				})
				suite.Require().NoError(err)

				// Make the admin a worker in the shop
				err = suite.DB.Create(&models.WorkerCoffeeShop{
					WorkerID:     &admin.ID,
					CoffeeShopID: &shop.ID,
					RoleID:       &suite.AdminRoleID,
				}).Error
				suite.Require().NoError(err)

				req := dto.CreateRewardTypeRequest{
					CoffeeShopID: shop.ID,
					Description:  "test reward type",
				}
				return token, req
			},
		},
		{
			name:           "create reward type with coffee shop not admin",
			expectedStatus: http.StatusForbidden,
			needCheckResp:  false,
			setup: func() (string, dto.CreateRewardTypeRequest) {
				user := models.User{
					Name:  "not admin name",
					Phone: "1241151",
				}
				token := suite.RegisterUserAndGetToken(&user)
				shop, err := suite.CoffeeShopRepo.CreateCoffeeShop(suite.Ctx, &models.CoffeeShop{
					Name:      "test coffee shop",
					CreatorID: user.ID,
				})
				suite.Require().NoError(err)
				req := dto.CreateRewardTypeRequest{
					CoffeeShopID: shop.ID,
					Description:  "test reward type",
				}
				return token, req
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			token, r := tt.setup()
			req := TestRequest{
				method:      http.MethodPost,
				path:        "/api/v1/admin/rewards/type",
				body:        r,
				contentType: "application/json",
				token:       token,
			}
			w := suite.MakeRequest(req)
			suite.Equal(tt.expectedStatus, w.Code)
			if tt.needCheckResp {
				var resp dto.RewardTypeResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				suite.Require().NoError(err)
				suite.Equal(resp.CoffeeShopID, r.CoffeeShopID)
				suite.Equal(resp.Description, r.Description)

			}
		})
	}
}
