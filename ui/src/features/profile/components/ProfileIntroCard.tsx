import type { ProfileResponse } from '@/features/profile/profileApi';

interface ProfileIntroCardProps {
  profile: ProfileResponse;
  onNavigateTab: (tab: string) => void;
}

export function ProfileIntroCard({
  profile,
  onNavigateTab,
}: ProfileIntroCardProps) {
  const friends = [
    { name: 'Sarah Chen', mutual: '18 mutual', avatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=200&q=80' },
    { name: 'Marcus V.', mutual: '12 mutual', avatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&w=200&q=80' },
    { name: 'Elena R.', mutual: '34 mutual', avatar: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=200&q=80' },
    { name: 'Alex Rivera', mutual: '9 mutual', avatar: 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&w=200&q=80' },
    { name: 'Liam D.', mutual: '22 mutual', avatar: 'https://images.unsplash.com/photo-1522075469751-3a6694fb2f61?auto=format&fit=crop&w=200&q=80' },
    { name: 'Maya Patel', mutual: '15 mutual', avatar: 'https://images.unsplash.com/photo-1517841905240-472988babdf9?auto=format&fit=crop&w=200&q=80' },
  ];

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
            <p className="text-xs text-muted-foreground">{profile.followers_count || 489}</p>
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
          {friends.slice(0, 4).map((f, i) => (
            <div
              key={i}
              onClick={() => onNavigateTab('community')}
              className="group flex cursor-pointer items-center gap-2.5 rounded-lg px-1 py-1.5 transition-colors hover:bg-muted/60"
            >
              <div className="size-10 shrink-0 overflow-hidden rounded-full bg-muted">
                <img
                  src={f.avatar}
                  alt={f.name}
                  loading="lazy"
                  className="size-full object-cover"
                />
              </div>
              <div className="min-w-0">
                <p className="truncate text-base font-bold leading-tight text-foreground group-hover:underline">{f.name}</p>
                <p className="text-xs leading-tight text-muted-foreground">{f.mutual}</p>
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
