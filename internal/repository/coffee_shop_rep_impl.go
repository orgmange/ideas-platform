package repository

import (
	"database/sql"
	"fmt"

	apperrors "github.com/GeorgiiMalishev/ideas-platform/internal/app_errors"
	"github.com/GeorgiiMalishev/ideas-platform/internal/models"
	"github.com/google/uuid"
)

type CoffeeShopRepImpl struct {
	db *sql.DB
}

func NewCoffeeShopRepository(db *sql.DB) CoffeeShopRep {
	return &CoffeeShopRepImpl{db: db}
}

func (r *CoffeeShopRepImpl) CreateCoffeeShop(shop *models.CoffeeShop) (*models.CoffeeShop, error) {
	query := `INSERT INTO coffee_shop (name, address, contacts, welcome_message, rules) 
			   VALUES ($1, $2, $3, $4, $5) 
			   RETURNING id, name, address, contacts, welcome_message, rules, updated_at, created_at`
	var createdShop models.CoffeeShop
	err := r.db.QueryRow(query, shop.Name, shop.Address, shop.Contacts, shop.WelcomeMessage, shop.Rules).
		Scan(&createdShop.ID, &createdShop.Name, &createdShop.Address, &createdShop.Contacts, &createdShop.WelcomeMessage, &createdShop.Rules, &createdShop.UpdatedAt, &createdShop.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &createdShop, nil
}

func (r *CoffeeShopRepImpl) UpdateCoffeeShop(shop *models.CoffeeShop) error {
	query := `UPDATE coffee_shop 
			   SET name = $1, address = $2, contacts = $3, welcome_message = $4, rules = $5, updated_at = NOW() 
			   WHERE id = $6`
	_, err := r.db.Exec(query, shop.Name, shop.Address, shop.Contacts, shop.WelcomeMessage, shop.Rules, shop.ID)
	if err == sql.ErrNoRows {
		return apperrors.NewErrNotFound("coffee_shop", shop.ID.String())
	}
	return err
}

func (r *CoffeeShopRepImpl) DeleteCoffeeShop(ID uuid.UUID) error {
	query := `DELETE FROM coffee_shop WHERE id = $1`
	_, err := r.db.Exec(query, ID)
	if err == sql.ErrNoRows {
		return apperrors.NewErrNotFound("coffee_shop", ID.String())
	}
	return err
}

func (r *CoffeeShopRepImpl) GetCoffeeShop(ID uuid.UUID) (*models.CoffeeShop, error) {
	query := `SELECT id, name, address, contacts, welcome_message, rules, updated_at, created_at 
			   FROM coffee_shop 
			   WHERE id = $1`
	var shop models.CoffeeShop
	err := r.db.QueryRow(query, ID).
		Scan(&shop.ID, &shop.Name, &shop.Address, &shop.Contacts, &shop.WelcomeMessage, &shop.Rules, &shop.UpdatedAt, &shop.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewErrNotFound("coffee_shop", ID.String())
		}
		return nil, err
	}
	return &shop, nil
}

func (r *CoffeeShopRepImpl) GetAllCoffeeShops(limit, offset int) ([]models.CoffeeShop, error) {
	query := `SELECT id, name, address, contacts, welcome_message, rules, updated_at, created_at 
			   FROM coffee_shop
			   LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shops []models.CoffeeShop
	for rows.Next() {
		var shop models.CoffeeShop
		if err := rows.Scan(&shop.ID, &shop.Name, &shop.Address, &shop.Contacts, &shop.WelcomeMessage, &shop.Rules, &shop.UpdatedAt, &shop.CreatedAt); err != nil {
			return nil, err
		}
		shops = append(shops, shop)
	}

	return shops, nil
}

func (r *CoffeeShopRepImpl) IsCoffeeShopExist(ID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM coffee_shop WHERE id = $1)`
	var exists bool
	err := r.db.QueryRow(query, ID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check coffee shop existence: %w", err)
	}
	return exists, nil
}
