package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
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

func (suite *RouterTestSuite) TestHealthCheck() {
	tests := []struct {
		name           string
		expectedStatus int
		checkResponse  func(body []byte)
	}{
		{
			name:           "health check",
			expectedStatus: http.StatusOK,
			checkResponse: func(body []byte) {
				suite.Contains(string(body), "ok")
			},
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			testRequest := TestRequest{
				method: "GET",
				path:   "/health",
			}
			w := suite.MakeRequest(testRequest)
			suite.Equal(test.expectedStatus, w.Code)
			test.checkResponse(w.Body.Bytes())
		})
	}
}

func (suite *RouterTestSuite) TestGetAllUsers() {
	type testCase struct {
		name           string
		setup          func()
		expectedStatus int
		checkResponse  func(body []byte)
	}

	tests := []testCase{
		{
			name: "get all with one user",
			setup: func() {
				_, err := suite.AuthRepo.CreateUser(&models.User{Name: "testuser", Phone: "12345"})
				suite.Require().NoError(err)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(body []byte) {
				var response []dto.UserResponse
				err := json.Unmarshal(body, &response)
				suite.NoError(err)
				suite.Len(response, 1)
				suite.Equal("testuser", response[0].Name)
			},
		},
		{
			name:           "get all with no users",
			setup:          func() { /* Do nothing, DB is clean */ },
			expectedStatus: http.StatusOK,
			checkResponse: func(body []byte) {
				var response []dto.UserResponse
				err := json.Unmarshal(body, &response)
				suite.NoError(err)
				suite.Len(response, 0)
			},
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			suite.TearDownTest() // Clean DB before each sub-test
			test.setup()
			testRequest := TestRequest{
				method: "GET",
				path:   "/api/v1/users",
			}
			w := suite.MakeRequest(testRequest)
			suite.Equal(test.expectedStatus, w.Code)
			test.checkResponse(w.Body.Bytes())
		})
	}
}

func (suite *RouterTestSuite) TestGetUser() {
	type testCase struct {
		name           string
		setup          func() string // returns userID
		expectedStatus int
		checkResponse  func(body []byte, userID string)
	}

	tests := []testCase{
		{
			name: "get existing user",
			setup: func() string {
				userID, err := suite.AuthRepo.CreateUser(&models.User{Name: "testuser", Phone: "12345"})
				suite.Require().NoError(err)
				return userID.String()
			}, expectedStatus: http.StatusOK,
			checkResponse: func(body []byte, userID string) {
				var response dto.UserResponse
				err := json.Unmarshal(body, &response)
				suite.NoError(err)
				suite.Equal("testuser", response.Name)
				uid, err := uuid.Parse(userID)
				suite.NoError(err)
				suite.Equal(uid, response.ID)
			},
		},
		{
			name: "get non-existing user",
			setup: func() string {
				return uuid.New().String()
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(body []byte, userID string) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				suite.NoError(err)
				suite.Contains(response["error"], "not found")
			},
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			userID := test.setup()
			testRequest := TestRequest{
				method: "GET",
				path:   fmt.Sprintf("/api/v1/users/%s", userID),
			}
			w := suite.MakeRequest(testRequest)
			suite.Equal(test.expectedStatus, w.Code)
			test.checkResponse(w.Body.Bytes(), userID)
		})
	}
}

func (suite *RouterTestSuite) TestUpdateUser() {
	type testCase struct {
		name           string
		setup          func() string // returns userID
		input          dto.UpdateUserRequest
		expectedStatus int
		checkResponse  func(w *httptest.ResponseRecorder, userID string)
	}
	tests := []testCase{
		{
			name: "update existing user",
			setup: func() string {
				userID, err := suite.AuthRepo.CreateUser(&models.User{Name: "testuser", Phone: "12345"})
				suite.Require().NoError(err)
				return userID.String()
			}, input: dto.UpdateUserRequest{
				Name: "updateduser",
			},
			expectedStatus: http.StatusNoContent,
			checkResponse: func(w *httptest.ResponseRecorder, userID string) {
				suite.Empty(w.Body.Bytes())

				getResp := suite.MakeRequest(TestRequest{
					method: "GET",
					path:   fmt.Sprintf("/api/v1/users/%s", userID),
				})
				suite.Equal(http.StatusOK, getResp.Code)

				var response dto.UserResponse
				err := json.Unmarshal(getResp.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Equal("updateduser", response.Name)
			},
		},
		{
			name: "update non-existing user",
			setup: func() string {
				return uuid.New().String()
			},
			input: dto.UpdateUserRequest{
				Name: "updateduser",
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(w *httptest.ResponseRecorder, userID string) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				suite.NoError(err)
				suite.Contains(response["error"], "not found")
			},
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			userID := test.setup()
			token := suite.GetRandomAuthToken()
			testRequest := TestRequest{
				method:      "PUT",
				path:        fmt.Sprintf("/api/v1/users/%s", userID),
				body:        test.input,
				contentType: "application/json",
				token:       token,
			}
			w := suite.MakeRequest(testRequest)
			suite.Equal(test.expectedStatus, w.Code)
			test.checkResponse(w, userID)
		})
	}
}

func (suite *RouterTestSuite) TestDeleteUser() {
	tests := []struct {
		name           string
		setup          func() string // Returns userID for the test
		expectedStatus int
		checkResponse  func(w *httptest.ResponseRecorder)
	}{
		{
			name: "delete existing user",
			setup: func() string {
				userID, err := suite.AuthRepo.CreateUser(&models.User{Name: "testuser", Phone: "12345"})
				suite.Require().NoError(err)
				return userID.String()
			},
			expectedStatus: http.StatusNoContent,
			checkResponse: func(w *httptest.ResponseRecorder) {
				// No body
			},
		},
		{
			name: "delete non-existing user",
			setup: func() string {
				return uuid.New().String()
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(w *httptest.ResponseRecorder) {
			},
		},
	}

	for _, test := range tests {
		test := test
		suite.Run(test.name, func() {
			userID := test.setup()
			token := suite.GetRandomAuthToken()
			testRequest := TestRequest{
				method: "DELETE",
				path:   fmt.Sprintf("/api/v1/users/%s", userID),
				token:  token,
			}
			w := suite.MakeRequest(testRequest)
			suite.Equal(test.expectedStatus, w.Code)
			test.checkResponse(w)
		})
	}
}
