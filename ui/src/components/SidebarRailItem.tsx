import { type ReactNode, useState } from 'react';
import { NavLink } from 'react-router';
import { cn } from 'cn';

export interface SidebarRailItemProps {
  to: string;
  label: string;
  icon?: ReactNode;
  initials?: string;
  badgeCount?: number;
  className?: string;
}

export function SidebarRailItem({
  to,
  label,
  icon,
  initials,
  badgeCount,
  className,
}: SidebarRailItemProps) {
  const [isHovered, setIsHovered] = useState(false);

  return (
    <div
      className="group relative flex w-full justify-center py-1"
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <NavLink
        to={to}
        aria-label={label}
        className={({ isActive }) =>
          cn(
            'relative flex size-12 items-center justify-center font-semibold text-sm select-none transition-all duration-200 ease-out focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring',
            isActive
              ? 'rounded-[16px] bg-primary text-primary-foreground shadow-sm'
              : 'rounded-[24px] bg-secondary/80 text-secondary-foreground hover:rounded-[16px] hover:bg-primary hover:text-primary-foreground',
            className
          )
        }
      >
        {({ isActive }) => (
          <>
            {/* Discord Active / Hover Indicator Pill */}
            <span
              data-slot="rail-pill"
              aria-hidden="true"
              className={cn(
                'absolute left-0 w-1 bg-foreground rounded-r-full transition-all duration-200 ease-out -translate-x-[12px]',
                isActive
                  ? 'h-10 opacity-100 scale-y-100'
                  : isHovered
                    ? 'h-5 opacity-100 scale-y-100'
                    : 'h-0 opacity-0 scale-y-0'
              )}
            />

            {/* Icon or Initials */}
            {icon ? (
              <span className="flex size-6 items-center justify-center [&>svg]:size-6">{icon}</span>
            ) : (
              <span>{initials ?? label.slice(0, 2).toUpperCase()}</span>
            )}

            {/* Optional Unread / Notification Badge */}
            {badgeCount !== undefined && badgeCount > 0 && (
              <span className="absolute -bottom-1 -right-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-destructive px-1 text-xs font-bold leading-none text-destructive-foreground ring-2 ring-sidebar">
                {badgeCount > 99 ? '99+' : badgeCount}
              </span>
            )}
          </>
        )}
      </NavLink>

      {/* Floating Tooltip Label */}
      <div
        role="tooltip"
        className={cn(
          'pointer-events-none absolute left-[78px] top-1/2 z-50 -translate-y-1/2 rounded-md border border-border bg-popover px-2.5 py-1 text-xs font-semibold text-popover-foreground shadow-md whitespace-nowrap transition-all duration-150 ease-out',
          isHovered
            ? 'visible translate-x-0 opacity-100'
            : 'invisible -translate-x-1 opacity-0'
        )}
      >
        {label}
        {/* Subtle arrow pointer */}
        <span className="absolute -left-1 top-1/2 size-2 -translate-y-1/2 rotate-45 border-b border-l border-border bg-popover" />
      </div>
    </div>
  );
}
