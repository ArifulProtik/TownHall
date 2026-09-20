import { M3Card } from '@/components/M3Card';
import { Avatar } from '@/components/Avatar';
import type { MockPost } from '@/features/feed/mocks';

export function PostCard({ post }: { post: MockPost }) {
  return (
    <M3Card className="space-y-2 p-4">
      <div className="flex items-center gap-3">
        <Avatar initials={post.author.slice(0, 2)} size="sm" />
        <div>
          <p className="text-sm font-medium text-on-surface">{post.author}</p>
          <p className="text-xs text-on-surface-variant">
            @{post.handle} · {post.time}
          </p>
        </div>
      </div>
      <p className="text-sm text-on-surface">{post.body}</p>
      <p className="text-xs text-on-surface-variant">
        {post.likes} likes · {post.replies} replies
      </p>
    </M3Card>
  );
}
