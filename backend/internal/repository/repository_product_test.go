package repository

import (
	"github.com/blueship581/cybuildprice/backend/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func TestProductRepositoryList(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err = model.Migrate(db); err != nil {
		t.Fatal(err)
	}
	category := model.Category{Name: "瓷砖"}
	db.Create(&category)
	db.Create(&model.Product{Name: "岩板", Brand: "石界", CategoryID: category.ID})
	repo := NewProductRepository(db)
	got, total, err := repo.List("岩", "", "", 1, 12)
	if err != nil || total != 1 || len(got) != 1 {
		t.Fatalf("got %d/%d, err %v", len(got), total, err)
	}
}
