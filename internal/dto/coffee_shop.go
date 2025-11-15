package dto

import "github.com/google/uuid"

type CreateCoffeeShopRequest struct {
	Name           string  `json:"name"`
	Address        string  `json:"address"`
	Contacts       *string `json:"contacts"`
	WelcomeMessage *string `json:"welcome_message"`
	Rules          *string `json:"rules"`
}

type UpdateCoffeeShopRequest struct {
	Name           string  `json:"name"`
	Address        string  `json:"address"`
	Contacts       *string `json:"contacts"`
	WelcomeMessage *string `json:"welcome_message"`
	Rules          *string `json:"rules"`
}

type CoffeeShopResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Address        string    `json:"address"`
	Contacts       *string   `json:"contacts"`
	WelcomeMessage *string   `json:"welcome_message"`
	Rules          *string   `json:"rules"`
}

