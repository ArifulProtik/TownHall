import * as React from 'react';
import { UserPlus, UserCheck } from '@phosphor-icons/react';
import { Button } from '@/components/ui/button';
import { useAppSelector } from '@/app/hooks';
import { selectIsAuthenticated } from '@/features/auth/authSlice';
import {
  useGetFollowStatusQuery,
  useFollowUserMutation,
  useUnfollowUserMutation,
} from '@/features/social/socialApi';

interface FollowButtonProps {
  handle: string;
}

export function FollowButton({ handle }: FollowButtonProps) {
  const isAuthenticated = useAppSelector(selectIsAuthenticated);
  const { data: status, isLoading } = useGetFollowStatusQuery(handle, {
    skip: !isAuthenticated,
  });
  const [follow, { isLoading: isFollowing }] = useFollowUserMutation();
  const [unfollow, { isLoading: isUnfollowing }] = useUnfollowUserMutation();
  const [error, setError] = React.useState<string | null>(null);

  if (!isAuthenticated || isLoading || !status || status.is_self) {
    return null;
  }

  const busy = isFollowing || isUnfollowing;

  async function handleClick() {
    setError(null);
    try {
      if (status?.is_following) {
        await unfollow(handle).unwrap();
      } else {
        await follow(handle).unwrap();
      }
    } catch {
      setError('Something went wrong. Please try again.');
    }
  }

  const label = status.is_friend
    ? 'Friends'
    : status.is_following
      ? 'Following'
      : status.is_followed_by
        ? 'Follow Back'
        : 'Follow';

  return (
    <span className="inline-flex flex-col items-end gap-1">
      <Button
        size="sm"
        variant={status.is_following ? 'outline' : 'default'}
        onClick={handleClick}
        disabled={busy}
        className="gap-1.5 font-medium"
      >
        {status.is_following ? (
          <UserCheck className="size-3.5" />
        ) : (
          <UserPlus className="size-3.5" />
        )}
        <span>{busy ? 'Saving…' : label}</span>
      </Button>
      {error && (
        <span role="alert" className="text-xs text-destructive">
          {error}
        </span>
      )}
    </span>
  );
}
