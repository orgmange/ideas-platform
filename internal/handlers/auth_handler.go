package handlers

import (
	"log/slog"
	"net/http"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	uc     usecase.AuthUsecase
	logger *slog.Logger
}

func NewAuthHandler(uc usecase.AuthUsecase, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		uc:     uc,
		logger: logger,
	}
}

// @Summary Get OTP
// @Description Get One-Time Password for phone number
// @Tags auth
// @Produce json
// @Param phone path string true "Phone number"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /auth/{phone} [get]
func (h *AuthHandler) GetOTP(c *gin.Context) {
	phone := c.Param("phone")
	err := h.uc.GetOTP(c.Request.Context(), phone)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Verify OTP
// @Description Verify One-Time Password and authenticate user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.VerifyOTPRequest true "OTP verification request"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /auth [post]
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "bad request"})
		return
	}

	authResp, err := h.uc.VerifyOTP(c.Request.Context(), &req)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, authResp)
}

// @Summary Register Admin and Coffee Shop
// @Description Register a new admin user and create a coffee shop associated with them
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterAdminRequest true "Admin registration and coffee shop creation request"
// @Success 200 {object} dto.AdminAuthResponse
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 409 {object} dto.ErrorResponse "Conflict (e.g., login already exists)"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /auth/register/admin [post]
func (h *AuthHandler) RegisterAdminAndCoffeeShop(c *gin.Context) {
	var req dto.RegisterAdminRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "bad request"})
		return
	}

	authResp, err := h.uc.RegisterAdminAndCoffeeShop(c.Request.Context(), &req)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, authResp)
}

// @Summary Login Admin
// @Description Login an admin user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.AdminLoginRequest true "Admin login request"
// @Success 200 {object} dto.AdminAuthResponse
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /auth/login/admin [post]
func (h *AuthHandler) LoginAdmin(c *gin.Context) {
	var req dto.AdminLoginRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "bad request"})
		return
	}

	authResp, err := h.uc.LoginAdmin(c.Request.Context(), &req)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, authResp)
}

// @Summary Refresh Access Token
// @Description Refresh access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Refresh token request"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var refreshReq dto.RefreshRequest
	err := c.ShouldBindJSON(&refreshReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "bad request"})
		return
	}

	authResp, err := h.uc.Refresh(c.Request.Context(), refreshReq.RefreshToken)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, authResp)
}

// @Summary Logout
// @Description Logout user by invalidating refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LogoutRequest true "Logout request"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /logout [post]
// @Security ApiKeyAuth
func (h *AuthHandler) Logout(c *gin.Context) {
	var logoutReq dto.LogoutRequest
	err := c.ShouldBindJSON(&logoutReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "bad request"})
		return
	}

	err = h.uc.Logout(c.Request.Context(), logoutReq.RefreshToken)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Logout Everywhere
// @Description Logout user from all devices by invalidating all refresh tokens
// @Tags auth
// @Produce json
// @Success 204 "No Content"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /logout-everywhere [post]
// @Security ApiKeyAuth
func (h *AuthHandler) LogoutEverywhere(c *gin.Context) {
	userID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}
	err := h.uc.LogoutEverywhere(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Message: "internal server error"})
		return
	}

	c.Status(http.StatusNoContent)
}
