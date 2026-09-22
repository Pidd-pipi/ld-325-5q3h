'use client';
import { useMemo, useState } from 'react';
import { ClipboardList, RefreshCw, ShieldX, Split } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { api } from '@/lib/api';
import type { Product, PurchaseOrderResult } from '@/lib/types';
import { money } from '@/lib/utils';
import { PurchaseAllocationTable } from './PurchaseAllocationTable';

const newIdempotencyKey = () => `web-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;

export function PurchasePanel({ product }: { product: Product | null }) {
  const [quantity, setQuantity] = useState(10);
  const [idempotencyKey, setIdempotencyKey] = useState(newIdempotencyKey);
  const [pending, setPending] = useState(false);
  const [result, setResult] = useState<PurchaseOrderResult | null>(null);
  const [refreshError, setRefreshError] = useState('');

  const eligibleOffers = useMemo(
    () => (product?.Offers ?? []).filter((offer) => offer.StockStatus === 'in_stock' && offer.Supplier.Status === 'approved'),
    [product],
  );
  const totalAvailable = eligibleOffers.reduce((sum, offer) => sum + Math.max(offer.AvailableStock, 0), 0);

  const submit = async () => {
    if (!product) return;
    setPending(true);
    setRefreshError('');
    try {
      const outcome = await api.createPurchaseOrder(product.ID, quantity, idempotencyKey);
      setResult(outcome.data ?? null);
    } finally {
      setPending(false);
    }
  };

  const resubmit = () => {
    setIdempotencyKey(newIdempotencyKey());
    setResult(null);
    setRefreshError('');
  };

  const refreshDetail = async () => {
    if (!result) return;
    setRefreshError('');
    try {
      setResult(await api.getPurchaseOrder(result.id));
    } catch (error) {
      setRefreshError(error instanceof Error ? error.message : '回读失败');
    }
  };

  return (
    <section id="purchase" className="purchase">
      <div>
        <p className="eyebrow">SPLIT PURCHASE / CLOSED LOOP</p>
        <h2>
          一个总量，<em>系统在有货报价间拆单。</em>
        </h2>
        <p>
          仅在已审核商家的有货报价中分配，满足各商家起订量且不超过可供量；单一商家不足时自动拆分。
          总量不足（含并发占用）时整单拒绝，采购单、明细与剩余量均不变。
        </p>
      </div>
      <div className="purchase-card">
        <div className="calc-heading">
          <Split size={20} />
          <b>建材采购拆单</b>
        </div>
        {!product ? (
          <p className="purchase-empty">请先在上方选择一款材料。</p>
        ) : (
          <>
            <label>
              采购材料
              <input value={`${product.Name}（${product.Brand}）`} readOnly />
            </label>
            <label>
              商品总量（{product.Unit}）
              <input
                type="number"
                min={1}
                value={quantity}
                onChange={(event) => setQuantity(Number(event.target.value))}
              />
            </label>
            <label>
              幂等键（重复提交返回原结果）
              <input value={idempotencyKey} onChange={(event) => setIdempotencyKey(event.target.value)} />
            </label>
            <div className="purchase-available">
              当前有货可分配总量：<strong>{totalAvailable}</strong> {product.Unit}
              （仅统计 {eligibleOffers.length} 家已审核商家的有货报价）
            </div>
            <div className="purchase-actions">
              <Button onClick={submit} disabled={pending || quantity <= 0 || idempotencyKey.trim().length < 8}>
                {pending ? '占用与拆单中…' : '提交采购单'}
              </Button>
              <button className="ghost-button" type="button" onClick={resubmit}>
                换幂等键重提
              </button>
            </div>
            {result && (
              <div className={`purchase-result ${result.status}`}>
                <div className="purchase-result-head">
                  {result.status === 'succeeded' ? (
                    <span className="result-title ok">
                      <ClipboardList size={16} /> 采购单 #{result.id} 已生效
                    </span>
                  ) : (
                    <span className="result-title fail">
                      <ShieldX size={16} /> 采购单 #{result.id} 整单拒绝
                    </span>
                  )}
                  <button className="ghost-button" type="button" onClick={refreshDetail}>
                    <RefreshCw size={13} /> 刷新回读
                  </button>
                </div>
                {result.status === 'succeeded' ? (
                  <>
                    <PurchaseAllocationTable items={result.items} unit={product.Unit} />
                    <div className="purchase-summary">
                      <span>合计 {result.total_quantity} {product.Unit} · 最长交期 {result.max_delivery_days} 天</span>
                      <strong>{money(result.total_amount)}</strong>
                    </div>
                  </>
                ) : (
                  <p className="purchase-failure">
                    失败原因（{result.failure_code}）：{result.failure_reason}
                    <br />
                    采购单、明细与商家剩余量均未发生变化；使用同一幂等键重试会返回本结果。
                  </p>
                )}
                {refreshError && <p className="purchase-error">{refreshError}</p>}
              </div>
            )}
          </>
        )}
      </div>
    </section>
  );
}
