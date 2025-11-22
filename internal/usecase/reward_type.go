package usecase

import (
	"github.com/GeorgiiMalishev/ideas-platform/internal/dto"
	"github.com/google/uuid"
)

type RewardTypeUsecase interface {
	// GetRewardType получает тип награды по его uuid. Возвращает ошибку, если типа награды с таким uuid не существует.
	GetRewardType(rewardTypeID uuid.UUID) (*dto.RewardTypeResponse, error)

	// GetRewardsTypeFromCoffeeShop Получить все типы нарграды по uuid Кофейни на которую они зарегестрированны. Возвращает ошибку, если есть проблемы с БД.
	GetRewardsTypesFromCoffeeShop(coffeeShopID uuid.UUID, page, limit int) ([]dto.RewardTypeResponse, error)

	// CreateRewardType Создать тип награды. Проверяется принадлежность создателя к кофейне и роль создателя.
	// Возвращает ошибку если этот пользователь не обладает правами чтобы создать тип награды в этой кофене
	// или когда произошли ошибки валидации
	CreateRewardType(creatorID uuid.UUID, creatorRole string, request *dto.CreateRewardTypeRequest) (*dto.RewardTypeResponse, error)

	// UpdateRewardType Обновить тип награды. Проверяется принадлежность пользователя к кофейне и роль пользователя.
	// Возвращает ошибку если этот пользователь не обладает правами чтобы обновить тип награды в этой кофене
	// или когда произошли ошибки валидации
	UpdateRewardType(updaterID uuid.UUID, updaterRole string, rewardTypeID uuid.UUID, request *dto.UpdateRewardTypeRequest) error

	// DeleteRewardType Удалить тип награды. Проверяется принадлежность пользователя к кофейне и роль пользователя.
	// Возвращает ошибку если этот пользователь не обладает правами чтобы удалить тип награды в этой кофене
	// или если типа награды с таким uuid нет
	DeleteRewardType(deleterID uuid.UUID, deleterRole string, rewardTypeID uuid.UUID) error
}
