package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
)

type IdeaHandler struct {
	uc     usecase.IdeaUsecase
	logger *slog.Logger
}

func NewIdeaHandler(uc usecase.IdeaUsecase, logger *slog.Logger) *IdeaHandler {
	return &IdeaHandler{
		uc:     uc,
		logger: logger,
	}
}

// @Summary Create a new idea
// @Description Create a new idea for a coffee shop
// @Tags ideas
// @Accept json
// @Produce json
// @Param idea body dto.CreateIdeaRequest true "Idea information"
// @Success 201 {object} dto.IdeaResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /ideas [post]
// @Security ApiKeyAuth
func (h *IdeaHandler) CreateIdea(c *gin.Context) {
	var req dto.CreateIdeaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind idea create request: ", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, ok := parseUserIDFromContext(c)
	if !ok {
		return
	}
	resp, err := h.uc.CreateIdea(userID, &req)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	h.logger.Info("idea created successfully: ", slog.String("idea_id", resp.ID.String()))
	c.JSON(http.StatusCreated, resp)
}

// @Summary Get all ideas by shop
// @Description Get a list of all ideas for a given coffee shop with optional pagination
// @Tags ideas
// @Produce json
// @Param shop_id path string true "Coffee Shop ID"
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param sort query string false "Sort order"
// @Success 200 {array} dto.IdeaResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /coffee-shops/{shop_id}/ideas [get]
func (h *IdeaHandler) GetIdeasFromShop(c *gin.Context) {
	shopID, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}
	pageRaw := c.Query("page")
	limitRaw := c.Query("limit")
	sort := c.Query("sort")

	page, _ := strconv.Atoi(pageRaw)
	limit, _ := strconv.Atoi(limitRaw)

	params := dto.GetIdeasRequest{
		Page:  page,
		Limit: limit,
		Sort:  sort,
	}

	resp, err := h.uc.GetAllIdeasByShop(shopID, params)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Get all ideas by user
// @Description Get a list of all ideas for a given user with optional pagination
// @Tags ideas
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param sort query string false "Sort order"
// @Success 200 {array} dto.IdeaResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/me/ideas [get]
// @Security ApiKeyAuth
func (h *IdeaHandler) GetIdeasFromUser(c *gin.Context) {
	userID, ok := parseUserIDFromContext(c)
	if !ok {
		return
	}
	var req dto.GetIdeasRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "bad request"})
		return
	}

	resp, err := h.uc.GetAllIdeasByUser(userID, req)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get idea by ID
// @Description Get idea details by ID
// @Tags ideas
// @Produce json
// @Param id path string true "Idea ID"
// @Success 200 {object} dto.IdeaResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /ideas/{id} [get]
func (h *IdeaHandler) GetIdea(c *gin.Context) {
	uuid, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}
	idea, err := h.uc.GetIdea(uuid)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}
	c.JSON(http.StatusOK, idea)
}

// @Summary Update idea by ID
// @Description Update idea details for the given ID
// @Tags ideas
// @Accept json
// @Produce json
// @Param id path string true "Idea ID"
// @Param idea body dto.UpdateIdeaRequest true "Idea update information"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /ideas/{id} [put]
// @Security ApiKeyAuth
func (h *IdeaHandler) UpdateIdea(c *gin.Context) {
	uuid, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}
	var req dto.UpdateIdeaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind idea update request: ", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	userID, ok := parseUserIDFromContext(c)
	if !ok {
		return
	}
	err := h.uc.UpdateIdea(userID, uuid, &req)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Delete idea by ID
// @Description Delete an idea by ID
// @Tags ideas
// @Produce json
// @Param id path string true "Idea ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /ideas/{id} [delete]
// @Security ApiKeyAuth
func (h *IdeaHandler) DeleteIdea(c *gin.Context) {
	uuid, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}
	userID, ok := parseUserIDFromContext(c)
	if !ok {
		return
	}
	err := h.uc.DeleteIdea(userID, uuid)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.Status(http.StatusNoContent)
}
