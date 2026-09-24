import { House, ChatCircleDots, Plus } from '@phosphor-icons/react';
import { SidebarRailItem } from '@/components/SidebarRailItem';
import { UserMenuPopover } from '@/components/UserMenuPopover';
import { mockSpaces } from '@/features/spaces/mocks';
import { useGetUnreadCountQuery } from '@/features/notifications/notificationsApi';
import { useNotificationStream } from '@/features/notifications/useNotificationStream';
import { NotificationsPopover } from '@/features/notifications/components/NotificationsPopover';
import { useSelector } from 'react-redux';
import { selectIsAuthenticated } from '@/features/auth/authSlice';

export function PrimarySidebar() {
  const isAuthed = useSelector(selectIsAuthenticated);
  useNotificationStream();
  const { data } = useGetUnreadCountQuery(undefined, { skip: !isAuthed });
  const count = data?.count ?? 0;
  return (
    <nav
      aria-label="Primary navigation rail"
      className="flex h-screen w-[72px] shrink-0 flex-col items-center border-r border-sidebar-border bg-sidebar py-3 text-sidebar-foreground select-none max-sm:hidden"
    >
      {/* Top Pinned: Home & DMs */}
      <div className="flex w-full flex-col items-center gap-1">
        <SidebarRailItem
          to="/"
          label="Home"
          icon={<House weight="fill" />}
        />
        <SidebarRailItem
          to="/dms"
          label="Direct Messages"
          icon={<ChatCircleDots weight="fill" />}
        />
      </div>

      {/* Pill Divider */}
      <div
        aria-hidden="true"
        className="my-2 h-[2px] w-8 shrink-0 rounded-full bg-sidebar-border"
      />

      {/* Middle Scrollable: Spaces & Explore (+) */}
      <div className="flex flex-1 w-full flex-col items-center gap-1 overflow-y-auto overflow-x-hidden py-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
        {mockSpaces.map((space) => (
          <SidebarRailItem
            key={space.id}
            to="/spaces"
            label={space.name}
            initials={space.name.slice(0, 2).toUpperCase()}
          />
        ))}

        <SidebarRailItem
          to="/spaces"
          label="Explore Spaces"
          icon={<Plus weight="bold" />}
          className="border border-dashed border-sidebar-border bg-transparent text-muted-foreground hover:border-primary hover:bg-primary hover:text-primary-foreground"
        />
      </div>

      {/* Bottom Pinned: Notifications & User Avatar */}
      <div className="mt-auto flex w-full flex-col items-center gap-1 pt-2">
        <NotificationsPopover count={count} />
        <UserMenuPopover />
      </div>
    </nav>
  );
}
