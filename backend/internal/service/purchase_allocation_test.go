package service

import (
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/model"
)

func allocOffer(id uint, price float64, moq, available int) model.Offer {
	offer := model.Offer{UnitPrice: price, MOQ: moq, AvailableQty: available}
	offer.ID = id
	return offer
}

func TestAllocate(t *testing.T) {
	cases := []struct {
		name      string
		offers    []model.Offer
		quantity  int
		wantPlan  []Allocation
		wantRejct string
	}{
		{
			name:     "single offer covers total",
			offers:   []model.Offer{allocOffer(1, 398, 10, 50)},
			quantity: 30,
			wantPlan: []Allocation{{Offer: allocOffer(1, 398, 10, 50), Quantity: 30}},
		},
		{
			name:     "splits across two offers when one is short",
			offers:   []model.Offer{allocOffer(1, 398, 10, 50), allocOffer(2, 412, 5, 30)},
			quantity: 60,
			wantPlan: []Allocation{{Offer: allocOffer(1, 398, 10, 50), Quantity: 50}, {Offer: allocOffer(2, 412, 5, 30), Quantity: 10}},
		},
		{
			name:      "rejects when total available is insufficient",
			offers:    []model.Offer{allocOffer(1, 398, 10, 50), allocOffer(2, 412, 5, 30)},
			quantity:  100,
			wantRejct: constants.RejectInsufficientStock,
		},
		{
			name:      "rejects when no offers are allocatable",
			offers:    nil,
			quantity:  10,
			wantRejct: constants.RejectNoAvailableOffers,
		},
		{
			name:      "rejects when remaining tail falls below every moq",
			offers:    []model.Offer{allocOffer(1, 398, 10, 50), allocOffer(2, 412, 5, 30)},
			quantity:  53,
			wantRejct: constants.RejectMOQUnfulfillable,
		},
		{
			name:     "skips offers whose available stock is below moq",
			offers:   []model.Offer{allocOffer(1, 100, 5, 3), allocOffer(2, 200, 5, 100)},
			quantity: 10,
			wantPlan: []Allocation{{Offer: allocOffer(2, 200, 5, 100), Quantity: 10}},
		},
		{
			name:     "allocates exact available quantity",
			offers:   []model.Offer{allocOffer(1, 398, 10, 40)},
			quantity: 40,
			wantPlan: []Allocation{{Offer: allocOffer(1, 398, 10, 40), Quantity: 40}},
		},
		{
			name:     "skips zero stock offers",
			offers:   []model.Offer{allocOffer(1, 100, 1, 0), allocOffer(2, 200, 2, 10)},
			quantity: 5,
			wantPlan: []Allocation{{Offer: allocOffer(2, 200, 2, 10), Quantity: 5}},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, reason := Allocate(tc.offers, tc.quantity)
			if reason != tc.wantRejct {
				t.Fatalf("reason = %q, want %q", reason, tc.wantRejct)
			}
			if tc.wantRejct != constants.RejectNone {
				if plan != nil {
					t.Fatalf("rejected allocation must not keep a partial plan, got %v", plan)
				}
				return
			}
			if len(plan) != len(tc.wantPlan) {
				t.Fatalf("plan length = %d, want %d", len(plan), len(tc.wantPlan))
			}
			total := 0
			for index, allocation := range plan {
				if allocation.Offer.ID != tc.wantPlan[index].Offer.ID || allocation.Quantity != tc.wantPlan[index].Quantity {
					t.Fatalf("plan[%d] = offer %d x %d, want offer %d x %d", index, allocation.Offer.ID, allocation.Quantity, tc.wantPlan[index].Offer.ID, tc.wantPlan[index].Quantity)
				}
				if allocation.Quantity < allocation.Offer.MOQ {
					t.Fatalf("plan[%d] quantity %d below moq %d", index, allocation.Quantity, allocation.Offer.MOQ)
				}
				if allocation.Quantity > allocation.Offer.AvailableQty {
					t.Fatalf("plan[%d] quantity %d above available %d", index, allocation.Quantity, allocation.Offer.AvailableQty)
				}
				total += allocation.Quantity
			}
			if total != tc.quantity {
				t.Fatalf("allocated total = %d, want %d", total, tc.quantity)
			}
		})
	}
}
