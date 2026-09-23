import { useParams } from 'react-router';
import { Card, CardContent } from '@/components/ui/card';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { PostCard } from '@/features/feed/PostCard';
import { mockProfiles, mockTimeline } from '@/features/profile/mocks';

export default function ProfilePage() {
  const { handle = '' } = useParams();
  const profile = mockProfiles.find((p) => p.handle === handle) ?? mockProfiles[0]!;
  const timeline = mockTimeline.filter((p) => p.handle === profile.handle);
  return (
    <div className="mx-auto w-full max-w-xl space-y-4">
      <Card>
        <CardContent className="flex items-center gap-4">
          <Avatar size="lg">
            <AvatarFallback>{profile.name.slice(0, 2)}</AvatarFallback>
          </Avatar>
          <div>
            <h2 className="text-lg font-medium text-foreground">{profile.name}</h2>
            <p className="text-sm text-muted-foreground">@{profile.handle}</p>
            <p className="mt-1 text-sm text-foreground">{profile.bio}</p>
            <p className="mt-1 text-xs text-muted-foreground">
              {profile.followers} followers · {profile.following} following
            </p>
          </div>
        </CardContent>
      </Card>
      <h3 className="text-base font-medium text-foreground">Timeline</h3>
      {timeline.length > 0 ? (
        timeline.map((post) => <PostCard key={post.id} post={post} />)
      ) : (
        <p className="text-sm text-muted-foreground">No posts yet.</p>
      )}
    </div>
  );
}
