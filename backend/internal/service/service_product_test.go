package service

import (
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"io"
	"log/slog"
	"testing"
)

type mockProductRepo struct{}

func (mockProductRepo) List(string, string, string, int, int) ([]model.Product, int64, error) {
	return []model.Product{{Name: "岩板"}}, 1, nil
}
func (mockProductRepo) Get(uint) (model.Product, error)         { return model.Product{}, nil }
func (mockProductRepo) Compare([]uint) ([]model.Product, error) { return nil, nil }
func TestProductServiceDefaultsPagination(t *testing.T) {
	svc := NewProductService(mockProductRepo{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	rows, total, page, pageSize, err := svc.List(dto.ProductQuery{})
	if err != nil || len(rows) != 1 || total != 1 {
		t.Fatalf("unexpected result: %d %d %v", len(rows), total, err)
	}
	if page != 1 || pageSize != 12 {
		t.Fatalf("unexpected pagination defaults: page=%d page_size=%d", page, pageSize)
	}
}
