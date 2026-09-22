package service

import (
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/gorm"
)

func offer(id, supplier uint, price float64, moq, stock int) model.Offer {
	return model.Offer{
		Model:          gorm.Model{ID: id},
		SupplierID:     supplier,
		Supplier:       model.Supplier{Model: gorm.Model{ID: supplier}, Name: "S"},
		UnitPrice:      price,
		MOQ:            moq,
		AvailableStock: stock,
	}
}

func quantityMap(lines []allocationLine) map[uint]int {
	out := map[uint]int{}
	for _, line := range lines {
		out[line.Offer.OfferID] = line.Quantity
	}
	return out
}

func TestPlanAllocation(t *testing.T) {
	cases := []struct {
		name    string
		offers  []model.Offer
		total   int
		ok      bool
		expect  map[uint]int
	}{
		{
			name:   "single supplier covers all",
			offers: []model.Offer{offer(1, 1, 10, 5, 100)},
			total:  20,
			ok:     true,
			expect: map[uint]int{1: 20},
		},
		{
			name:   "split when cheapest insufficient",
			offers: []model.Offer{offer(1, 1, 10, 5, 10), offer(2, 2, 12, 5, 50)},
			total:  30,
			ok:     true,
			expect: map[uint]int{1: 10, 2: 20},
		},
		{
			name:   "demand below moq goes to cheapest alone",
			offers: []model.Offer{offer(1, 1, 10, 5, 100), offer(2, 2, 12, 8, 100)},
			total:  6,
			ok:     true,
			expect: map[uint]int{1: 6},
		},
		{
			name:   "sub-moq remainder pulled back to hit next supplier moq",
			offers: []model.Offer{offer(1, 1, 10, 5, 20), offer(2, 2, 12, 8, 50)},
			total:  23,
			ok:     true,
			expect: map[uint]int{1: 15, 2: 8},
		},
		{
			name:   "pullback starts from most expensive prior supplier",
			offers: []model.Offer{offer(1, 1, 10, 5, 20), offer(2, 2, 11, 5, 30), offer(3, 3, 12, 10, 100)},
			total:  55,
			ok:     true,
			expect: map[uint]int{1: 20, 2: 25, 3: 10},
		},
		{
			name:   "total supply insufficient rejects",
			offers: []model.Offer{offer(1, 1, 10, 5, 10), offer(2, 2, 12, 5, 20)},
			total:  40,
			ok:     false,
		},
		{
			name:   "supplier with stock below moq skipped",
			offers: []model.Offer{offer(1, 1, 10, 50, 10), offer(2, 2, 12, 5, 30)},
			total:  20,
			ok:     true,
			expect: map[uint]int{2: 20},
		},
		{
			name:   "no usable supplier rejects",
			offers: []model.Offer{offer(1, 1, 10, 50, 10)},
			total:  5,
			ok:     false,
		},
		{
			name:   "never exceeds available stock",
			offers: []model.Offer{offer(1, 1, 10, 1, 7), offer(2, 2, 11, 1, 7)},
			total:  14,
			ok:     true,
			expect: map[uint]int{1: 7, 2: 7},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lines, ok := planAllocation(tc.offers, tc.total)
			if ok != tc.ok {
				t.Fatalf("ok=%v want %v (lines=%v)", ok, tc.ok, lines)
			}
			if !tc.ok {
				return
			}
			got := quantityMap(lines)
			byID := map[uint]model.Offer{}
			for _, o := range tc.offers {
				byID[o.ID] = o
			}
			sum := 0
			for offerID, qty := range got {
				sum += qty
				source := byID[offerID]
				if qty < source.MOQ {
					t.Fatalf("offer %d got %d below MOQ %d", offerID, qty, source.MOQ)
				}
				if qty > source.AvailableStock {
					t.Fatalf("offer %d got %d above stock %d", offerID, qty, source.AvailableStock)
				}
			}
			if sum != tc.total {
				t.Fatalf("allocated %d want total %d", sum, tc.total)
			}
			for offerID, want := range tc.expect {
				if got[offerID] != want {
					t.Fatalf("offer %d got %d want %d; full=%v", offerID, got[offerID], want, got)
				}
			}
		})
	}
}
