package model

import (
	"fmt"
	"github.com/blueship581/cybuildprice/backend/internal/constants"
	"gorm.io/gorm"
	"time"
)

func Seed(db *gorm.DB) error {
	var count int64
	if err := db.Model(&Product{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count seed products: %w", err)
	}
	if count > 0 {
		return nil
	}
	categories := []Category{}
	for index, name := range constants.ProductCategories {
		categories = append(categories, Category{Name: name, SortOrder: index + 1})
	}
	if err := db.Create(&categories).Error; err != nil {
		return fmt.Errorf("seed categories: %w", err)
	}
	suppliers := []Supplier{{Name: "筑家优选旗舰店", Address: "上海市浦东新区建材路 88 号", Contact: "400-820-6801", Rating: 4.9, Status: constants.SupplierApproved, Qualification: "品牌授权与质检报告齐全"}, {Name: "森木地板仓", Address: "杭州市余杭区良渚大道 16 号", Contact: "400-889-2186", Rating: 4.7, Status: constants.SupplierApproved, Qualification: "仓储直供"}, {Name: "匠心卫浴馆", Address: "南京市江宁区天印大道 36 号", Contact: "400-866-3019", Rating: 4.6, Status: constants.SupplierPending, Qualification: "等待平台复核"}}
	if err := db.Create(&suppliers).Error; err != nil {
		return fmt.Errorf("seed suppliers: %w", err)
	}
	products := []Product{{Name: "云纹岩板 900×1800", Brand: "石界", ModelNumber: "YB-918-W", Unit: "片", Thumbnail: "slab", CategoryID: categories[0].ID, SalesCount: 2380, Rating: 4.8}, {Name: "橡木多层地板 15mm", Brand: "森禾", ModelNumber: "SM-15-O", Unit: "㎡", Thumbnail: "floor", CategoryID: categories[1].ID, SalesCount: 1650, Rating: 4.7}, {Name: "净味内墙乳胶漆 18L", Brand: "森氧", ModelNumber: "SY-NW-18", Unit: "桶", Thumbnail: "paint", CategoryID: categories[2].ID, SalesCount: 3120, Rating: 4.9}, {Name: "智能恒温花洒套装", Brand: "沐光", ModelNumber: "MG-T90", Unit: "套", Thumbnail: "bath", CategoryID: categories[3].ID, SalesCount: 920, Rating: 4.6}, {Name: "304 不锈钢门锁", Brand: "启合", ModelNumber: "QH-304-L", Unit: "把", Thumbnail: "hardware", CategoryID: categories[4].ID, SalesCount: 1180, Rating: 4.5}, {Name: "断桥铝平开窗", Brand: "清朗", ModelNumber: "QL-70-C", Unit: "㎡", Thumbnail: "window", CategoryID: categories[5].ID, SalesCount: 760, Rating: 4.7}}
	if err := db.Create(&products).Error; err != nil {
		return fmt.Errorf("seed products: %w", err)
	}
	offers := []Offer{{ProductID: products[0].ID, SupplierID: suppliers[0].ID, UnitPrice: 398, MOQ: 10, Freight: "满 30 片送货上门", DeliveryDays: 3, StockStatus: constants.StatusInStock}, {ProductID: products[0].ID, SupplierID: suppliers[1].ID, UnitPrice: 412, MOQ: 5, Freight: "同城配送 80 元", DeliveryDays: 2, StockStatus: constants.StatusInStock}, {ProductID: products[1].ID, SupplierID: suppliers[1].ID, UnitPrice: 188, MOQ: 20, Freight: "满 100 ㎡免运费", DeliveryDays: 5, StockStatus: constants.StatusInStock}, {ProductID: products[1].ID, SupplierID: suppliers[0].ID, UnitPrice: 205, MOQ: 10, Freight: "物流到楼下", DeliveryDays: 4, StockStatus: constants.StatusOutOfStock}, {ProductID: products[2].ID, SupplierID: suppliers[0].ID, UnitPrice: 538, MOQ: 2, Freight: "满 6 桶免运费", DeliveryDays: 2, StockStatus: constants.StatusInStock}, {ProductID: products[3].ID, SupplierID: suppliers[2].ID, UnitPrice: 1299, MOQ: 1, Freight: "包邮入户", DeliveryDays: 7, StockStatus: constants.StatusInStock}, {ProductID: products[4].ID, SupplierID: suppliers[0].ID, UnitPrice: 79, MOQ: 5, Freight: "满 500 元包邮", DeliveryDays: 2, StockStatus: constants.StatusInStock}, {ProductID: products[5].ID, SupplierID: suppliers[1].ID, UnitPrice: 680, MOQ: 8, Freight: "测量后报价", DeliveryDays: 12, StockStatus: constants.StatusInStock}}
	if err := db.Create(&offers).Error; err != nil {
		return fmt.Errorf("seed offers: %w", err)
	}
	now := time.Now().AddDate(0, 0, -(constants.HistorySeedDays - 1))
	for _, product := range products {
		base := float64(100 + product.ID*45)
		for day := 0; day < constants.HistorySeedDays; day++ {
			wave := float64((day%7)-3) * 1.8
			history := PriceHistory{ProductID: product.ID, OfferID: offers[0].ID, Price: base + wave, RecordedAt: now.AddDate(0, 0, day)}
			if err := db.Create(&history).Error; err != nil {
				return fmt.Errorf("seed price history: %w", err)
			}
		}
	}
	return nil
}
