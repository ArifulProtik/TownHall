import { PostCard } from '@/features/feed/PostCard';
import { mockPosts } from '@/features/feed/mocks';

export default function FeedPage() {
  return (
    <div className="mx-auto w-full max-w-xl space-y-4">
      <h2 className="text-xl font-extrabold text-foreground">Feed</h2>
      {mockPosts.map((post) => (
        <PostCard key={post.id} post={post} />
      ))}
    </div>
  );
}
