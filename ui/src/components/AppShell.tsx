import { NavLink, Outlet } from 'react-router';
import { TopAppBar } from '@/components/TopAppBar';
import { ThemeToggle } from '@/components/ThemeToggle';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';

const railItems = [
  { to: '/spaces', label: 'Spaces', initials: 'Sp' },
  { to: '/feed', label: 'Feed', initials: 'Fe' },
  { to: '/u/joe', label: 'Profile', initials: 'Jo' },
];

export function AppShell() {
  return (
    <div className="flex min-h-screen bg-background text-foreground">
      <nav
        aria-label="Spaces"
        className="flex w-16 flex-col items-center gap-2 bg-muted py-3 max-sm:hidden"
      >
        {railItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            aria-label={item.label}
            className={({ isActive }) =>
              `rounded-md p-1 ${isActive ? 'bg-accent' : ''}`
            }
          >
            <Avatar>
              <AvatarFallback>{item.initials}</AvatarFallback>
            </Avatar>
          </NavLink>
        ))}
      </nav>
      <div className="flex min-w-0 flex-1 flex-col">
        <TopAppBar title="TownHall" actions={<ThemeToggle />} />
        <div className="flex-1 p-4">
          <Outlet />
        </div>
        <nav
          aria-label="Primary"
          className="hidden gap-1 bg-muted p-2 max-sm:flex"
        >
          {railItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `flex-1 rounded-md px-3 py-2 text-center text-sm ${
                  isActive
                    ? 'bg-accent text-accent-foreground'
                    : 'text-muted-foreground'
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </div>
    </div>
  );
}
