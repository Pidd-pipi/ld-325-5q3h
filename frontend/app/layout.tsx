import type { Metadata } from 'next';
import './globals.css';
export const metadata: Metadata = { title: '筑价 · 建材比价平台', description: '面向装修业主和施工队的建材价格对比平台' };
export default function RootLayout({ children }: Readonly<{children: React.ReactNode}>) { return <html lang="zh-CN"><body>{children}</body></html>; }
