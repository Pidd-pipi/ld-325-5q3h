'use client';
import { Truck } from 'lucide-react';
import type { PurchaseAllocationItem } from '@/lib/types';
import { money } from '@/lib/utils';

export function PurchaseAllocationTable({ items, unit }: { items: PurchaseAllocationItem[]; unit: string }) {
  return (
    <div className="allocation-table">
      <div className="allocation-row allocation-head">
        <span>商家 / 报价</span>
        <span>分配数量</span>
        <span>剩余量</span>
        <span>交期</span>
        <span>小计</span>
      </div>
      {items.map((item) => (
        <div className="allocation-row" key={item.offer_id}>
          <span className="allocation-merchant">
            <b>{item.supplier_name}</b>
            <small>
              {money(item.unit_price)}/{unit} · 起订 {item.moq} · {item.freight}
            </small>
          </span>
          <span>
            {item.quantity} {unit}
          </span>
          <span className="allocation-stock">
            {item.available_before} → {item.available_after}
          </span>
          <span className="delivery">
            <Truck size={12} /> {item.delivery_days} 天
          </span>
          <span>{money(item.line_amount)}</span>
        </div>
      ))}
    </div>
  );
}
