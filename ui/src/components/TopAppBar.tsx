import type { ReactNode } from 'react';
import { useLocation } from 'react-router';

interface TopAppBarProps {
  title?: string;
  subtitle?: string;
  actions?: ReactNode;
}

function resolvePageTitle(pathname: string): string {
  if (pathname === '/' || pathname === '') {
    return 'Home';
  }
  if (pathname.startsWith('/dms')) {
    return 'Direct Messages';
  }
  if (pathname.startsWith('/spaces')) {
    return 'Spaces';
  }
  if (pathname.startsWith('/feed')) {
    return 'Feed';
  }
  if (pathname.startsWith('/u/')) {
    return 'Profile';
  }
  // Fallback: capitalize path segment
  const segment = pathname.split('/').filter(Boolean)[0];
  return segment ? segment.charAt(0).toUpperCase() + segment.slice(1) : 'TownHall';
}

export function TopAppBar({ title, subtitle, actions }: TopAppBarProps) {
  const location = useLocation();
  const displayTitle = title ?? resolvePageTitle(location.pathname);

  return (
    <header className="flex h-11 shrink-0 items-center justify-between border-b border-border bg-background/95 px-4 backdrop-blur">
      <div className="flex min-w-0 items-baseline gap-2">
        <h1 className="shrink-0 text-sm font-semibold text-foreground tracking-tight">
          {displayTitle}
        </h1>
        {subtitle && (
          <span className="truncate text-xs font-normal text-muted-foreground">
            {subtitle}
          </span>
        )}
      </div>

      <div className="flex items-center gap-1.5">
        {actions}
      </div>
    </header>
  );
}
