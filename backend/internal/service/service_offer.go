package service

import (
	"fmt"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
)

type OfferService struct{ repo repository.OfferRepository }

func NewOfferService(repo repository.OfferRepository) *OfferService { return &OfferService{repo} }
func (s *OfferService) List(productID uint) ([]model.Offer, error) {
	rows, err := s.repo.ListByProduct(productID)
	if err != nil {
		return nil, fmt.Errorf("list offer service: %w", err)
	}
	return rows, nil
}
func (s *OfferService) UpdateStatus(id uint, status string) (model.Offer, error) {
	return s.repo.UpdateStatus(id, status)
}
