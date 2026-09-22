export type Supplier = { ID: number; Name: string; Rating: number; Address: string; Status: string };
export type Offer = { ID: number; UnitPrice: number; MOQ: number; Freight: string; DeliveryDays: number; StockStatus: string; AvailableQty: number; Supplier: Supplier };
export type Product = { ID: number; Name: string; Brand: string; Model: string; Unit: string; Thumbnail: string; SalesCount: number; Rating: number; Category: { Name: string }; Offers: Offer[] };
export type ApiEnvelope<T> = { code: number; message: string; data: T };
export type TrendPoint = { Price: number; RecordedAt: string };
export type Trend = { range: string; highest: number; lowest: number; average: number; points: TrendPoint[] };
export type PurchaseOrderItem = { offer_id: number; supplier_id: number; supplier_name: string; quantity: number; unit_price: number; amount: number; delivery_days: number };
export type PurchaseOrder = { id: number; product_id: number; product_name: string; unit: string; quantity: number; status: 'confirmed' | 'rejected'; failure_reason?: string; total_amount: number; delivery_days: number; idempotency_key: string; created_at: string; items: PurchaseOrderItem[] };
