package service

import (
	"github.com/blueship581/cybuildprice/backend/internal/model"
)

// allocationOffer is the lock-consistent snapshot of an eligible quote:
// supplier approved, offer in stock.
type allocationOffer struct {
	OfferID      uint
	SupplierID   uint
	SupplierName string
	UnitPrice    float64
	MOQ          int
	Available    int
	DeliveryDays int
	Freight      string
}

// allocationLine is one supplier allocation produced by the splitter.
type allocationLine struct {
	Offer      allocationOffer
	Quantity   int
	LineAmount float64
}

// planAllocation splits totalQuantity across the cheapest eligible quotes.
// Rules:
//   - every used supplier receives >= its MOQ and <= its available stock;
//   - a single supplier shortfall is covered by splitting to other suppliers;
//   - cheaper suppliers are filled first, up to their stock;
//   - when the next supplier only needs a sub-MOQ remainder, quantity is pulled
//     back from earlier suppliers (down to their MOQ) so this supplier can take
//     a full MOQ, keeping the grand total exactly totalQuantity; pullback starts
//     from the most expensive earlier supplier to limit the cost increase;
//   - if exact fulfillment is impossible, ok=false is returned.
//
// Offers must already be filtered to (approved supplier, in stock) and ordered
// by unit price ascending; tie order is left to the caller.
func planAllocation(offers []model.Offer, totalQuantity int) (lines []allocationLine, ok bool) {
	candidates := make([]allocationOffer, 0, len(offers))
	for _, offer := range offers {
		if offer.AvailableStock <= 0 || offer.MOQ <= 0 {
			continue
		}
		candidates = append(candidates, allocationOffer{
			OfferID:      offer.ID,
			SupplierID:   offer.SupplierID,
			SupplierName: offer.Supplier.Name,
			UnitPrice:    offer.UnitPrice,
			MOQ:          offer.MOQ,
			Available:    offer.AvailableStock,
			DeliveryDays: offer.DeliveryDays,
			Freight:      offer.Freight,
		})
	}

	remaining := totalQuantity
	for index := range candidates {
		if remaining == 0 {
			break
		}
		candidate := candidates[index]
		if candidate.Available < candidate.MOQ {
			continue
		}
		take := remaining
		if take > candidate.Available {
			take = candidate.Available
		}
		if take < candidate.MOQ {
			// take == remaining < MOQ: pull demand back from earlier suppliers
			// (down to their MOQ) so this supplier can receive a full MOQ.
			need := candidate.MOQ - take
			if !pullback(lines, need) {
				continue
			}
			remaining += need
			take = candidate.MOQ
		}
		lines = append(lines, allocationLine{
			Offer:      candidate,
			Quantity:   take,
			LineAmount: float64(take) * candidate.UnitPrice,
		})
		remaining -= take
	}
	return lines, remaining == 0
}

// pullback removes up to need units from already-allocated earlier suppliers,
// never below their MOQ. The most expensive prior supplier is reduced first so
// the (more expensive) MOQ top-up adds as little cost as possible.
func pullback(lines []allocationLine, need int) bool {
	if need <= 0 {
		return true
	}
	reducible := 0
	for i := range lines {
		reducible += lines[i].Quantity - lines[i].Offer.MOQ
	}
	if reducible < need {
		return false
	}
	for i := len(lines) - 1; i >= 0 && need > 0; i-- {
		room := lines[i].Quantity - lines[i].Offer.MOQ
		sub := room
		if sub > need {
			sub = need
		}
		lines[i].Quantity -= sub
		lines[i].LineAmount = float64(lines[i].Quantity) * lines[i].Offer.UnitPrice
		need -= sub
	}
	return need == 0
}
