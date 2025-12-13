package service

import (
	"GoCart/internal/domain"
	"GoCart/internal/repo"
	"errors"
	"fmt"
)

type ProductService interface {
	CheckProductExists(productID uint) bool
}

type CartService struct {
	repo          *repo.CartRepository
	productClient ProductService
}

func NewCartService(repo *repo.CartRepository, productClient ProductService) *CartService {
	return &CartService{repo: repo, productClient: productClient}
}

func (s *CartService) AddItem(userID, productID uint, quantity int) error {
	if quantity < 1 {
		return errors.New("quantity must be >= 1")
	}

	if !s.productClient.CheckProductExists(productID) {
		return fmt.Errorf("product %d not found", productID)
	}

	item, err := s.repo.GetItem(userID, productID)

	if err == nil {

		item.Quantity += quantity
		return s.repo.Update(item)
	}

	return s.repo.Create(&domain.CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
	})
}

func (s *CartService) UpdateItem(userID, productID uint, quantity int) error {
	if quantity < 1 {
		return errors.New("quantity must be >= 1")
	}

	item, err := s.repo.GetItem(userID, productID)
	if err != nil {
		return err
	}

	item.Quantity = quantity
	return s.repo.Update(item)
}

func (s *CartService) DeleteItem(userID, productID uint) error {
	return s.repo.Delete(userID, productID)
}

func (s *CartService) GetCart(userID uint) ([]domain.CartItem, error) {
	return s.repo.GetCart(userID)
}
