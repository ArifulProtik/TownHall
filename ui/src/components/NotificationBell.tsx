import { useState, type Ref } from 'react';
import { Bell } from '@phosphor-icons/react';
import { cn } from 'cn';

interface NotificationBellProps {
  hasUnread?: boolean;
  unreadCount?: number;
  onClick?: () => void;
  ref?: Ref<HTMLButtonElement>;
}

/**
 * Rail-styled notification button for the primary sidebar: same squircle
 * treatment as SidebarRailItem, with an unread count badge. Rendered above
 * the profile avatar at the pinned bottom of the rail.
 * Forwards ref so it can serve as a Base UI Popover trigger.
 */
export function NotificationBell({ hasUnread = true, unreadCount = 0, onClick, ref }: NotificationBellProps) {
  const [isHovered, setIsHovered] = useState(false);
  const showCount = hasUnread && unreadCount > 0;

  return (
    <div
      className="group relative flex w-full justify-center py-1"
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <button
        ref={ref}
        type="button"
        aria-label="Notifications"
        onClick={onClick}
        className="relative flex size-12 cursor-pointer items-center justify-center rounded-[24px] bg-secondary/80 text-secondary-foreground transition-all duration-200 ease-out select-none hover:rounded-[16px] hover:bg-primary hover:text-primary-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring"
      >
        <span className="flex size-6 items-center justify-center [&>svg]:size-6">
          <Bell />
        </span>
        {hasUnread && !showCount && (
          <span
            data-slot="notification-dot"
            aria-hidden="true"
            className="absolute top-1.5 right-1.5 size-2 rounded-full bg-primary ring-2 ring-sidebar group-hover:bg-primary-foreground"
          />
        )}
        {showCount && (
          <span
            data-slot="notification-count"
            aria-hidden="true"
            className="absolute -top-0.5 -right-0.5 flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1 text-[11px] font-bold text-primary-foreground ring-2 ring-sidebar"
          >
            {unreadCount > 99 ? '99+' : unreadCount}
          </span>
        )}
      </button>

      {/* Floating Tooltip Label */}
      <div
        role="tooltip"
        className={cn(
          'pointer-events-none absolute left-[78px] top-1/2 z-50 -translate-y-1/2 rounded-md border border-border bg-popover px-2.5 py-1 text-xs font-semibold whitespace-nowrap text-popover-foreground shadow-md transition-all duration-150 ease-out',
          isHovered
            ? 'visible translate-x-0 opacity-100'
            : 'invisible -translate-x-1 opacity-0'
        )}
      >
        Notifications
        <span className="absolute -left-1 top-1/2 size-2 -translate-y-1/2 rotate-45 border-b border-l border-border bg-popover" />
      </div>
    </div>
  );
}
