package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/GeorgiiMalishev/ideas-platform/internal/usecase"
	"github.com/gin-gonic/gin"
)

type RewardHandler struct {
	uc     usecase.RewardUsecase
	logger *slog.Logger
}

func NewRewardHandler(uc usecase.RewardUsecase, logger *slog.Logger) *RewardHandler {
	return &RewardHandler{
		uc:     uc,
		logger: logger,
	}
}

// @Summary Give a reward
// @Description Creates a reward for the author of a specific idea. The caller must be the owner of the coffee shop.
// @Tags rewards
// @Accept json
// @Produce json
// @Param reward_info body dto.GiveRewardRequest true "Information about the reward to be given"
// @Success 201 {object} dto.RewardResponse
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Not Found"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /rewards [post]
// @Security ApiKeyAuth
func (h *RewardHandler) GiveReward(c *gin.Context) {
	var req dto.GiveRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("failed to bind give reward request", "error", err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Message: "bad request: " + err.Error()})
		return
	}

	actorID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}

	resp, err := h.uc.GiveReward(c.Request.Context(), actorID, &req)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	h.logger.Info("reward given successfully", "reward_id", resp.ID.String())
	c.JSON(http.StatusCreated, resp)
}

// @Summary Revoke a reward
// @Description Deletes a reward by its ID. The caller must be the owner of the coffee shop.
// @Tags rewards
// @Produce json
// @Param id path string true "Reward ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 404 {object} dto.ErrorResponse "Not Found"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /rewards/{id} [delete]
// @Security ApiKeyAuth
func (h *RewardHandler) RevokeReward(c *gin.Context) {
	rewardID, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}

	actorID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}

	err := h.uc.RevokeReward(c.Request.Context(), actorID, rewardID)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Get a reward
// @Description Retrieves details of a single reward by its ID.
// @Tags rewards
// @Produce json
// @Param id path string true "Reward ID"
// @Success 200 {object} dto.RewardResponse
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 404 {object} dto.ErrorResponse "Not Found"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /rewards/{id} [get]
func (h *RewardHandler) GetReward(c *gin.Context) {
	rewardID, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}

	resp, err := h.uc.GetReward(c.Request.Context(), rewardID)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get rewards for a coffee shop
// @Description Retrieves a paginated list of rewards associated with a specific coffee shop. The caller must be the owner.
// @Tags rewards
// @Produce json
// @Param id path string true "Coffee Shop ID"
// @Param page query int false "Page number" default(0)
// @Param limit query int false "Items per page" default(25)
// @Success 200 {array} dto.RewardResponse
// @Failure 400 {object} dto.ErrorResponse "Bad Request"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 403 {object} dto.ErrorResponse "Forbidden"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /coffee-shops/{id}/rewards [get]
// @Security ApiKeyAuth
func (h *RewardHandler) GetRewardsForCoffeeShop(c *gin.Context) {
	coffeeShopID, ok := parseUUID(h.logger, c)
	if !ok {
		return
	}

	actorID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	resp, err := h.uc.GetRewardsForCoffeeShop(c.Request.Context(), actorID, coffeeShopID, page, limit)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Get my rewards
// @Description Retrieves a paginated list of rewards the currently authenticated user has received.
// @Tags rewards
// @Produce json
// @Param page query int false "Page number" default(0)
// @Param limit query int false "Items per page" default(25)
// @Success 200 {array} dto.RewardResponse
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal Server Error"
// @Router /users/me/rewards [get]
// @Security ApiKeyAuth
func (h *RewardHandler) GetMyRewards(c *gin.Context) {
	userID, ok := parseActorIDFromContext(h.logger, c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	resp, err := h.uc.GetMyRewards(c.Request.Context(), userID, page, limit)
	if err != nil {
		HandleAppErrors(err, h.logger, c)
		return
	}

	c.JSON(http.StatusOK, resp)
}
