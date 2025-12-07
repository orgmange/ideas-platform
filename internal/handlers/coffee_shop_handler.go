package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
)

type CoffeeShopHandler struct {
	coffeeShopUsecase usecase.CoffeeShopUsecase
	logger            *slog.Logger
}

func NewCoffeeShopHandler(uc usecase.CoffeeShopUsecase, logger *slog.Logger) *CoffeeShopHandler {
	return &CoffeeShopHandler{uc, logger}
}

// @Summary Create a new coffee shop
// @Description Create a new coffee shop with the provided information
// @Tags coffee-shops
// @Accept json
// @Produce json
// @Param coffee_shop body dto.CreateCoffeeShopRequest true "Coffee shop information"
// @Success 201 {object} dto.CoffeeShopResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /coffee-shops [post]
// @Security ApiKeyAuth
func (h *CoffeeShopHandler) CreateCoffeeShop(c *gin.Context) {
	var req dto.CreateCoffeeShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind coffee shop create request: ", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}
	resp, err := h.coffeeShopUsecase.CreateCoffeeShop(c.Request.Context(), userID, &req)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	h.logger.Info("coffee shop created successfully: ", slog.String("coffee_shop_id", resp.ID.String()))
	c.JSON(http.StatusCreated, resp)
}

// @Summary Get all coffee shops
// @Description Get a list of all coffee shops with optional pagination
// @Tags coffee-shops
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Success 200 {array} dto.CoffeeShopResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /coffee-shops [get]
func (h *CoffeeShopHandler) GetAllCoffeeShops(c *gin.Context) {
	pageRaw := c.Query("page")
	limitRaw := c.Query("limit")

	page, _ := strconv.Atoi(pageRaw)
	limit, _ := strconv.Atoi(limitRaw)

	resp, err := h.coffeeShopUsecase.GetAllCoffeeShops(c.Request.Context(), page, limit)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary Get coffee shop by ID
// @Description Get coffee shop details by ID
// @Tags coffee-shops
// @Produce json
// @Param id path string true "Coffee Shop ID"
// @Success 200 {object} dto.CoffeeShopResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /coffee-shops/{id} [get]
func (h *CoffeeShopHandler) GetCoffeeShop(c *gin.Context) {
	uuid, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}
	shop, err := h.coffeeShopUsecase.GetCoffeeShop(c.Request.Context(), uuid)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}
	c.JSON(http.StatusOK, shop)
}

// @Summary Update coffee shop by ID
// @Description Update coffee shop details for the given ID
// @Tags coffee-shops
// @Accept json
// @Produce json
// @Param id path string true "Coffee Shop ID"
// @Param coffee_shop body dto.UpdateCoffeeShopRequest true "Coffee shop update information"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /coffee-shops/{id} [put]
// @Security ApiKeyAuth
func (h *CoffeeShopHandler) UpdateCoffeeShop(c *gin.Context) {
	uuid, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}
	var req dto.UpdateCoffeeShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind coffee shop update request: ", slog.String("error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	userID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}
	err := h.coffeeShopUsecase.UpdateCoffeeShop(c.Request.Context(), userID, uuid, &req)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Delete coffee shop by ID
// @Description Delete a coffee shop by ID
// @Tags coffee-shops
// @Produce json
// @Param id path string true "Coffee Shop ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /coffee-shops/{id} [delete]
// @Security ApiKeyAuth
func (h *CoffeeShopHandler) DeleteCoffeeShop(c *gin.Context) {
	uuid, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}
	userID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}
	err := h.coffeeShopUsecase.DeleteCoffeeShop(c.Request.Context(), userID, uuid)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.Status(http.StatusNoContent)
}
