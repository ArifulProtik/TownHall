import { useState } from 'react';
import { Link, useNavigate } from 'react-router';
import { Popover } from '@base-ui/react/popover';
import { User, SignOut, Moon, Sun, Gear } from '@phosphor-icons/react';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { resolveMediaUrl } from '@/lib/media';
import { Separator } from '@/components/ui/separator';
import { useTheme } from '@/components/ThemeToggle';
import { authApi, useLogoutMutation, useGetMeQuery } from '@/features/auth/authApi';
import { clearCredentials } from '@/features/auth/authSlice';
import { useAppDispatch } from '@/app/hooks';

function getInitials(name?: string): string {
  if (!name || !name.trim()) return 'TH';
  const parts = name.trim().split(/\s+/);
  if (parts.length === 1) return parts[0]!.slice(0, 2).toUpperCase();
  return (parts[0]![0]! + parts[parts.length - 1]![0]!).toUpperCase();
}

export function UserMenuPopover() {
  const [open, setOpen] = useState(false);
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const [logout] = useLogoutMutation();
  const { data: user } = useGetMeQuery();
  const { dark, toggle: toggleTheme } = useTheme();

  async function onLogout() {
    try {
      await logout().unwrap();
    } catch {
      // ignore network errors on logout to allow clean client logout
    } finally {
      dispatch(clearCredentials());
      dispatch(authApi.util.resetApiState());
      setOpen(false);
      navigate('/login', { replace: true });
    }
  }

  const initials = getInitials(user?.name);
  const displayName = user?.name ?? 'TownHall User';
  const displayHandle = user?.username
    ? `@${user.username}`
    : user?.email
      ? `@${user.email.split('@')[0]}`
      : '@member';
  const profileHandle = user?.username || 'me';

  return (
    <Popover.Root open={open} onOpenChange={setOpen}>
      <Popover.Trigger
        type="button"
        aria-label="User settings"
        className="group relative flex size-12 cursor-pointer items-center justify-center rounded-full transition-transform focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring hover:scale-105"
      >
        <Avatar className="size-10 border border-sidebar-border bg-sidebar-accent">
          {user?.avatar_url && (
            <AvatarImage src={resolveMediaUrl(user.avatar_url)} alt={displayName} />
          )}
          <AvatarFallback className="text-sm font-semibold text-foreground">
            {initials}
          </AvatarFallback>
        </Avatar>
        {/* Discord-style Online status dot */}
        <span
          data-slot="online-dot"
          className="absolute bottom-1 right-1 size-3 rounded-full bg-emerald-500 ring-2 ring-sidebar"
        />
      </Popover.Trigger>

      <Popover.Portal>
        <Popover.Positioner side="right" align="end" sideOffset={14} className="z-50">
          <Popover.Popup className="w-60 rounded-xl border border-border bg-popover p-2 text-popover-foreground shadow-xl outline-none focus:outline-none">
            {/* User Details Header */}
            <div className="flex items-center gap-3 p-2">
              <Avatar className="size-9 bg-primary/10">
                {user?.avatar_url && (
                  <AvatarImage src={resolveMediaUrl(user.avatar_url)} alt={displayName} />
                )}
                <AvatarFallback className="text-xs font-semibold text-primary">
                  {initials}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-foreground">{displayName}</p>
                <p className="truncate text-xs text-muted-foreground">{displayHandle}</p>
              </div>
            </div>

            <Separator className="my-1.5" />

            {/* Profile Navigation */}
            <Link
              to={`/u/${profileHandle}`}
              onClick={() => setOpen(false)}
              className="flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm text-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
            >
              <User className="size-4 shrink-0 text-muted-foreground" />
              <span>My Profile</span>
            </Link>

            {/* Settings Navigation */}
            <Link
              to="/settings"
              onClick={() => setOpen(false)}
              className="flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm text-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
            >
              <Gear className="size-4 shrink-0 text-muted-foreground" />
              <span>Settings</span>
            </Link>

            {/* Theme Toggle Item - Consistent row layout and height */}
            <button
              type="button"
              onClick={toggleTheme}
              className="flex w-full cursor-pointer items-center justify-between rounded-lg px-2.5 py-2 text-sm text-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
            >
              <div className="flex items-center gap-2.5">
                {dark ? (
                  <Sun className="size-4 shrink-0 text-muted-foreground" />
                ) : (
                  <Moon className="size-4 shrink-0 text-muted-foreground" />
                )}
                <span>Theme</span>
              </div>
              <span className="text-xs text-muted-foreground capitalize">
                {dark ? 'Dark' : 'Light'}
              </span>
            </button>

            <Separator className="my-1.5" />

            {/* Log out */}
            <button
              type="button"
              onClick={onLogout}
              className="flex w-full cursor-pointer items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm text-destructive transition-colors hover:bg-destructive/10"
            >
              <SignOut className="size-4 shrink-0" />
              <span>Log out</span>
            </button>
          </Popover.Popup>
        </Popover.Positioner>
      </Popover.Portal>
    </Popover.Root>
  );
}
