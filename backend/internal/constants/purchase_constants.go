package constants

const (
	PurchaseStatusSucceeded = "succeeded"
	PurchaseStatusFailed    = "failed"

	PurchaseOrderPath         = "/purchase-orders"
	PurchaseOrderDetailPath   = "/purchase-orders/:id"

	PurchaseFailureStockChanged   = "stock_changed"
	PurchaseFailureNoAvailable    = "no_available_offer"
	PurchaseFailureInsufficient   = "insufficient_supply"

	PurchaseReasonProductNotFound       = "商品不存在"
	PurchaseReasonNoInStockOffer        = "没有已审核商家的有货报价"
	PurchaseReasonBelowMOQ              = "可供量不满足起订量"
	PurchaseReasonInsufficientTemplate  = "满足起订量的有货报价总量 %d，不足需求 %d"
	PurchaseReasonConcurrentTemplate    = "并发占用导致可供量变化，剩余总量不足 %d"
	PurchaseReasonQuantityChanged       = "可供量在下单过程中发生变化，请重试"
)
