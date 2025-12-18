package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"github.com/google/uuid"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
)

type IdeaHandler struct {
	uc          usecase.IdeaUsecase
	imageUsecase usecase.ImageUsecase
	logger      *slog.Logger
}

func NewIdeaHandler(uc usecase.IdeaUsecase, imageUsecase usecase.ImageUsecase, logger *slog.Logger) *IdeaHandler {
	return &IdeaHandler{
		uc:          uc,
		imageUsecase: imageUsecase,
		logger:      logger,
	}
}

// @Summary Create a new idea
// @Description Create a new idea for a coffee shop
// @Tags ideas
// @Accept mpfd
// @Produce json
// @Param idea formData dto.CreateIdeaRequest true "Idea information"
// @Param image formData file false "Image file"
// @Success 201 {object} dto.IdeaResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /ideas [post]
// @Security ApiKeyAuth
func (h *IdeaHandler) CreateIdea(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	coffeeShopIDStr := c.PostForm("coffee_shop_id")
	categoryIDStr := c.PostForm("category_id")

	coffeeShopID, err := uuid.Parse(coffeeShopIDStr)
	if err != nil {
		h.logger.Error("failed to parse coffeeShopID: ", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coffee_shop_id"})
		return
	}
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		h.logger.Error("failed to parse categoryID: ", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		return
	}

	req := &dto.CreateIdeaRequest{
		Title:        title,
		Description:  description,
		CoffeeShopID: coffeeShopID,
		CategoryID:   categoryID,
	}

	userID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}

	file, err := c.FormFile("image")
	if err != nil && err != http.ErrMissingFile {
		h.logger.Error("failed to get image from form: ", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image file"})
		return
	}

	var imageURL *string
	if file != nil {
		uploadedURL, err := h.imageUsecase.UploadImage(c.Request.Context(), file)
		if err != nil {
			h.logger.Error("failed to upload image: ", slog.String("error", err.Error()))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload image"})
			return
		}
		imageURL = &uploadedURL
	}

	resp, err := h.uc.CreateIdea(c.Request.Context(), userID, req, imageURL)
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
// @Param id path string true "Coffee Shop ID"
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param sort query string false "Sort order"
// @Success 200 {array} dto.IdeaResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /coffee-shops/{id}/ideas [get]
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

	resp, err := h.uc.GetAllIdeasByShop(c.Request.Context(), shopID, params)
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
	userID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}
	pageRaw := c.Query("page")
	limitRaw := c.Query("limit")
	sort := c.Query("sort")

	page, _ := strconv.Atoi(pageRaw)
	limit, _ := strconv.Atoi(limitRaw)

	req := dto.GetIdeasRequest{
		Page:  page,
		Limit: limit,
		Sort:  sort,
	}

	resp, err := h.uc.GetAllIdeasByUser(c.Request.Context(), userID, req)
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
	idea, err := h.uc.GetIdea(c.Request.Context(), uuid)
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
	userID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}
	err := h.uc.UpdateIdea(c.Request.Context(), userID, uuid, &req)
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
	userID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}
	err := h.uc.DeleteIdea(c.Request.Context(), userID, uuid)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.Status(http.StatusNoContent)
}
