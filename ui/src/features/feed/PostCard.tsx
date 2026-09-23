import { Card, CardContent } from '@/components/ui/card';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import type { MockPost } from '@/features/feed/mocks';

export function PostCard({ post }: { post: MockPost }) {
  return (
    <Card>
      <CardContent className="space-y-2">
        <div className="flex items-center gap-3">
          <Avatar size="sm">
            <AvatarFallback>{post.author.slice(0, 2)}</AvatarFallback>
          </Avatar>
          <div>
            <p className="text-sm font-medium text-foreground">{post.author}</p>
            <p className="text-xs text-muted-foreground">
              @{post.handle} · {post.time}
            </p>
          </div>
        </div>
        <p className="text-sm text-foreground">{post.body}</p>
        <p className="text-xs text-muted-foreground">
          {post.likes} likes · {post.replies} replies
        </p>
      </CardContent>
    </Card>
  );
}
