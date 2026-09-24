import * as React from 'react';
import { Link } from 'react-router';
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

const PAGE_SIZE = 30;

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

function profilePath(u: FollowListUser): string {
  return `/u/${encodeURIComponent(u.username || u.id)}`;
}

interface FollowListProps {
  handle: string;
  tab: FollowListTab;
  /** Optional client-side filter over loaded users (matches name/@username). */
  filter?: string;
  /** Called when a row is clicked (e.g. to close a wrapping modal). */
  onNavigate?: () => void;
}

export function FollowList({ handle, tab, filter, onNavigate }: FollowListProps) {
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
  const allUsers = [...firstUsers, ...extra];
  const q = (filter ?? '').trim().toLowerCase();
  const users = q
    ? allUsers.filter(
        (u) =>
          u.name.toLowerCase().includes(q) ||
          (u.username ?? '').toLowerCase().includes(q),
      )
    : allUsers;
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
        {q ? 'No matches.' : 'Nothing here yet.'}
      </p>
    );
  }
  return (
    <div>
      <ul className="divide-y divide-border">
        {users.map((u) => (
          <li key={u.id}>
            <Link
              to={profilePath(u)}
              onClick={onNavigate}
              className="flex items-center gap-2.5 rounded-lg py-2.5 transition-colors hover:bg-muted/40"
            >
              <Avatar className="size-9">
                <AvatarFallback>
                  {u.name ? u.name.slice(0, 2).toUpperCase() : 'TH'}
                </AvatarFallback>
              </Avatar>
              <div className="min-w-0">
                <p className="truncate text-sm font-semibold text-foreground">{u.name}</p>
                <p className="truncate text-xs text-muted-foreground">@{u.username || 'member'}</p>
              </div>
            </Link>
          </li>
        ))}
      </ul>
      {loadError && (
        <p role="alert" className="pt-2 text-center text-xs text-destructive">
          {loadError}
        </p>
      )}
      {hasMore && !q && (
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
