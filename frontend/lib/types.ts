export type Supplier = { ID: number; Name: string; Rating: number; Address: string; Status: string };
export type Offer = { ID: number; UnitPrice: number; MOQ: number; Freight: string; DeliveryDays: number; StockStatus: string; Supplier: Supplier };
export type Product = { ID: number; Name: string; Brand: string; Model: string; Unit: string; Thumbnail: string; SalesCount: number; Rating: number; Category: { Name: string }; Offers: Offer[] };
export type ApiEnvelope<T> = { code: number; message: string; data: T };
export type TrendPoint = { Price: number; RecordedAt: string };
export type Trend = { range: string; highest: number; lowest: number; average: number; points: TrendPoint[] };
