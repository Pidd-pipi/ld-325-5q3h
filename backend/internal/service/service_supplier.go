package service

import (
	"fmt"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
)

type SupplierService struct{ repo repository.SupplierRepository }

func NewSupplierService(repo repository.SupplierRepository) *SupplierService {
	return &SupplierService{repo}
}
func (s *SupplierService) List() ([]model.Supplier, error) {
	rows, err := s.repo.List()
	if err != nil {
		return nil, fmt.Errorf("list suppliers service: %w", err)
	}
	return rows, nil
}
func (s *SupplierService) UpdateStatus(id uint, status string) (model.Supplier, error) {
	return s.repo.UpdateStatus(id, status)
}
