package service

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/dto"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"github.com/blueship581/cybuildprice/backend/internal/repository"
)

type fakePurchaseOrderRepo struct {
	offers  []model.Offer
	orders  map[string]model.PurchaseOrder
	nextID  uint
	lastKey string
}

func newFakePurchaseOrderRepo(offers []model.Offer) *fakePurchaseOrderRepo {
	return &fakePurchaseOrderRepo{offers: offers, orders: map[string]model.PurchaseOrder{}, nextID: 1}
}

func (f *fakePurchaseOrderRepo) WithTx(fn func(repository.PurchaseOrderRepository) error) error {
	return fn(f)
}

func (f *fakePurchaseOrderRepo) FindByIdempotencyKey(userID, key string) (model.PurchaseOrder, error) {
	if order, ok := f.orders[userID+"/"+key]; ok {
		return order, nil
	}
	return model.PurchaseOrder{}, apperrors.ErrNotFound
}

func (f *fakePurchaseOrderRepo) LockEligibleOffers(uint) ([]model.Offer, error) {
	return f.offers, nil
}

func (f *fakePurchaseOrderRepo) CreateOrder(order *model.PurchaseOrder) error {
	if _, exists := f.orders[order.UserID+"/"+order.IdempotencyKey]; exists {
		return apperrors.ErrDuplicateKey
	}
	order.ID = f.nextID
	f.nextID++
	f.orders[order.UserID+"/"+order.IdempotencyKey] = *order
	f.lastKey = order.IdempotencyKey
	return nil
}

func (f *fakePurchaseOrderRepo) DecrementOfferStock(offerID uint, quantity int) error {
	for index := range f.offers {
		if f.offers[index].ID == offerID {
			if f.offers[index].AvailableQty < quantity {
				return apperrors.ErrStockChanged
			}
			f.offers[index].AvailableQty -= quantity
			return nil
		}
	}
	return apperrors.ErrNotFound
}

func (f *fakePurchaseOrderRepo) GetByID(id uint, userID string) (model.PurchaseOrder, error) {
	for _, order := range f.orders {
		if order.ID == id && order.UserID == userID {
			return order, nil
		}
	}
	return model.PurchaseOrder{}, apperrors.ErrNotFound
}

func (f *fakePurchaseOrderRepo) ListByUser(userID string) ([]model.PurchaseOrder, error) {
	var orders []model.PurchaseOrder
	for _, order := range f.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

type fakePurchaseProductRepo struct{ product model.Product }

func (f fakePurchaseProductRepo) List(string, string, string, int, int) ([]model.Product, int64, error) {
	return nil, 0, nil
}
func (f fakePurchaseProductRepo) Get(id uint) (model.Product, error) {
	if f.product.ID == id {
		return f.product, nil
	}
	return model.Product{}, apperrors.ErrNotFound
}
func (f fakePurchaseProductRepo) Compare([]uint) ([]model.Product, error) { return nil, nil }

func newPurchaseService(offers []model.Offer) (*PurchaseOrderService, *fakePurchaseOrderRepo) {
	repo := newFakePurchaseOrderRepo(offers)
	product := model.Product{Name: "云纹岩板", Unit: "片"}
	product.ID = 7
	svc := NewPurchaseOrderService(repo, fakePurchaseProductRepo{product}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	return svc, repo
}

func purchaseRequest(quantity int, key string) dto.CreatePurchaseOrderRequest {
	return dto.CreatePurchaseOrderRequest{ProductID: 7, Quantity: quantity, IdempotencyKey: key}
}

func TestPurchaseOrderPlaceConfirmedSplit(t *testing.T) {
	svc, repo := newPurchaseService([]model.Offer{allocOffer(1, 398, 10, 50), allocOffer(2, 412, 5, 30)})
	view, err := svc.Place("demo-user", purchaseRequest(60, "key-confirmed-1"))
	if err != nil {
		t.Fatal(err)
	}
	if view.Status != constants.PurchaseOrderConfirmed || len(view.Items) != 2 {
		t.Fatalf("unexpected view: %+v", view)
	}
	if view.Items[0].Quantity != 50 || view.Items[1].Quantity != 10 {
		t.Fatalf("unexpected split: %+v", view.Items)
	}
	if view.TotalAmount != 50*398+10*412 {
		t.Fatalf("unexpected total: %v", view.TotalAmount)
	}
	if repo.offers[0].AvailableQty != 0 || repo.offers[1].AvailableQty != 20 {
		t.Fatalf("stock not deducted: %+v", repo.offers)
	}
}

func TestPurchaseOrderPlaceRejectedKeepsStock(t *testing.T) {
	svc, repo := newPurchaseService([]model.Offer{allocOffer(1, 398, 10, 50), allocOffer(2, 412, 5, 30)})
	view, err := svc.Place("demo-user", purchaseRequest(100, "key-rejected-1"))
	if err != nil {
		t.Fatal(err)
	}
	if view.Status != constants.PurchaseOrderRejected || view.FailureReason == "" {
		t.Fatalf("unexpected view: %+v", view)
	}
	if len(view.Items) != 0 || view.TotalAmount != 0 {
		t.Fatalf("rejected order must not carry allocations: %+v", view)
	}
	if repo.offers[0].AvailableQty != 50 || repo.offers[1].AvailableQty != 30 {
		t.Fatalf("rejected order must not change stock: %+v", repo.offers)
	}
}

func TestPurchaseOrderIdempotentReplay(t *testing.T) {
	svc, repo := newPurchaseService([]model.Offer{allocOffer(1, 398, 10, 50)})
	first, err := svc.Place("demo-user", purchaseRequest(20, "key-replay-1"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.Place("demo-user", purchaseRequest(20, "key-replay-1"))
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || second.Status != constants.PurchaseOrderConfirmed {
		t.Fatalf("replay must return the original order: %+v vs %+v", first, second)
	}
	if repo.offers[0].AvailableQty != 30 {
		t.Fatalf("stock must be deducted once, got %d", repo.offers[0].AvailableQty)
	}
}

func TestPurchaseOrderRejectedWhenEligibleStockExhausted(t *testing.T) {
	svc, repo := newPurchaseService([]model.Offer{allocOffer(1, 398, 10, 0), allocOffer(2, 412, 5, 0)})
	view, err := svc.Place("demo-user", purchaseRequest(10, "key-exhausted-1"))
	if err != nil {
		t.Fatal(err)
	}
	if view.Status != constants.PurchaseOrderRejected || len(view.Items) != 0 {
		t.Fatalf("exhausted stock must reject without items: %+v", view)
	}
	if view.FailureReason == constants.RejectMessageNoOffers {
		t.Fatalf("exhausted eligible offers must not be reported as no offers: %s", view.FailureReason)
	}
	if repo.offers[0].AvailableQty != 0 {
		t.Fatalf("rejected order must not change stock: %d", repo.offers[0].AvailableQty)
	}
}

func TestPurchaseOrderProductMissing(t *testing.T) {
	svc, _ := newPurchaseService(nil)
	missing := dto.CreatePurchaseOrderRequest{ProductID: 99, Quantity: 10, IdempotencyKey: "key-missing-1"}
	_, err := svc.Place("demo-user", missing)
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
