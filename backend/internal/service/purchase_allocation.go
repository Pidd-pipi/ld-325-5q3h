package service

import (
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/model"
)

// Allocation is one split line of a purchase order: how many units of the
// requested quantity are assigned to a single offer.
type Allocation struct {
	Offer    model.Offer
	Quantity int
}

// Allocate splits the requested quantity across the given offers, which must
// already be filtered to in-stock offers of approved suppliers and ordered by
// allocation preference (cheapest first). Every line respects the offer MOQ
// and available quantity. When the total cannot be met exactly, no partial
// plan is kept and a rejection reason code is returned instead.
func Allocate(offers []model.Offer, quantity int) ([]Allocation, string) {
	remaining := quantity
	plan := make([]Allocation, 0, len(offers))
	for _, offer := range offers {
		if remaining == 0 {
			break
		}
		if offer.AvailableQty <= 0 {
			continue
		}
		take := min(remaining, offer.AvailableQty)
		if take < offer.MOQ {
			continue
		}
		plan = append(plan, Allocation{Offer: offer, Quantity: take})
		remaining -= take
	}
	if remaining > 0 {
		return nil, rejectReason(offers, quantity)
	}
	return plan, constants.RejectNone
}

func rejectReason(offers []model.Offer, quantity int) string {
	if len(offers) == 0 {
		return constants.RejectNoAvailableOffers
	}
	available := 0
	for _, offer := range offers {
		available += offer.AvailableQty
	}
	if available < quantity {
		return constants.RejectInsufficientStock
	}
	return constants.RejectMOQUnfulfillable
}
