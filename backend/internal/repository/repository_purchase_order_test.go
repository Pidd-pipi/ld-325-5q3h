package repository

import (
	"errors"
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	apperrors "github.com/blueship581/cybuildprice/backend/internal/errors"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPurchaseDB(t *testing.T) (*gorm.DB, []model.Offer) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = model.Migrate(db); err != nil {
		t.Fatal(err)
	}
	approved := model.Supplier{Name: "已审核商家", Status: constants.SupplierApproved}
	pending := model.Supplier{Name: "待审核商家", Status: constants.SupplierPending}
	db.Create(&approved)
	db.Create(&pending)
	product := model.Product{Name: "岩板", Unit: "片"}
	db.Create(&product)
	offers := []model.Offer{
		{ProductID: product.ID, SupplierID: approved.ID, UnitPrice: 412, MOQ: 5, StockStatus: constants.StatusInStock, AvailableQty: 30},
		{ProductID: product.ID, SupplierID: approved.ID, UnitPrice: 398, MOQ: 10, StockStatus: constants.StatusInStock, AvailableQty: 50},
		{ProductID: product.ID, SupplierID: approved.ID, UnitPrice: 300, MOQ: 1, StockStatus: constants.StatusOutOfStock, AvailableQty: 0},
		{ProductID: product.ID, SupplierID: pending.ID, UnitPrice: 100, MOQ: 1, StockStatus: constants.StatusInStock, AvailableQty: 999},
	}
	for index := range offers {
		db.Create(&offers[index])
	}
	return db, offers
}

func TestLockEligibleOffersFiltersAndSorts(t *testing.T) {
	db, offers := setupPurchaseDB(t)
	repo := NewPurchaseOrderRepository(db)
	locked, err := repo.LockEligibleOffers(offers[0].ProductID)
	if err != nil {
		t.Fatal(err)
	}
	if len(locked) != 2 {
		t.Fatalf("only approved in-stock offers are eligible, got %d", len(locked))
	}
	if locked[0].UnitPrice != 398 || locked[1].UnitPrice != 412 {
		t.Fatalf("offers must be ordered by price, got %v then %v", locked[0].UnitPrice, locked[1].UnitPrice)
	}
	if locked[0].Supplier.Name != "已审核商家" {
		t.Fatalf("supplier must be preloaded, got %+v", locked[0].Supplier)
	}
}

func TestPurchaseOrderCreateAndReadBack(t *testing.T) {
	db, offers := setupPurchaseDB(t)
	repo := NewPurchaseOrderRepository(db)
	order := model.PurchaseOrder{
		UserID: "demo-user", ProductID: offers[0].ProductID, Quantity: 60,
		Status: constants.PurchaseOrderConfirmed, IdempotencyKey: "key-readback-1",
		Items: []model.PurchaseOrderItem{
			{OfferID: offers[1].ID, SupplierID: offers[1].SupplierID, Quantity: 50, UnitPrice: 398, Amount: 19900, DeliveryDays: 3},
			{OfferID: offers[0].ID, SupplierID: offers[0].SupplierID, Quantity: 10, UnitPrice: 412, Amount: 4120, DeliveryDays: 2},
		},
	}
	if err := repo.CreateOrder(&order); err != nil {
		t.Fatal(err)
	}
	if order.ID == 0 || len(order.Items) == 0 || order.Items[0].ID == 0 {
		t.Fatalf("order and items must be persisted: %+v", order)
	}
	loaded, err := repo.GetByID(order.ID, "demo-user")
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Items) != 2 || loaded.Items[0].Supplier.Name != "已审核商家" {
		t.Fatalf("read-back must include allocations and suppliers: %+v", loaded)
	}
	if _, err = repo.GetByID(order.ID, "other-user"); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("orders of other users must not be readable, got %v", err)
	}
	replayed, err := repo.FindByIdempotencyKey("demo-user", "key-readback-1")
	if err != nil || replayed.ID != order.ID {
		t.Fatalf("idempotency lookup must return the stored order: %v", err)
	}
}

func TestPurchaseOrderDuplicateKeyRejected(t *testing.T) {
	db, offers := setupPurchaseDB(t)
	repo := NewPurchaseOrderRepository(db)
	order := model.PurchaseOrder{UserID: "demo-user", ProductID: offers[0].ProductID, Quantity: 5, Status: constants.PurchaseOrderConfirmed, IdempotencyKey: "key-dupe-1"}
	if err := repo.CreateOrder(&order); err != nil {
		t.Fatal(err)
	}
	dupe := model.PurchaseOrder{UserID: "demo-user", ProductID: offers[0].ProductID, Quantity: 5, Status: constants.PurchaseOrderConfirmed, IdempotencyKey: "key-dupe-1"}
	if err := repo.CreateOrder(&dupe); !errors.Is(err, apperrors.ErrDuplicateKey) {
		t.Fatalf("duplicate key must be reported, got %v", err)
	}
}

func TestDecrementOfferStockGuard(t *testing.T) {
	db, offers := setupPurchaseDB(t)
	repo := NewPurchaseOrderRepository(db)
	if err := repo.DecrementOfferStock(offers[1].ID, 20); err != nil {
		t.Fatal(err)
	}
	var reloaded model.Offer
	db.First(&reloaded, offers[1].ID)
	if reloaded.AvailableQty != 30 {
		t.Fatalf("available qty = %d, want 30", reloaded.AvailableQty)
	}
	if err := repo.DecrementOfferStock(offers[1].ID, 40); !errors.Is(err, apperrors.ErrStockChanged) {
		t.Fatalf("overdraft must be rejected, got %v", err)
	}
	db.First(&reloaded, offers[1].ID)
	if reloaded.AvailableQty != 30 {
		t.Fatalf("failed decrement must not change stock, got %d", reloaded.AvailableQty)
	}
}
