import { House, ChatCircleDots, Plus } from '@phosphor-icons/react';
import { SidebarRailItem } from '@/components/SidebarRailItem';
import { UserMenuPopover } from '@/components/UserMenuPopover';
import { mockSpaces } from '@/features/spaces/mocks';

export function PrimarySidebar() {
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

      {/* Bottom Pinned: User Avatar & Popover */}
      <div className="mt-auto flex w-full flex-col items-center pt-2">
        <UserMenuPopover />
      </div>
    </nav>
  );
}
