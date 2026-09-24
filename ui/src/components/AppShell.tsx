import { NavLink, Outlet, useMatches } from 'react-router';
import { TopAppBar } from '@/components/TopAppBar';
import { PrimarySidebar } from '@/components/PrimarySidebar';
import { House, ChatCircleDots, Compass, User } from '@phosphor-icons/react';
import { useGetMeQuery } from '@/features/auth/authApi';

interface RouteHandle {
  subtitle?: string;
}

export function AppShell() {
  const { data: user } = useGetMeQuery();
  const matches = useMatches();
  const profileHandle = user?.username || 'me';

  // Deepest route with a subtitle handle wins (e.g. Settings).
  const subtitle = [...matches]
    .reverse()
    .find((m) => (m.handle as RouteHandle | undefined)?.subtitle)?.handle as
    | RouteHandle
    | undefined;

  const mobileNavItems = [
    { to: '/', label: 'Home', icon: <House className="size-5" /> },
    { to: '/dms', label: 'DMs', icon: <ChatCircleDots className="size-5" /> },
    { to: '/spaces', label: 'Spaces', icon: <Compass className="size-5" /> },
    { to: `/u/${profileHandle}`, label: 'Profile', icon: <User className="size-5" /> },
  ];

  return (
    <div className="flex h-screen overflow-hidden bg-background text-foreground">
      {/* Desktop Discord-style Primary Sidebar */}
      <PrimarySidebar />

      {/* Main Content Area */}
      <div className="flex min-w-0 flex-1 flex-col">
        <TopAppBar subtitle={subtitle?.subtitle} />
        <div className="flex-1 overflow-y-auto p-4">
          <Outlet />
        </div>

        {/* Mobile Bottom Navigation Bar */}
        <nav
          aria-label="Mobile primary"
          className="hidden border-t border-border bg-card p-1.5 max-sm:flex items-center justify-around"
        >
          {mobileNavItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              aria-label={item.label}
              className={({ isActive }) =>
                `flex flex-col items-center gap-0.5 rounded-lg px-3 py-1 text-xs font-medium transition-colors ${
                  isActive
                    ? 'text-primary font-semibold'
                    : 'text-muted-foreground hover:text-foreground'
                }`
              }
            >
              {item.icon}
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>
      </div>
    </div>
  );
}
