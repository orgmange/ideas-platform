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

func (h *AuthHandler) GetOTP(c *gin.Context) {
	phone := c.Param("phone")
	err := h.uc.GetOTP(phone)
	if err != nil {
		handleAppErrors(err, h.logger, c)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "bad request"})
		return
	}

	token, err := h.uc.VerifyOTP(&req)
	if err != nil {
		handleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
