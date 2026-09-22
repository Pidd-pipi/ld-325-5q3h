'use client';

import { ArrowRight, X } from 'lucide-react';

import type { Product } from '@/lib/types';
import { money } from '@/lib/utils';

interface ComparisonTrayProps {
  items: Product[];
  onRemove: (id: number) => void;
}

export function ComparisonTray({ items, onRemove }: ComparisonTrayProps) {
  if (!items.length) {
    return (
      <aside className="compare-tray empty" id="compare">
        <p>
          <b>对比清单</b> · 最多选 4 款材料
        </p>
        <span>从材料卡片中加入，系统会标出当前最低报价。</span>
      </aside>
    );
  }

  return (
    <aside className="compare-tray" id="compare">
      <div>
        <p className="eyebrow">COMPARE LIST · {items.length}/4</p>
        <h2>报价并排看，差异不靠猜。</h2>
      </div>
      <div className="compare-items">
        {items.map((item) => {
          const price = Math.min(...item.Offers.map((offer) => offer.UnitPrice));
          return (
            <div className="compare-item" key={item.ID}>
              <button onClick={() => onRemove(item.ID)} aria-label="移出对比">
                <X size={15} />
              </button>
              <span>{item.Brand}</span>
              <b>{item.Name}</b>
              <strong>{money(price)} 起</strong>
            </div>
          );
        })}
        <div className="compare-cta">
          最低报价将被标红
          <br />
          <a href="#catalog">
            查看商家报价 <ArrowRight size={15} />
          </a>
        </div>
      </div>
    </aside>
  );
}
