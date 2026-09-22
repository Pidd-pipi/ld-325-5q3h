package service

import (
	"fmt"
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"log/slog"
)

type ProductService struct {
	repo   repository.ProductRepository
	logger *slog.Logger
}

func NewProductService(repo repository.ProductRepository, logger *slog.Logger) *ProductService {
	return &ProductService{repo, logger}
}
func (s *ProductService) List(q dto.ProductQuery) ([]model.Product, int64, int, int, error) {
	if q.Page < constants.DefaultPage {
		q.Page = constants.DefaultPage
	}
	if q.PageSize < constants.DefaultPageSize {
		q.PageSize = constants.DefaultPageSize
	}
	if q.PageSize > constants.MaxPageSize {
		q.PageSize = constants.MaxPageSize
	}
	rows, total, err := s.repo.List(q.Query, q.Category, q.Sort, q.Page, q.PageSize)
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("list product service: %w", err)
	}
	return rows, total, q.Page, q.PageSize, nil
}
func (s *ProductService) Get(id uint) (model.Product, error)          { return s.repo.Get(id) }
func (s *ProductService) Compare(ids []uint) ([]model.Product, error) { return s.repo.Compare(ids) }
