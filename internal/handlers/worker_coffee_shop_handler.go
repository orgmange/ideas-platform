package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
)

type WorkerCoffeeShopHandler struct {
	uc     usecase.WorkerCoffeeShopUsecase
	logger *slog.Logger
}

func NewWorkerCoffeeShopHandler(uc usecase.WorkerCoffeeShopUsecase, logger *slog.Logger) *WorkerCoffeeShopHandler {
	return &WorkerCoffeeShopHandler{
		uc:     uc,
		logger: logger,
	}
}

// @Summary Add a worker to a coffee shop
// @Description Adds a user as a worker to a specific coffee shop. Requires creator or admin access to the coffee shop.
// @Tags worker-coffee-shops
// @Accept json
// @Produce json
// @Param request body dto.AddWorkerToShopRequest true "Worker and Coffee Shop IDs"
// @Success 201 {object} dto.WorkerCoffeeShopResponse
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Not Found"
// @Failure 409 {object} dto.ErrorResponse "Conflict"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /worker-coffee-shops [post]
// @Security ApiKeyAuth
func (h *WorkerCoffeeShopHandler) AddWorker(c *gin.Context) {
	var req dto.AddWorkerToShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind add worker request", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "bad request"})
		return
	}

	actorID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}

	resp, err := h.uc.AddWorker(c.Request.Context(), actorID, &req)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	h.logger.Info("worker added to coffee shop successfully", slog.String("relation_id", resp.ID.String()))
	c.JSON(http.StatusCreated, resp)
}

// @Summary Remove a worker from a coffee shop
// @Description Removes a worker-coffee shop relationship by its ID. Requires creator or admin access to the coffee shop.
// @Tags worker-coffee-shops
// @Produce json
// @Param id path string true "Worker Coffee Shop Relationship ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Not Found"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /worker-coffee-shops/{id} [delete]
// @Security ApiKeyAuth
func (h *WorkerCoffeeShopHandler) RemoveWorker(c *gin.Context) {
	workerShopRelationID, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}

	actorID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}

	err := h.uc.RemoveWorker(c.Request.Context(), actorID, workerShopRelationID)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	h.logger.Info("worker removed from coffee shop successfully", slog.String("relation_id", workerShopRelationID.String()))
	c.Status(http.StatusNoContent)
}

// @Summary List workers in a coffee shop
// @Description Retrieves a paginated list of users working in a specific coffee shop. Requires admin access to the coffee shop.
// @Tags worker-coffee-shops
// @Produce json
// @Param id path string true "Coffee Shop ID"
// @Param page query int false "Page number" default(0)
// @Param limit query int false "Items per page" default(25)
// @Success 200 {array} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Not Found"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /admin/coffee-shops/{id}/workers [get]
// @Security ApiKeyAuth
func (h *WorkerCoffeeShopHandler) ListWorkersInShop(c *gin.Context) {
	shopID, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}

	actorID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}

	pageRaw := c.Query("page")
	limitRaw := c.Query("limit")

	page, _ := strconv.Atoi(pageRaw)
	limit, _ := strconv.Atoi(limitRaw)

	resp, err := h.uc.ListWorkers(c.Request.Context(), actorID, shopID, page, limit)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	h.logger.Info("listed workers for coffee shop", slog.String("shop_id", shopID.String()), slog.Int("count", len(resp)))
	c.JSON(http.StatusOK, resp)
}

// @Summary List coffee shops for a worker
// @Description Retrieves a paginated list of coffee shops a specific user works for. A user can only view their own list.
// @Tags worker-coffee-shops
// @Produce json
// @Param id path string true "Worker User ID"
// @Param page query int false "Page number" default(0)
// @Param limit query int false "Items per page" default(25)
// @Success 200 {array} dto.CoffeeShopResponse
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Not Found"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /users/{id}/coffee-shops [get]
// @Security ApiKeyAuth
func (h *WorkerCoffeeShopHandler) ListCoffeeShopsForWorker(c *gin.Context) {
	workerID, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}

	actorID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}

	pageRaw := c.Query("page")
	limitRaw := c.Query("limit")

	page, _ := strconv.Atoi(pageRaw)
	limit, _ := strconv.Atoi(limitRaw)

	resp, err := h.uc.ListShopsForWorker(c.Request.Context(), actorID, workerID, page, limit)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	h.logger.Info("listed coffee shops for worker", slog.String("worker_id", workerID.String()), slog.Int("count", len(resp)))
	c.JSON(http.StatusOK, resp)
}
