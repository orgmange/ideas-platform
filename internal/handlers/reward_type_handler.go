package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
)

type RewardTypeHandler struct {
	uc     usecase.RewardTypeUsecase
	logger *slog.Logger
}

func NewRewardTypeHandler(uc usecase.RewardTypeUsecase, logger *slog.Logger) *RewardTypeHandler {
	return &RewardTypeHandler{
		uc:     uc,
		logger: logger,
	}
}

func (h *RewardTypeHandler) GetRewardType(c *gin.Context) {
	rewardTypeID, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}
	rewardType, err := h.uc.GetRewardType(rewardTypeID)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, rewardType)
}

func (h *RewardTypeHandler) GetRewardTypesByCoffeeShop(c *gin.Context) {
	coffeeShopID, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}
	pageRaw := c.Query("page")
	limitRaw := c.Query("limit")

	page, _ := strconv.Atoi(pageRaw)
	limit, _ := strconv.Atoi(limitRaw)

	rewardTypes, err := h.uc.GetRewardsTypesFromCoffeeShop(coffeeShopID, page, limit)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, rewardTypes)
}

func (h *RewardTypeHandler) CreateRewardType(c *gin.Context) {
	var req dto.CreateRewardTypeRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, &dto.ErrorResponse{Message: "bad request"})
		return
	}
	creatorID, ok := parseUserIDFromContext(h.logger, c)
	if !ok {
		return
	}
	creatorRole, ok := parseRoleFromContext(h.logger, c)
	if !ok {
		return
	}
	resp, err := h.uc.CreateRewardType(creatorID, creatorRole, &req)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *RewardTypeHandler) UpdateRewardType(c *gin.Context) {
	rewardTypeID, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}
	var req dto.UpdateRewardTypeRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, &dto.ErrorResponse{Message: "bad request"})
		return
	}
	updaterID, ok := parseUserIDFromContext(h.logger, c)
	if !ok {
		return
	}
	updaterRole, ok := parseRoleFromContext(h.logger, c)
	if !ok {
		return
	}
	err = h.uc.UpdateRewardType(updaterID, updaterRole, rewardTypeID, &req)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *RewardTypeHandler) DeleteRewardType(c *gin.Context) {
	rewardTypeID, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}
	deleterID, ok := parseUserIDFromContext(h.logger, c)
	if !ok {
		return
	}
	deleterRole, ok := parseRoleFromContext(h.logger, c)
	if !ok {
		return
	}
	err := h.uc.DeleteRewardType(deleterID, deleterRole, rewardTypeID)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.Status(http.StatusNoContent)
}
