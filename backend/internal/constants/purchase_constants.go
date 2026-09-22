package constants

const (
	PurchaseOrderConfirmed = "confirmed"
	PurchaseOrderRejected  = "rejected"

	RejectNone              = ""
	RejectNoAvailableOffers = "no_available_offers"
	RejectInsufficientStock = "insufficient_stock"
	RejectMOQUnfulfillable  = "moq_unfulfillable"

	RejectMessageNoOffers     = "没有已审核商家的有货报价，整单已拒绝"
	RejectMessageInsufficient = "已审核商家可供总量 %d %s，不足以覆盖需求 %d %s，整单已拒绝"
	RejectMessageMOQ          = "拆分后的剩余数量低于所有可用报价的起订量，整单已拒绝"

	PurchaseOrderListLimit = 50
	PostgresDialect        = "postgres"
)
