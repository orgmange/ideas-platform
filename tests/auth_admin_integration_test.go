package tests

import (
	"net/http"
	"testing"

	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/stretchr/testify/suite"
)

type AuthAdminIntegrationTestSuite struct {
	BaseTestSuite
}

func (suite *AuthAdminIntegrationTestSuite) SetupSuite() {
	suite.BaseTestSuite.SetupSuite()
}

func (suite *AuthAdminIntegrationTestSuite) TearDownTest() {
	suite.BaseTestSuite.TearDownTest()
}

func TestAuthAdminIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AuthAdminIntegrationTestSuite))
}

func (suite *AuthAdminIntegrationTestSuite) TestHealthCheck() {
	token := suite.RegisterUserAndGetToken(&models.User{
		Phone: "2685",
		Name:  "adminuser",
	})
	tests := []struct {
		name           string
		expectedStatus int
		token          string
		checkResponse  func(body []byte)
	}{
		{
			name:           "health check",
			expectedStatus: http.StatusOK,
			token:          token,
			checkResponse: func(body []byte) {
				suite.Contains(string(body), "ok")
			},
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			testRequest := TestRequest{
				method: "GET",
				path:   "/api/v1/admin/health",
				token:  test.token,
			}
			w := suite.MakeRequest(testRequest)
			suite.Equal(test.expectedStatus, w.Code)
			test.checkResponse(w.Body.Bytes())
		})
	}
}
