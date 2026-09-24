import { useState } from 'react';
import { useNavigate } from 'react-router';
import { Popover } from '@base-ui/react/popover';
import { Checks, X } from '@phosphor-icons/react';
import { NotificationBell } from '@/components/NotificationBell';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { resolveMediaUrl } from '@/lib/media';
import {
  useGetNotificationsQuery,
  useMarkAllNotificationsReadMutation,
  useMarkNotificationReadMutation,
  type NotificationItem,
} from '@/features/notifications/notificationsApi';
import { cn } from 'cn';

function timeAgo(iso: string): string {
  const ms = Date.now() - new Date(iso).getTime();
  const mins = Math.max(1, Math.floor(ms / 60000));
  if (mins < 60) return `${mins}m`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h`;
  return `${Math.floor(hrs / 24)}d`;
}

function actionFor(item: NotificationItem): string {
  switch (item.type) {
    case 'follow':
      return 'started following you';
    case 'like':
      return 'liked your post';
    case 'comment':
      return 'commented on your post';
    case 'friend':
      return 'followed you back — you are now friends';
    default:
      return 'sent you a notification';
  }
}

interface Props {
  count: number;
}

// Stock Base UI popover (same pattern as the avatar menu): the trigger
// toggles, outside press / Escape / focus-out dismisses — no custom
// outside-click code. The popup is a full-height panel flush against the
// rail. Rows keep new/old styling while viewed; everything seen is marked
// read on close so the badge clears after viewing.
export function NotificationsPopover({ count }: Props) {
  const [open, setOpen] = useState(false);
  const navigate = useNavigate();
  const { data, isLoading } = useGetNotificationsQuery(
    { limit: 30 },
    { skip: !open },
  );
  const [markRead] = useMarkNotificationReadMutation();
  const [markAll] = useMarkAllNotificationsReadMutation();

  const close = () => {
    setOpen(false);
    if (count > 0) {
      void markAll();
    }
  };

  const handleOpenChange = (v: boolean) => {
    setOpen(v);
    if (!v && count > 0) {
      void markAll();
    }
  };

  const openActor = (item: NotificationItem) => {
    void markRead(item.id);
    close();
    navigate(`/u/${item.actor.username || item.actor.id}`);
  };

  const items = data?.notifications ?? [];

  return (
    <Popover.Root open={open} onOpenChange={handleOpenChange}>
      <Popover.Trigger
        render={
          <NotificationBell hasUnread={count > 0} unreadCount={count} />
        }
      />
      <Popover.Portal>
        <Popover.Positioner side="right" align="end" sideOffset={0} className="z-50">
          <Popover.Popup
            aria-label="Notifications"
            className="flex h-dvh w-[380px] max-w-[90vw] flex-col border-l border-border bg-popover text-popover-foreground shadow-xl outline-none focus:outline-none"
          >
            <div className="flex items-center justify-between border-b border-border px-4 py-3">
              <h2 className="text-sm font-semibold">Notifications</h2>
              <div className="flex items-center gap-1">
                <button
                  type="button"
                  onClick={() => void markAll()}
                  className="flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-xs text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                >
                  <Checks className="size-4" /> Mark all read
                </button>
                <button
                  type="button"
                  aria-label="Close notifications"
                  onClick={close}
                  className="cursor-pointer rounded-md p-1.5 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                >
                  <X className="size-4" />
                </button>
              </div>
            </div>
            <div className="flex-1 overflow-y-auto py-1">
              {isLoading ? (
                <p className="px-4 py-6 text-sm text-muted-foreground">Loading…</p>
              ) : items.length === 0 ? (
                <p className="px-4 py-6 text-sm text-muted-foreground">
                  No notifications yet. When someone follows you, it will show up here.
                </p>
              ) : (
                <ul>
                  {items.map((item) => (
                    <li key={item.id}>
                      <button
                        type="button"
                        onClick={() => openActor(item)}
                        className={cn(
                          'flex w-full cursor-pointer items-start gap-3 px-4 py-3 text-left transition-colors hover:bg-accent',
                          !item.read_at && 'bg-accent/40',
                        )}
                      >
                        <Avatar className="size-10 shrink-0">
                          {item.actor.avatar_url && (
                            <AvatarImage
                              src={resolveMediaUrl(item.actor.avatar_url)}
                              alt={item.actor.name}
                            />
                          )}
                          <AvatarFallback className="text-xs font-semibold">
                            {(item.actor.name || '?').slice(0, 1).toUpperCase()}
                          </AvatarFallback>
                        </Avatar>
                        <span className="min-w-0 flex-1">
                          <span className="block text-sm leading-snug">
                            <span className="font-semibold">{item.actor.name}</span>{' '}
                            <span className="font-normal">{actionFor(item)}</span>
                          </span>
                          <span className="mt-0.5 block text-xs text-muted-foreground">
                            {timeAgo(item.created_at)}
                            {!item.read_at && ' · new'}
                          </span>
                        </span>
                        {!item.read_at && (
                          <span
                            aria-hidden="true"
                            className="mt-1.5 size-2 shrink-0 rounded-full bg-primary"
                          />
                        )}
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </Popover.Popup>
        </Popover.Positioner>
      </Popover.Portal>
    </Popover.Root>
  );
}
