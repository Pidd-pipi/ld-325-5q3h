package service

import (
	"fmt"
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
	"time"
)

type PriceHistoryService struct {
	repo repository.PriceHistoryRepository
}
type TrendResult struct {
	Range   string               `json:"range"`
	Highest float64              `json:"highest"`
	Lowest  float64              `json:"lowest"`
	Average float64              `json:"average"`
	Points  []model.PriceHistory `json:"points"`
}

func NewPriceHistoryService(repo repository.PriceHistoryRepository) *PriceHistoryService {
	return &PriceHistoryService{repo}
}
func (s *PriceHistoryService) Trend(productID uint, rangeValue string) (TrendResult, error) {
	days := 30
	if rangeValue == constants.TrendDays90 {
		days = 90
	}
	if rangeValue == constants.TrendDaysYear {
		days = 365
	}
	rows, err := s.repo.List(productID, time.Now().AddDate(0, 0, -days))
	if err != nil {
		return TrendResult{}, fmt.Errorf("get trend: %w", err)
	}
	result := TrendResult{Range: rangeValue, Points: rows}
	for i, row := range rows {
		if i == 0 || row.Price > result.Highest {
			result.Highest = row.Price
		}
		if i == 0 || row.Price < result.Lowest {
			result.Lowest = row.Price
		}
		result.Average += row.Price
	}
	if len(rows) > 0 {
		result.Average /= float64(len(rows))
	}
	return result, nil
}
