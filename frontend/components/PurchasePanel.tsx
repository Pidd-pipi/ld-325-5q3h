'use client';

import { useCallback, useEffect, useState } from 'react';
import { History, PackageCheck, PackageX, RefreshCw, SplitSquareHorizontal } from 'lucide-react';

import { api } from '@/lib/api';
import type { Product, PurchaseOrder } from '@/lib/types';
import { money } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

interface PurchasePanelProps {
  product: Product | null;
  onNotify: (text: string) => void;
}

export function PurchasePanel({ product, onNotify }: PurchasePanelProps) {
  const [quantity, setQuantity] = useState(10);
  const [draftKey, setDraftKey] = useState('');
  const [pending, setPending] = useState(false);
  const [active, setActive] = useState<PurchaseOrder | null>(null);
  const [orders, setOrders] = useState<PurchaseOrder[]>([]);

  const productID = product?.ID;
  useEffect(() => { setDraftKey(''); setActive(null); }, [productID, quantity]);

  const reload = useCallback(() => { api.listPurchaseOrders().then(setOrders).catch(() => {}); }, []);
  useEffect(() => { reload(); }, [reload]);

  if (!product) return null;

  const allocatable = (product.Offers || []).filter((offer) => offer.StockStatus === 'in_stock' && offer.Supplier?.Status === 'approved');
  const totalAvailable = allocatable.reduce((sum, offer) => sum + offer.AvailableQty, 0);

  const submit = async () => {
    if (pending || quantity < 1) return;
    const key = draftKey || crypto.randomUUID();
    setDraftKey(key);
    setPending(true);
    try {
      const order = await api.createPurchaseOrder(product.ID, quantity, key);
      setActive(order);
      reload();
      onNotify(order.status === 'confirmed' ? `采购单 #${order.id} 已确认，库存占用已生效` : `采购单 #${order.id} 被整单拒绝，库存未变化`);
    } catch {
      onNotify('采购单提交失败，请稍后重试');
    } finally {
      setPending(false);
    }
  };

  const readBack = async (id: number) => {
    try { setActive(await api.getPurchaseOrder(id)); } catch { onNotify('采购单详情读取失败'); }
  };

  return (
    <section className="purchase" id="purchase">
      <div className="section-title">
        <div>
          <p className="eyebrow">SPLIT PROCUREMENT / 采购拆单</p>
          <h2>填一个总量，<em>系统替你拆到多家店铺。</em></h2>
        </div>
        <p>只在已审核商家的有货报价间分配，逐笔满足起订量且不超可供量；总量不足时整单拒绝，采购单、明细与剩余量都不变。</p>
      </div>
      <div className="purchase-layout">
        <div className="purchase-form">
          <div className="calc-heading"><SplitSquareHorizontal size={20} /><b>新建采购单 · {product.Name}</b></div>
          <div className="purchase-stock">
            <div><span>已审核 · 有货报价</span><strong>{allocatable.length} 家</strong></div>
            <div><span>当前可供总量</span><strong>{totalAvailable} {product.Unit}</strong></div>
          </div>
          <label>需求总量（{product.Unit}）
            <input type="number" min="1" value={quantity} onChange={(event) => setQuantity(Math.max(1, Number(event.target.value) || 1))} />
          </label>
          <Button onClick={submit} disabled={pending}>{pending ? '正在拆单…' : '生成采购单'}</Button>
          <p className="purchase-hint">同一需求重复提交（如网络重试、双击）会通过幂等键返回原采购单，不会重复占用库存。</p>
        </div>
        {active && (
          <div className={`purchase-result ${active.status}`}>
            <div className="purchase-result-head">
              <b>采购单 #{active.id}</b>
              {active.status === 'confirmed'
                ? <Badge tone="good"><PackageCheck size={11} /> 已确认</Badge>
                : <Badge tone="alert"><PackageX size={11} /> 已拒绝</Badge>}
            </div>
            {active.status === 'rejected' && <p className="purchase-failure">{active.failure_reason}</p>}
            {active.items.length > 0 && (
              <div className="purchase-items">
                {active.items.map((item) => (
                  <div className="purchase-item" key={item.offer_id}>
                    <b>{item.supplier_name}</b>
                    <span>{item.quantity} {active.unit} × {money(item.unit_price)}</span>
                    <span>{item.delivery_days} 天交货</span>
                    <strong>{money(item.amount)}</strong>
                  </div>
                ))}
                <div className="purchase-total">
                  <span>合计 {active.quantity} {active.unit} · 整单交期 {active.delivery_days} 天</span>
                  <strong>{money(active.total_amount)}</strong>
                </div>
              </div>
            )}
            <p className="purchase-key">幂等键 {active.idempotency_key.slice(0, 18)}… · 重复提交返回本单</p>
          </div>
        )}
      </div>
      <div className="purchase-history">
        <div className="offer-title">
          <b><History size={14} /> 我的采购单</b>
          <button className="purchase-refresh" onClick={reload} aria-label="刷新采购单列表"><RefreshCw size={13} /> 刷新回读</button>
        </div>
        {orders.length === 0 && <p className="purchase-empty">还没有采购单，从上方填写总量开始。</p>}
        {orders.map((order) => (
          <button className="purchase-row" key={order.id} onClick={() => readBack(order.id)}>
            <span>#{order.id}</span>
            <b>{order.product_name}</b>
            <span>{order.quantity} {order.unit}</span>
            <span className={order.status === 'confirmed' ? 'row-ok' : 'row-bad'}>{order.status === 'confirmed' ? '已确认' : '已拒绝'}</span>
            <span>{order.status === 'confirmed' ? money(order.total_amount) : (order.failure_reason || '').slice(0, 16) + '…'}</span>
          </button>
        ))}
      </div>
    </section>
  );
}
