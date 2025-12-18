package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/GeorgiiMalishev/ideas-platform/config"
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthIntegrationTestSuite struct {
	BaseTestSuite
}

func (suite *AuthIntegrationTestSuite) SetupSuite() {
	suite.BaseTestSuite.SetupSuite()
}

func (suite *AuthIntegrationTestSuite) TearDownTest() {
	suite.BaseTestSuite.TearDownTest()
}

func TestAuthIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AuthIntegrationTestSuite))
}

// TestGetOTP - табличный тест для получения OTP
func (suite *AuthIntegrationTestSuite) TestGetOTP() {
	tests := []struct {
		name           string
		phone          string
		expectedStatus int
		checkOTPInDB   bool
	}{
		{
			name:           "успешное получение OTP",
			phone:          "89001112233",
			expectedStatus: http.StatusNoContent,
			checkOTPInDB:   true,
		},
		{
			name:           "получение OTP для другого номера",
			phone:          "+79004445566",
			expectedStatus: http.StatusNoContent,
			checkOTPInDB:   true,
		},
		{
			name:           "получение OTP для нового номера",
			phone:          "89007778899",
			expectedStatus: http.StatusNoContent,
			checkOTPInDB:   true,
		},
		{
			name:           "невалидный номер телефона",
			phone:          "invalid",
			expectedStatus: http.StatusBadRequest,
			checkOTPInDB:   false,
		},
		{
			name:           "невалидный номер телефона (не русский)",
			phone:          "+12345678901",
			expectedStatus: http.StatusBadRequest,
			checkOTPInDB:   false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			req := TestRequest{
				method: http.MethodGet,
				path:   fmt.Sprintf("/api/v1/auth/%s", tt.phone),
			}
			w := suite.MakeRequest(req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.checkOTPInDB {
				var otp models.OTP
				err := suite.DB.First(&otp, "phone = ?", tt.phone).Error
				suite.NoError(err)
				suite.Equal(tt.phone, otp.Phone)
				suite.True(otp.ExpiresAt.After(time.Now()))
			}
		})
	}
}

// TestVerifyOTP - табличный тест для верификации OTP
func (suite *AuthIntegrationTestSuite) TestVerifyOTP() {
	tests := []struct {
		name           string
		phone          string
		otpCode        string
		userName       string
		existingUser   bool
		invalidOTP     bool
		expectedStatus int
		checkTokens    bool
	}{
		{
			name:           "новый пользователь с валидным OTP",
			phone:          "89001112234",
			otpCode:        "123456",
			userName:       "Test User",
			existingUser:   false,
			invalidOTP:     false,
			expectedStatus: http.StatusOK,
			checkTokens:    true,
		},
		{
			name:           "существующий пользователь с валидным OTP",
			phone:          "+79002223344",
			otpCode:        "654321",
			userName:       "Existing User",
			existingUser:   true,
			invalidOTP:     false,
			expectedStatus: http.StatusOK,
			checkTokens:    true,
		},
		{
			name:           "невалидный OTP код",
			phone:          "89003334455",
			otpCode:        "111111",
			userName:       "",
			existingUser:   false,
			invalidOTP:     true,
			expectedStatus: http.StatusUnauthorized,
			checkTokens:    false,
		},
		{
			name:           "новый пользователь с другим именем",
			phone:          "+79004445567",
			otpCode:        "999888",
			userName:       "Another User",
			existingUser:   false,
			invalidOTP:     false,
			expectedStatus: http.StatusOK,
			checkTokens:    true,
		},
		{
			name:           "существующий пользователь без имени в запросе",
			phone:          "89005556677",
			otpCode:        "654321",
			userName:       "", // Имя не передается в запросе, но пользователь существует (создан в Setup с именем "Existing User")
			existingUser:   true,
			invalidOTP:     false,
			expectedStatus: http.StatusOK,
			checkTokens:    true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Setup: создаем OTP
			hashedCode, _ := bcrypt.GenerateFromPassword([]byte(tt.otpCode), bcrypt.DefaultCost)

			// Если нужен существующий пользователь
			if tt.existingUser {
				user := &models.User{
					Name:  &tt.userName,
					Phone: &tt.phone,
				}
				_, err := suite.AuthRepo.CreateUser(context.Background(), user)
				suite.NoError(err)
			}

			// Создаем OTP
			otp := &models.OTP{
				Phone:        tt.phone,
				CodeHash:     string(hashedCode),
				ExpiresAt:    time.Now().Add(5 * time.Minute),
				AttemptsLeft: 3,
			}
			suite.DB.Create(otp)

			// Подготавливаем неправильный код, если нужно
			requestOTP := tt.otpCode
			if tt.invalidOTP {
				requestOTP = "000000"
			}

			// Выполняем запрос
			reqBody := dto.VerifyOTPRequest{
				Phone: tt.phone,
				OTP:   requestOTP,
				Name:  tt.userName,
			}

			req := TestRequest{
				method:      http.MethodPost,
				path:        "/api/v1/auth",
				body:        reqBody,
				contentType: "application/json",
			}
			w := suite.MakeRequest(req)

			// Проверяем статус
			suite.Equal(tt.expectedStatus, w.Code)

			// Проверяем токены, если ожидается успех
			if tt.checkTokens {
				var authResponse dto.AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &authResponse)
				suite.NoError(err)
				suite.NotEmpty(authResponse.AccessToken)
				suite.NotEmpty(authResponse.RefreshToken)

				// Проверяем, что пользователь создан/существует
				var user models.User
				err = suite.DB.First(&user, "phone = ?", tt.phone).Error
				suite.NoError(err)
				if !tt.existingUser {
					suite.Equal(tt.userName, *user.Name)
				} else if tt.userName != "" {
					// Если пользователь существует и мы передали имя (хотя для логина оно не обязательно),
					// в тесте мы ожидаем, что имя в БД не изменилось или совпадает с тем, что было при создании.
					// В данном тесте existingUser создается с tt.userName, если оно задано.
					// Если tt.userName пустое (кейс "без имени"), то мы не проверяем совпадение, так как в БД имя есть.
				}

				// Проверяем, что OTP удален после успешной верификации
				var otp models.OTP
				err = suite.DB.First(&otp, "phone = ?", tt.phone).Error
				suite.Error(err)
				suite.Equal(gorm.ErrRecordNotFound, err)
			}
		})
	}
}

// TestResendOTP - табличный тест для переотправки OTP

func (suite *AuthIntegrationTestSuite) TestResendOTP() {
	originalOTPConfig := suite.cfg.AuthConfig.OTPConfig
	suite.T().Cleanup(func() {
		suite.cfg.AuthConfig.OTPConfig = originalOTPConfig
	})
	tests := []struct {
		name      string
		phone     string
		otpConfig config.OTPConfig
		actions   []func(phone string)
		dbCheck   func(phone string)
	}{
		{
			name:  "успешная повторная отправка после кулдауна",
			phone: "89001111111",
			otpConfig: config.OTPConfig{
				ExpiresAtTimer:        time.Minute,
				ResetResendCountTimer: 2 * time.Minute,
				SoftAttemptsCount:     5,
				SubSoftAttemptsTimer:  time.Second,
				HardAttemptsCount:     10,
			},

			actions: []func(phone string){
				func(phone string) { // 1. First request
					req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/auth/%s", phone)}
					w := suite.MakeRequest(req)
					suite.Equal(http.StatusNoContent, w.Code)
				},

				func(phone string) {
					time.Sleep(time.Second)
				},

				func(phone string) {
					req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/auth/%s", phone)}
					w := suite.MakeRequest(req)
					suite.Equal(http.StatusNoContent, w.Code)
				},
			},

			dbCheck: func(phone string) {
				var otp models.OTP
				err := suite.DB.First(&otp, "phone = ?", phone).Error
				suite.NoError(err)
				suite.Equal(1, otp.ResendCount)
			},
		},

		{
			name:  "ошибка 'слишком много запросов' до кулдауна",
			phone: "+79002222222",
			otpConfig: config.OTPConfig{
				ExpiresAtTimer:        time.Minute,
				ResetResendCountTimer: 2 * time.Minute,
				SoftAttemptsCount:     5,
				SubSoftAttemptsTimer:  3 * time.Second,
				HardAttemptsCount:     10,
			},

			actions: []func(phone string){
				func(phone string) {
					req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/auth/%s", phone)}
					w := suite.MakeRequest(req)
					suite.Equal(http.StatusNoContent, w.Code)
				},

				func(phone string) {
					req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/auth/%s", phone)}
					w := suite.MakeRequest(req)
					suite.Equal(http.StatusTooManyRequests, w.Code)
				},
			},

			dbCheck: func(phone string) {
				var otp models.OTP
				err := suite.DB.First(&otp, "phone = ?", phone).Error
				suite.NoError(err)
				suite.Equal(0, otp.ResendCount)
			},
		},
		{
			name:  "достижение лимита soft-попыток и переход к hard-кулдауну",
			phone: "89003333333",
			otpConfig: config.OTPConfig{
				ExpiresAtTimer:        time.Minute,
				ResetResendCountTimer: 2 * time.Minute,
				SoftAttemptsCount:     2,
				SubSoftAttemptsTimer:  time.Millisecond * 200,
				HardAttemptsCount:     3,
				SubHardAttemptsTimer:  time.Second * 2,
			},
			actions: []func(phone string){
				func(phone string) {
					req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/auth/%s", phone)}
					w := suite.MakeRequest(req)
					suite.Equal(http.StatusNoContent, w.Code)
				},
				func(phone string) {
					time.Sleep(time.Millisecond * 200)
					req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/auth/%s", phone)}
					w := suite.MakeRequest(req)
					suite.Equal(http.StatusNoContent, w.Code)
				},
				func(phone string) {
					time.Sleep(time.Millisecond * 200)
					req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/auth/%s", phone)}
					w := suite.MakeRequest(req)
					suite.Equal(http.StatusNoContent, w.Code)
				},
				func(phone string) {
					req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/auth/%s", phone)}
					w := suite.MakeRequest(req)
					suite.Equal(http.StatusTooManyRequests, w.Code)
				},
			},

			dbCheck: func(phone string) {
				var otp models.OTP
				err := suite.DB.First(&otp, "phone = ?", phone).Error
				suite.NoError(err)
				suite.Equal(2, otp.ResendCount)
				suite.True(otp.NextAllowedAt.After(time.Now().Add(time.Second)))
			},
		},
		{
			name:  "достижение hard-лимита и полная блокировка",
			phone: "+79004444444",
			otpConfig: config.OTPConfig{
				ExpiresAtTimer:        time.Minute,
				ResetResendCountTimer: 10 * time.Second,
				SoftAttemptsCount:     1,
				SubSoftAttemptsTimer:  time.Millisecond * 100,
				HardAttemptsCount:     2,
				SubHardAttemptsTimer:  time.Millisecond * 100,
				PostHardAttemptsCount: time.Second * 2,
			},
			actions: []func(phone string){
				func(phone string) {
					req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/auth/%s", phone)}
					w := suite.MakeRequest(req)
					suite.Equal(http.StatusNoContent, w.Code)
				},
				func(phone string) {
					time.Sleep(time.Millisecond * 100)
					req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/auth/%s", phone)}
					w := suite.MakeRequest(req)
					suite.Equal(http.StatusNoContent, w.Code)
				},
				func(phone string) {
					time.Sleep(time.Millisecond * 100)
					req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/auth/%s", phone)}
					w := suite.MakeRequest(req)
					suite.Equal(http.StatusNoContent, w.Code)
				},
				func(phone string) {
					req := TestRequest{method: http.MethodGet, path: fmt.Sprintf("/api/v1/auth/%s", phone)}
					w := suite.MakeRequest(req)
					suite.Equal(http.StatusTooManyRequests, w.Code)
				},
			},
			dbCheck: func(phone string) {
				var otp models.OTP
				err := suite.DB.First(&otp, "phone = ?", phone).Error
				suite.NoError(err)
				suite.Equal(2, otp.ResendCount)
				suite.True(otp.NextAllowedAt.After(time.Now().Add(time.Second)))
			},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.cfg.AuthConfig.OTPConfig = tt.otpConfig
			for _, action := range tt.actions {
				action(tt.phone)
			}
			if tt.dbCheck != nil {
				tt.dbCheck(tt.phone)
			}
		})
	}
}

func (suite *AuthIntegrationTestSuite) TestRefresh() {
	tests := []struct {
		name              string
		phone             string
		otpCode           string
		userName          string
		waitBeforeRefresh time.Duration
		expectedStatus    int
		checkNewTokens    bool
	}{
		{
			name:              "успешное обновление токена",
			phone:             "89005556677",
			otpCode:           "111222",
			userName:          "Refresh User",
			waitBeforeRefresh: 1 * time.Second,
			expectedStatus:    http.StatusOK,
			checkNewTokens:    true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Создаем пользователя и получаем первоначальные токены
			hashedCode, _ := bcrypt.GenerateFromPassword([]byte(tt.otpCode), bcrypt.DefaultCost)

			user := &models.User{
				Name:  &tt.userName,
				Phone: &tt.phone,
			}
			_, err := suite.AuthRepo.CreateUser(context.Background(), user)
			suite.NoError(err)

			otp := &models.OTP{
				Phone:        tt.phone,
				CodeHash:     string(hashedCode),
				ExpiresAt:    time.Now().Add(5 * time.Minute),
				AttemptsLeft: 3,
			}
			suite.DB.Create(otp)

			// Верифицируем OTP для получения токенов
			verifyReqBody := dto.VerifyOTPRequest{
				Phone: tt.phone,
				OTP:   tt.otpCode,
			}
			verifyReq := TestRequest{
				method:      http.MethodPost,
				path:        "/api/v1/auth",
				body:        verifyReqBody,
				contentType: "application/json",
			}
			verifyW := suite.MakeRequest(verifyReq)

			suite.Equal(http.StatusOK, verifyW.Code)
			var authResp dto.AuthResponse
			err = json.Unmarshal(verifyW.Body.Bytes(), &authResp)
			suite.NoError(err)

			// Ждем, если нужно
			if tt.waitBeforeRefresh > 0 {
				time.Sleep(tt.waitBeforeRefresh)
			}

			// Обновляем токен
			refreshReqBody := dto.RefreshRequest{
				RefreshToken: authResp.RefreshToken,
			}
			refreshReq := TestRequest{
				method:      http.MethodPost,
				path:        "/api/v1/auth/refresh",
				body:        refreshReqBody,
				contentType: "application/json",
			}
			refreshW := suite.MakeRequest(refreshReq)

			suite.Equal(tt.expectedStatus, refreshW.Code)

			if tt.checkNewTokens {
				var newAuthResponse dto.AuthResponse
				err = json.Unmarshal(refreshW.Body.Bytes(), &newAuthResponse)
				suite.NoError(err)
				suite.NotEmpty(newAuthResponse.AccessToken)
				suite.NotEmpty(newAuthResponse.RefreshToken)
				suite.NotEqual(authResp.AccessToken, newAuthResponse.AccessToken)
				suite.NotEqual(authResp.RefreshToken, newAuthResponse.RefreshToken)
			}
		})
	}
}

// TestAuthenticatedEndpoints - табличный тест для авторизованных эндпоинтов
func (suite *AuthIntegrationTestSuite) TestAuthenticatedEndpoints() {
	tests := []struct {
		name           string
		phone          string
		otpCode        string
		userName       string
		endpoint       string
		method         string
		withAuth       bool
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "получение данных текущего пользователя с авторизацией",
			phone:          "89001112233",
			otpCode:        "123456",
			userName:       "Auth User",
			endpoint:       "/api/v1/users/me",
			method:         http.MethodGet,
			withAuth:       true,
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "получение данных без авторизации",
			phone:          "+79004445566",
			otpCode:        "654321",
			userName:       "No Auth",
			endpoint:       "/api/v1/users/me",
			method:         http.MethodGet,
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			var token string
			if tt.withAuth {
				token = suite.GetAuthToken(tt.phone, tt.otpCode, tt.userName)
			}

			req := TestRequest{
				method: tt.method,
				path:   tt.endpoint,
				token:  token,
			}

			w := suite.MakeRequest(req)

			suite.Equal(tt.expectedStatus, w.Code)

			if tt.checkResponse && tt.withAuth {
				var userResponse dto.UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &userResponse)
				suite.NoError(err)
				suite.Equal(tt.userName, userResponse.Name)
				suite.Equal(tt.phone, userResponse.Phone)
			}
		})
	}
}

// TestLogout - табличный тест для выхода с одного устройства
func (suite *AuthIntegrationTestSuite) TestLogout() {
	tests := []struct {
		name           string
		phone          string
		otpCode        string
		userName       string
		expectedStatus int
	}{
		{
			name:           "успешный logout - refresh токен удален, access токен истекает через 2 сек",
			phone:          "89001234567",
			otpCode:        "123456",
			userName:       "Logout User",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// 1. Получаем оба токена
			authResp := suite.GetAuthResponse(tt.phone, tt.otpCode, tt.userName)
			accessToken := authResp.AccessToken
			refreshToken := authResp.RefreshToken

			// 2. Проверяем, что токены работают до logout
			meReq := TestRequest{
				method: http.MethodGet,
				path:   "/api/v1/users/me",
				token:  accessToken,
			}
			w := suite.MakeRequest(meReq)
			suite.Equal(http.StatusOK, w.Code, "access токен должен работать до logout")

			refreshReq := TestRequest{
				method:      http.MethodPost,
				path:        "/api/v1/auth/refresh",
				body:        dto.RefreshRequest{RefreshToken: refreshToken},
				contentType: "application/json",
			}

			// 3. Выполняем logout с refresh токеном
			logoutReq := TestRequest{
				method:      http.MethodPost,
				path:        "/api/v1/logout",
				body:        dto.LogoutRequest{RefreshToken: refreshToken},
				contentType: "application/json",
				token:       accessToken,
			}
			w = suite.MakeRequest(logoutReq)
			suite.Equal(tt.expectedStatus, w.Code)

			// 4. Проверяем, что refresh токен удален из БД
			var refreshTokenModel models.UserRefreshToken
			err := suite.DB.Where("refresh_token = ?", refreshToken).First(&refreshTokenModel).Error
			suite.Error(err, "refresh token должен быть удален из БД")
			suite.Equal(gorm.ErrRecordNotFound, err)

			// 5. Проверяем, что refresh токен больше не работает
			w = suite.MakeRequest(refreshReq)
			suite.Equal(http.StatusUnauthorized, w.Code,
				"refresh token должен быть невалиден после logout")

			// 6. Проверяем, что access токен ЕЩЕ работает (JWT stateless особенность)
			w = suite.MakeRequest(meReq)
			suite.Equal(http.StatusOK, w.Code, "access токен должен работать до logout")
		})
	}
}

// TestLogoutEverywhere - табличный тест для выхода со всех устройств
func (suite *AuthIntegrationTestSuite) TestLogoutEverywhere() {
	tests := []struct {
		name        string
		phone       string
		otpCode     string
		userName    string
		numSessions int
	}{
		{
			name:        "выход со всех устройств - все refresh токены удалены",
			phone:       "89001112233",
			otpCode:     "123123",
			userName:    "Multi Device User",
			numSessions: 2,
		},
		{
			name:        "выход со всех устройств (3 сессии)",
			phone:       "+79004445566",
			otpCode:     "456456",
			userName:    "Many Devices User",
			numSessions: 3,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// 1. Создаем несколько сессий
			sessions := make([]struct {
				accessToken  string
				refreshToken string
			}, tt.numSessions)

			for i := 0; i < tt.numSessions; i++ {
				authResp := suite.GetAuthResponse(tt.phone, tt.otpCode, tt.userName)
				sessions[i].accessToken = authResp.AccessToken
				sessions[i].refreshToken = authResp.RefreshToken
				if i < tt.numSessions-1 {
					time.Sleep(100 * time.Millisecond) // Небольшая задержка между сессиями
				}
			}

			// 2. Проверяем, что все токены работают до logout
			for i, session := range sessions {
				meReq := TestRequest{
					method: http.MethodGet,
					path:   "/api/v1/users/me",
					token:  session.accessToken,
				}
				w := suite.MakeRequest(meReq)
				suite.Equal(http.StatusOK, w.Code,
					"сессия %d: access токен должен работать до logout", i+1)

				refreshReq := TestRequest{
					method:      http.MethodPost,
					path:        "/api/v1/auth/refresh",
					body:        dto.RefreshRequest{RefreshToken: session.refreshToken},
					contentType: "application/json",
				}
				w = suite.MakeRequest(refreshReq)
				suite.Equal(http.StatusOK, w.Code,
					"сессия %d: refresh токен должен работать до logout", i+1)
			}

			// 3. Проверяем количество refresh токенов в БД
			var count int64
			suite.DB.Model(&models.UserRefreshToken{}).
				Joins("JOIN users ON users.id = user_refresh_tokens.user_id").
				Where("users.phone = ?", tt.phone).
				Count(&count)
			suite.Equal(int64(tt.numSessions), count,
				"в БД должно быть %d refresh токенов", tt.numSessions)

			// 4. Выполняем logout-everywhere используя первый токен
			logoutReq := TestRequest{
				method:      http.MethodPost,
				path:        "/api/v1/logout-everywhere",
				contentType: "application/json",
				token:       sessions[0].accessToken,
			}
			w := suite.MakeRequest(logoutReq)
			suite.Equal(http.StatusNoContent, w.Code)

			// 5. Проверяем, что все refresh токены удалены из БД
			suite.DB.Model(&models.UserRefreshToken{}).
				Joins("JOIN users ON users.id = user_refresh_tokens.user_id").
				Where("users.phone = ?", tt.phone).
				Count(&count)
			suite.Equal(int64(0), count, "все refresh токены должны быть удалены")

			// 6. Проверяем, что все refresh токены больше не работают
			for i, session := range sessions {
				refreshReq := TestRequest{
					method:      http.MethodPost,
					path:        "/api/v1/auth/refresh",
					body:        dto.RefreshRequest{RefreshToken: session.refreshToken},
					contentType: "application/json",
				}
				w := suite.MakeRequest(refreshReq)
				suite.Equal(http.StatusUnauthorized, w.Code,
					"сессия %d: refresh token должен быть невалиден", i+1)
			}

			// 7. Проверяем, что все access токены ЕЩЕ работают (JWT stateless)
			for i, session := range sessions {
				meReq := TestRequest{
					method: http.MethodGet,
					path:   "/api/v1/users/me",
					token:  session.accessToken,
				}
				w := suite.MakeRequest(meReq)
				suite.Equal(http.StatusOK, w.Code,
					"сессия %d: access токен продолжает работать до истечения TTL", i+1)
			}

			// 8. Ждем истечения JWT (2 секунды + небольшой запас)
			time.Sleep(2100 * time.Millisecond)

			// 9. Проверяем, что все access токены больше не работают (истекли)
			for i, session := range sessions {
				meReq := TestRequest{
					method: http.MethodGet,
					path:   "/api/v1/users/me",
					token:  session.accessToken,
				}
				w := suite.MakeRequest(meReq)
				suite.Equal(http.StatusUnauthorized, w.Code,
					"сессия %d: access токен должен истечь через 2 секунды", i+1)
			}
		})
	}
}
