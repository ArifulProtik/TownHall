import { useState } from 'react';
import { Link, useNavigate } from 'react-router';
import { Popover } from '@base-ui/react/popover';
import { User, SignOut } from '@phosphor-icons/react';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Separator } from '@/components/ui/separator';
import { ThemeToggle } from '@/components/ThemeToggle';
import { authApi, useLogoutMutation } from '@/features/auth/authApi';
import { clearCredentials } from '@/features/auth/authSlice';
import { useAppDispatch } from '@/app/hooks';

export function UserMenuPopover() {
  const [open, setOpen] = useState(false);
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const [logout] = useLogoutMutation();

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

  return (
    <Popover.Root open={open} onOpenChange={setOpen}>
      <Popover.Trigger
        type="button"
        aria-label="User settings"
        className="group relative flex size-12 cursor-pointer items-center justify-center rounded-full transition-transform focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring hover:scale-105"
      >
        <Avatar className="size-10 border border-sidebar-border bg-sidebar-accent">
          <AvatarFallback className="text-sm font-semibold text-foreground">
            TH
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
                <AvatarFallback className="text-xs font-semibold text-primary">
                  TH
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-foreground">TownHall User</p>
                <p className="truncate text-xs text-muted-foreground">@townhall_member</p>
              </div>
            </div>

            <Separator className="my-1.5" />

            {/* Profile Navigation */}
            <Link
              to="/u/joe"
              onClick={() => setOpen(false)}
              className="flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm text-foreground hover:bg-accent hover:text-accent-foreground"
            >
              <User className="size-4" />
              <span>My Profile</span>
            </Link>

            {/* Theme Toggle Item */}
            <div className="flex items-center justify-between rounded-lg px-2.5 py-1.5 text-sm text-foreground hover:bg-accent hover:text-accent-foreground">
              <span>Theme</span>
              <ThemeToggle />
            </div>

            <Separator className="my-1.5" />

            {/* Log out */}
            <button
              type="button"
              onClick={onLogout}
              className="flex w-full cursor-pointer items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm text-destructive hover:bg-destructive/10"
            >
              <SignOut className="size-4" />
              <span>Log out</span>
            </button>
          </Popover.Popup>
        </Popover.Positioner>
      </Popover.Portal>
    </Popover.Root>
  );
}
