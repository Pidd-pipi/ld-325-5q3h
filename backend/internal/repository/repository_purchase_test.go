package repository

import (
	"testing"

	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newPurchaseTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := model.Migrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func seedPurchaseScenario(t *testing.T, db *gorm.DB) (model.Product, model.Supplier, model.Supplier) {
	t.Helper()
	product := model.Product{Name: "岩板", Unit: "片"}
	approvedA := model.Supplier{Name: "甲商家", Status: constants.SupplierApproved}
	approvedB := model.Supplier{Name: "乙商家", Status: constants.SupplierApproved}
	pending := model.Supplier{Name: "待审商家", Status: constants.SupplierPending}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}
	for _, s := range []*model.Supplier{&approvedA, &approvedB, &pending} {
		if err := db.Create(s).Error; err != nil {
			t.Fatal(err)
		}
	}
	offers := []model.Offer{
		{ProductID: product.ID, SupplierID: approvedA.ID, UnitPrice: 10, MOQ: 5, AvailableStock: 10, StockStatus: constants.StatusInStock},
		{ProductID: product.ID, SupplierID: approvedB.ID, UnitPrice: 12, MOQ: 5, AvailableStock: 50, StockStatus: constants.StatusInStock},
		{ProductID: product.ID, SupplierID: pending.ID, UnitPrice: 8, MOQ: 1, AvailableStock: 999, StockStatus: constants.StatusInStock},
	}
	for i := range offers {
		if err := db.Create(&offers[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	return product, approvedA, approvedB
}

func testPlanner(offers []model.Offer, total int) (map[uint]int, bool) {
	type line struct {
		id  uint
		qty int
	}
	remaining := total
	var lines []line
	for _, offer := range offers {
		if remaining == 0 {
			break
		}
		if offer.AvailableStock < offer.MOQ {
			continue
		}
		take := remaining
		if take > offer.AvailableStock {
			take = offer.AvailableStock
		}
		if take < offer.MOQ {
			continue
		}
		lines = append(lines, line{offer.ID, take})
		remaining -= take
	}
	if remaining != 0 {
		return nil, false
	}
	out := map[uint]int{}
	for _, l := range lines {
		out[l.id] = l.qty
	}
	return out, true
}

func TestPurchasePlaceSuccessSplitsAndOccupies(t *testing.T) {
	db := newPurchaseTestDB(t)
	product, approvedA, approvedB := seedPurchaseScenario(t, db)
	repo := NewPurchaseRepository(db)

	order, rejected, err := repo.Place("user-1", "key-success", product.ID, 30, testPlanner)
	if err != nil || rejected {
		t.Fatalf("place: rejected=%v err=%v", rejected, err)
	}
	if order.Status != constants.PurchaseStatusSucceeded || len(order.Items) != 2 {
		t.Fatalf("unexpected order: %+v items=%d", order, len(order.Items))
	}
	var stockA, stockB int64
	db.Model(&model.Offer{}).Where("supplier_id = ?", approvedA.ID).Select("available_stock").Scan(&stockA)
	db.Model(&model.Offer{}).Where("supplier_id = ?", approvedB.ID).Select("available_stock").Scan(&stockB)
	if stockA != 0 || stockB != 30 {
		t.Fatalf("stock after order: A=%d B=%d", stockA, stockB)
	}
}

func TestPurchasePlaceRejectionLeavesStockUntouched(t *testing.T) {
	db := newPurchaseTestDB(t)
	product, _, _ := seedPurchaseScenario(t, db)
	repo := NewPurchaseRepository(db)

	order, rejected, err := repo.Place("user-1", "key-fail", product.ID, 200, testPlanner)
	if err != nil {
		t.Fatal(err)
	}
	if !rejected || order.Status != constants.PurchaseStatusFailed || order.FailureReason == "" {
		t.Fatalf("expected rejected failed order, got %+v", order)
	}
	if len(order.Items) != 0 {
		t.Fatalf("rejected order must have no items")
	}
	var totalStock int64
	db.Model(&model.Offer{}).Where("stock_status = ?", constants.StatusInStock).Select("COALESCE(SUM(available_stock),0)").Scan(&totalStock)
	if totalStock != 1059 {
		t.Fatalf("stock changed after rejection: %d", totalStock)
	}
}

func TestPurchaseIdempotencyReturnsOriginal(t *testing.T) {
	db := newPurchaseTestDB(t)
	product, _, _ := seedPurchaseScenario(t, db)
	repo := NewPurchaseRepository(db)

	first, rejected, err := repo.Place("user-1", "key-dup", product.ID, 30, testPlanner)
	if err != nil || rejected {
		t.Fatalf("first place: %v %v", rejected, err)
	}
	second, rejected, err := repo.Place("user-1", "key-dup", product.ID, 999, testPlanner)
	if err != nil || rejected {
		t.Fatalf("replay: rejected=%v err=%v", rejected, err)
	}
	if second.ID != first.ID || second.TotalQuantity != 30 || len(second.Items) != 2 {
		t.Fatalf("replay did not return original order: first=%d second=%+v", first.ID, second)
	}
	var totalStock int64
	db.Model(&model.Offer{}).Where("stock_status = ?", constants.StatusInStock).Select("COALESCE(SUM(available_stock),0)").Scan(&totalStock)
	if totalStock != 1029 {
		t.Fatalf("stock occupied twice: %d", totalStock)
	}
}
