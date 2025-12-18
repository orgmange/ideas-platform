package handlers

import (
	"log/slog"
	"net/http"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IdeaStatusHandler struct {
	statusUsecase usecase.IdeaStatusUsecase
	logger        *slog.Logger
}

func NewIdeaStatusHandler(statusUsecase usecase.IdeaStatusUsecase, logger *slog.Logger) *IdeaStatusHandler {
	return &IdeaStatusHandler{
		statusUsecase: statusUsecase,
		logger:        logger,
	}
}

func (h *IdeaStatusHandler) Create(c *gin.Context) {
	var req dto.CreateIdeaStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind json", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.statusUsecase.Create(c.Request.Context(), req)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetAllStatuses godoc
// @Summary Get all idea statuses
// @Description Get all idea statuses
// @Tags statuses
// @Produce json
// @Success 200 {array} dto.IdeaStatusResponse
// @Failure 500 {object} map[string]string
// @Router /statuses [get]
func (h *IdeaStatusHandler) GetAll(c *gin.Context) {
	statuses, err := h.statusUsecase.GetAll(c.Request.Context())
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, statuses)
}

// GetStatusByID godoc
// @Summary Get idea status by ID
// @Description Get idea status by ID
// @Tags statuses
// @Produce json
// @Param id path string true "Status ID"
// @Success 200 {object} dto.IdeaStatusResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /statuses/{id} [get]
func (h *IdeaStatusHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	status, err := h.statusUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, status)
}

func (h *IdeaStatusHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	var req dto.UpdateIdeaStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind json", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.statusUsecase.Update(c.Request.Context(), id, req); err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status updated successfully"})
}

func (h *IdeaStatusHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	if err := h.statusUsecase.Delete(c.Request.Context(), id); err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "status deleted successfully"})
}
