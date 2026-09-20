import { NavLink, Outlet } from 'react-router';
import { TopAppBar } from '@/components/TopAppBar';
import { ThemeToggle } from '@/components/ThemeToggle';
import { Avatar } from '@/components/Avatar';

const railItems = [
  { to: '/spaces', label: 'Spaces', initials: 'Sp' },
  { to: '/feed', label: 'Feed', initials: 'Fe' },
  { to: '/u/joe', label: 'Profile', initials: 'Jo' },
];

export function AppShell() {
  return (
    <div className="flex min-h-screen bg-surface text-on-surface">
      <nav aria-label="Spaces" className="flex w-16 flex-col items-center gap-2 bg-surface-container py-3 max-sm:hidden">
        {railItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            aria-label={item.label}
            className={({ isActive }) =>
              `rounded-m3-md p-1 ${isActive ? 'bg-primary-container' : ''}`
            }
          >
            <Avatar initials={item.initials} />
          </NavLink>
        ))}
      </nav>
      <div className="flex min-w-0 flex-1 flex-col">
        <TopAppBar title="TownHall" actions={<ThemeToggle />} />
        <div className="flex-1 p-4">
          <Outlet />
        </div>
        <nav aria-label="Primary" className="hidden gap-1 bg-surface-container p-2 max-sm:flex">
          {railItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                `flex-1 rounded-m3-md px-3 py-2 text-center text-sm ${
                  isActive ? 'bg-primary-container text-on-primary-container' : 'text-on-surface-variant'
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
