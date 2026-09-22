export type Supplier = { ID: number; Name: string; Rating: number; Address: string; Status: string };
export type Offer = { ID: number; UnitPrice: number; MOQ: number; Freight: string; DeliveryDays: number; StockStatus: string; AvailableStock: number; Supplier: Supplier };
export type Product = { ID: number; Name: string; Brand: string; Model: string; Unit: string; Thumbnail: string; SalesCount: number; Rating: number; Category: { Name: string }; Offers: Offer[] };
export type ApiEnvelope<T> = { code: number; message: string; data: T };
export type TrendPoint = { Price: number; RecordedAt: string };
export type Trend = { range: string; highest: number; lowest: number; average: number; points: TrendPoint[] };

export type PurchaseAllocationItem = {
  offer_id: number;
  supplier_id: number;
  supplier_name: string;
  unit_price: number;
  quantity: number;
  moq: number;
  available_before: number;
  available_after: number;
  delivery_days: number;
  freight: string;
  line_amount: number;
};

export type PurchaseOrderResult = {
  id: number;
  product_id: number;
  product_name: string;
  product_unit: string;
  total_quantity: number;
  status: 'succeeded' | 'failed';
  total_amount: number;
  max_delivery_days: number;
  failure_code?: string;
  failure_reason?: string;
  items: PurchaseAllocationItem[];
  created_at: string;
};
