import type { PropsWithChildren } from 'react';
export function Badge({ children, tone = 'neutral' }: PropsWithChildren<{tone?: 'neutral'|'good'|'alert'}>) { return <span className={`badge ${tone}`}>{children}</span>; }
