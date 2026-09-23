import type { ReactNode } from 'react';
import { useLocation } from 'react-router';
import { Bell } from '@phosphor-icons/react';
import { Button } from '@/components/ui/button';

interface TopAppBarProps {
  title?: string;
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

export function TopAppBar({ title, actions }: TopAppBarProps) {
  const location = useLocation();
  const displayTitle = title ?? resolvePageTitle(location.pathname);

  return (
    <header className="flex h-11 shrink-0 items-center justify-between border-b border-border bg-background/95 px-4 backdrop-blur">
      <h1 className="text-sm font-semibold text-foreground tracking-tight">{displayTitle}</h1>

      <div className="flex items-center gap-1.5">
        {actions}
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label="Notifications"
          className="relative size-8 text-muted-foreground hover:text-foreground"
        >
          <Bell className="size-4.5" />
          {/* Subtle unread notification indicator dot */}
          <span
            data-slot="notification-dot"
            aria-hidden="true"
            className="absolute top-1.5 right-1.5 size-2 rounded-full bg-primary ring-2 ring-background"
          />
        </Button>
      </div>
    </header>
  );
}
