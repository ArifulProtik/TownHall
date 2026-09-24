import { Link } from 'react-router';
import { resolveMediaUrl } from '@/lib/media';
import { useGetFriendsQuery } from '@/features/social/socialApi';
import type { ProfileResponse } from '@/features/profile/profileApi';

interface ProfileIntroCardProps {
  profile: ProfileResponse;
  onNavigateTab: (tab: string) => void;
}

export function ProfileIntroCard({
  profile,
  onNavigateTab,
}: ProfileIntroCardProps) {
  const handle = profile.username || profile.id;
  const { data } = useGetFriendsQuery({ handle, limit: 4 });
  const friends = data?.users ?? [];
  const moreCount = data?.has_more ? '+' : '';

  const photoPreviews = [
    'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&w=300&q=80',
    'https://images.unsplash.com/photo-1579783900882-c0d3dad7b119?auto=format&fit=crop&w=300&q=80',
    'https://images.unsplash.com/photo-1607604276583-eef5d076aa5f?auto=format&fit=crop&w=300&q=80',
    'https://images.unsplash.com/photo-1550745165-9bc0b252726f?auto=format&fit=crop&w=300&q=80',
    'https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?auto=format&fit=crop&w=300&q=80',
    'https://images.unsplash.com/photo-1504384308090-c894fdcc538d?auto=format&fit=crop&w=300&q=80',
  ];

  return (
    <div className="space-y-3">
      {/* Photos — plain bordered shell, same system as header + posts */}
      <section className="rounded-2xl border border-border bg-card p-4">
        <div className="flex items-baseline justify-between mb-2">
          <h3 className="text-xl font-extrabold text-foreground">Photos</h3>
          <button
            type="button"
            onClick={() => onNavigateTab('photos')}
            className="text-sm text-primary hover:underline font-medium cursor-pointer"
          >
            See all photos
          </button>
        </div>
        <div className="grid grid-cols-3 gap-1 rounded-xl overflow-hidden">
          {photoPreviews.map((src, i) => (
            <div
              key={i}
              onClick={() => onNavigateTab('photos')}
              className="group aspect-square overflow-hidden bg-muted cursor-pointer"
            >
              <img
                src={src}
                alt=""
                loading="lazy"
                className="size-full object-cover transition-transform duration-200 group-hover:scale-105"
              />
            </div>
          ))}
        </div>
      </section>

      {/* Friends — X-style rows so names never truncate */}
      <section className="rounded-2xl border border-border bg-card p-4">
        <div className="flex items-baseline justify-between mb-2">
          <div className="flex items-baseline gap-1.5">
            <h3 className="text-xl font-extrabold text-foreground">Friends</h3>
            <p className="text-xs text-muted-foreground">{friends.length}{moreCount}</p>
          </div>
          <button
            type="button"
            onClick={() => onNavigateTab('community')}
            className="text-sm text-primary hover:underline font-medium cursor-pointer"
          >
            See all
          </button>
        </div>
        <div className="space-y-0.5">
          {friends.map((f) => (
            <Link
              key={f.id}
              to={`/u/${encodeURIComponent(f.username || f.id)}`}
              className="group flex cursor-pointer items-center gap-2.5 rounded-lg px-1 py-1.5 transition-colors hover:bg-muted/60"
            >
              <div className="size-10 shrink-0 overflow-hidden rounded-full bg-muted">
                {f.avatar_url ? (
                  <img
                    src={resolveMediaUrl(f.avatar_url)}
                    alt={f.name}
                    loading="lazy"
                    className="size-full object-cover"
                  />
                ) : (
                  <span className="flex size-full items-center justify-center text-xs font-bold text-muted-foreground">
                    {f.name ? f.name.slice(0, 2).toUpperCase() : 'TH'}
                  </span>
                )}
              </div>
              <div className="min-w-0">
                <p className="truncate text-base font-bold leading-tight text-foreground group-hover:underline">{f.name}</p>
                <p className="text-xs leading-tight text-muted-foreground">@{f.username || 'member'}</p>
              </div>
            </Link>
          ))}
          {friends.length === 0 && (
            <p className="py-2 text-sm text-muted-foreground">No friends yet.</p>
          )}
        </div>
      </section>
    </div>
  );
}
