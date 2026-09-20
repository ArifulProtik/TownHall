import { useParams } from 'react-router';
import { M3Card } from '@/components/M3Card';
import { Avatar } from '@/components/Avatar';
import { PostCard } from '@/features/feed/PostCard';
import { mockProfiles, mockTimeline } from '@/features/profile/mocks';

export default function ProfilePage() {
  const { handle = '' } = useParams();
  const profile = mockProfiles.find((p) => p.handle === handle) ?? mockProfiles[0]!;
  const timeline = mockTimeline.filter((p) => p.handle === profile.handle);
  return (
    <div className="mx-auto w-full max-w-xl space-y-4">
      <M3Card className="flex items-center gap-4 p-4">
        <Avatar initials={profile.name.slice(0, 2)} size="lg" />
        <div>
          <h2 className="text-lg font-medium text-on-surface">{profile.name}</h2>
          <p className="text-sm text-on-surface-variant">@{profile.handle}</p>
          <p className="mt-1 text-sm text-on-surface">{profile.bio}</p>
          <p className="mt-1 text-xs text-on-surface-variant">
            {profile.followers} followers · {profile.following} following
          </p>
        </div>
      </M3Card>
      <h3 className="text-base font-medium text-on-surface">Timeline</h3>
      {timeline.length > 0 ? (
        timeline.map((post) => <PostCard key={post.id} post={post} />)
      ) : (
        <p className="text-sm text-on-surface-variant">No posts yet.</p>
      )}
    </div>
  );
}
