import * as React from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import {
  useGetFollowersQuery,
  useGetFollowingQuery,
  useGetFriendsQuery,
  useLazyGetFollowersQuery,
  useLazyGetFollowingQuery,
  useLazyGetFriendsQuery,
  type FollowListUser,
} from '@/features/social/socialApi';

export type FollowListTab = 'followers' | 'following' | 'friends';

interface FollowListModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  handle: string;
  initialTab: FollowListTab;
}

const PAGE_SIZE = 30;

const TAB_LABELS: Record<FollowListTab, string> = {
  followers: 'Followers',
  following: 'Following',
  friends: 'Friends',
};

function useFirstPage(tab: FollowListTab, handle: string) {
  const args = { handle, limit: PAGE_SIZE };
  const followers = useGetFollowersQuery(args, { skip: tab !== 'followers' });
  const following = useGetFollowingQuery(args, { skip: tab !== 'following' });
  const friends = useGetFriendsQuery(args, { skip: tab !== 'friends' });
  return tab === 'followers' ? followers : tab === 'following' ? following : friends;
}

function useLazyPage(tab: FollowListTab) {
  const [lazyFollowers] = useLazyGetFollowersQuery();
  const [lazyFollowing] = useLazyGetFollowingQuery();
  const [lazyFriends] = useLazyGetFriendsQuery();
  if (tab === 'followers') return lazyFollowers;
  if (tab === 'following') return lazyFollowing;
  return lazyFriends;
}

function FollowListPane({ handle, tab }: { handle: string; tab: FollowListTab }) {
  const firstPage = useFirstPage(tab, handle);
  const lazyPage = useLazyPage(tab);
  // Pages after the first accumulate here, appended in the Load-more handler
  // (event, not effect) so no setState-in-effect is needed.
  const [extra, setExtra] = React.useState<FollowListUser[]>([]);
  const [nextCursor, setNextCursor] = React.useState<string | undefined>(undefined);
  const [hasMoreExtra, setHasMoreExtra] = React.useState(false);
  const [loadingMore, setLoadingMore] = React.useState(false);
  const [loadError, setLoadError] = React.useState<string | null>(null);

  const firstUsers = firstPage.data?.users ?? [];
  const users = [...firstUsers, ...extra];
  const hasMore = extra.length > 0 ? hasMoreExtra : (firstPage.data?.has_more ?? false);
  const cursor = extra.length > 0 ? nextCursor : firstPage.data?.next_cursor;

  async function handleLoadMore() {
    if (!cursor || loadingMore) return;
    setLoadingMore(true);
    setLoadError(null);
    try {
      const res = await lazyPage({ handle, limit: PAGE_SIZE, cursor }).unwrap();
      setExtra((prev) => {
        const seen = new Set([...firstUsers, ...prev].map((u) => u.id));
        return [...prev, ...res.users.filter((u) => !seen.has(u.id))];
      });
      setNextCursor(res.next_cursor);
      setHasMoreExtra(res.has_more);
    } catch {
      setLoadError('Couldn\u2019t load more. Please try again.');
    } finally {
      setLoadingMore(false);
    }
  }

  if (firstPage.isLoading) {
    return <p className="py-8 text-center text-sm text-muted-foreground">Loading…</p>;
  }
  if (firstPage.isError) {
    return (
      <p className="py-8 text-center text-sm text-destructive">
        Couldn&apos;t load this list. Please try again.
      </p>
    );
  }
  if (users.length === 0) {
    return (
      <p className="py-8 text-center text-sm text-muted-foreground">
        Nothing here yet.
      </p>
    );
  }
  return (
    <div>
      <ul className="divide-y divide-border">
        {users.map((u) => (
          <li key={u.id} className="flex items-center gap-2.5 py-2.5">
            <Avatar className="size-9">
              <AvatarFallback>
                {u.name ? u.name.slice(0, 2).toUpperCase() : 'TH'}
              </AvatarFallback>
            </Avatar>
            <div className="min-w-0">
              <p className="truncate text-sm font-semibold text-foreground">{u.name}</p>
              <p className="truncate text-xs text-muted-foreground">@{u.username || 'member'}</p>
            </div>
          </li>
        ))}
      </ul>
      {loadError && (
        <p role="alert" className="pt-2 text-center text-xs text-destructive">
          {loadError}
        </p>
      )}
      {hasMore && (
        <div className="flex justify-center pt-2">
          <Button
            variant="outline"
            size="sm"
            disabled={loadingMore || firstPage.isFetching}
            onClick={handleLoadMore}
          >
            {loadingMore ? 'Loading…' : 'Load more'}
          </Button>
        </div>
      )}
    </div>
  );
}

export function FollowListModal({ open, onOpenChange, handle, initialTab }: FollowListModalProps) {
  // Parent remounts per opened tab (key), so the initializer is enough.
  const [tab, setTab] = React.useState<FollowListTab>(initialTab);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{TAB_LABELS[tab]}</DialogTitle>
        </DialogHeader>
        <div className="flex gap-1 border-b border-border">
          {(Object.keys(TAB_LABELS) as FollowListTab[]).map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => setTab(t)}
              className={`px-3 py-2 text-sm transition-colors cursor-pointer ${
                tab === t
                  ? 'font-bold text-foreground border-b-2 border-primary -mb-px'
                  : 'font-medium text-muted-foreground hover:text-foreground'
              }`}
            >
              {TAB_LABELS[t]}
            </button>
          ))}
        </div>
        {open && <FollowListPane key={`${handle}-${tab}`} handle={handle} tab={tab} />}
      </DialogContent>
    </Dialog>
  );
}
