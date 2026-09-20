import type { ReactNode } from 'react';

export function TopAppBar({ title, actions }: { title: string; actions?: ReactNode }) {
  return (
    <header className="flex h-14 items-center justify-between bg-surface px-4">
      <h1 className="text-base font-medium text-on-surface">{title}</h1>
      <div className="flex items-center gap-2">{actions}</div>
    </header>
  );
}
