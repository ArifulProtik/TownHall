import * as React from 'react';
import { Link } from 'react-router';
import { PaperPlaneTilt, UserCheck } from '@phosphor-icons/react';
import { Button } from '@/components/ui/button';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { resolveMediaUrl } from '@/lib/media';
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
  /** `rows` for modal lists, `cards` for the profile Friends grid. */
  variant?: 'rows' | 'cards';
}

function UserAvatar({ user, className }: { user: FollowListUser; className?: string }) {
  return (
    <Avatar className={className ?? 'size-9'}>
      {user.avatar_url ? (
        <AvatarImage
          src={resolveMediaUrl(user.avatar_url)}
          alt={user.name}
          className="object-cover"
        />
      ) : null}
      <AvatarFallback>
        {user.name ? user.name.slice(0, 2).toUpperCase() : 'TH'}
      </AvatarFallback>
    </Avatar>
  );
}

export function FollowList({ handle, tab, filter, onNavigate, variant = 'rows' }: FollowListProps) {
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
  if (variant === 'cards') {
    return (
      <div>
        <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
          {users.map((u) => (
            <div
              key={u.id}
              className="hover:bg-muted/40 flex items-center gap-2.5 rounded-lg p-2 transition-colors"
            >
              <UserAvatar user={u} className="size-12 shrink-0" />
              <div className="min-w-0 flex-1">
                <Link to={profilePath(u)} onClick={onNavigate}>
                  <h4 className="text-foreground truncate text-sm font-semibold hover:underline">
                    {u.name}
                  </h4>
                </Link>
                <p className="text-muted-foreground text-xs">@{u.username || 'member'}</p>
              </div>
              <div className="flex shrink-0 gap-1">
                <Link to="/dms">
                  <Button variant="outline" size="icon-xs" aria-label="Message">
                    <PaperPlaneTilt className="size-3" />
                  </Button>
                </Link>
                <Button variant="secondary" size="xs" className="gap-0.5">
                  <UserCheck className="size-3" />
                  Friends
                </Button>
              </div>
            </div>
          ))}
        </div>
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
              <UserAvatar user={u} />
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
