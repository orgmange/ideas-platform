package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	uc     usecase.UserUsecase
	logger *slog.Logger
}

func NewUserHandler(uc usecase.UserUsecase, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		uc:     uc,
		logger: logger,
	}
}

// @Summary Get all users
// @Description Get a list of all users with optional pagination
// @Tags users
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Success 200 {array} dto.UserResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users [get]
func (h UserHandler) GetAllUsers(c *gin.Context) {
	pageRaw := c.Query("page")
	limitRaw := c.Query("limit")

	page, _ := strconv.Atoi(pageRaw)
	limit, _ := strconv.Atoi(limitRaw)

	resp, err := h.uc.GetAllUsers(page, limit)
	if err != nil {
		handleAppErrors(err, h.logger, c)
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Get user by ID
// @Description Get user details by user ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [get]
func (h UserHandler) GetUser(c *gin.Context) {
	uuid, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}
	user, err := h.uc.GetUser(*uuid)
	if err != nil {
		handleAppErrors(err, h.logger, c)
		return
	}
	c.JSON(http.StatusOK, user)
}

// @Summary Update user by ID
// @Description Update user details for the given user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body dto.UpdateUserRequest true "User update information"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [put]
func (h UserHandler) UpdateUser(c *gin.Context) {
	uuid, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}
	var req dto.UpdateUserRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		h.logger.Error("failed to bind user update request: ", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	err = h.uc.UpdateUser(*uuid, &req)
	if err != nil {
		handleAppErrors(err, h.logger, c)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Delete user by ID
// @Description Delete a user by user ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/{id} [delete]
func (h UserHandler) DeleteUser(c *gin.Context) {
	uuid, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}

	err := h.uc.DeleteUser(*uuid)
	if err != nil {
		handleAppErrors(err, h.logger, c)
	}

	c.Status(http.StatusNoContent)
}
