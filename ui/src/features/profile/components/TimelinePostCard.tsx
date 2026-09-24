import { useState } from 'react';
import {
  Heart,
  ChatCircle,
  ShareFat,
  DotsThree,
  PaperPlaneRight,
} from '@phosphor-icons/react';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { resolveMediaUrl } from '@/lib/media';
import type { ProfileResponse } from '@/features/profile/profileApi';

export interface TimelinePost {
  id: string;
  authorName: string;
  authorHandle: string;
  authorAvatar?: string;
  time: string;
  content: string;
  image?: string;
  likesCount: number;
  commentsCount: number;
  sharesCount: number;
  comments?: Array<{
    id: string;
    authorName: string;
    authorAvatar?: string;
    text: string;
    time: string;
    likes: number;
  }>;
}

interface TimelinePostCardProps {
  post: TimelinePost;
  currentProfile: ProfileResponse;
}

export function TimelinePostCard({ post, currentProfile }: TimelinePostCardProps) {
  const [isLiked, setIsLiked] = useState(false);
  const [likesCount, setLikesCount] = useState(post.likesCount);
  const [showComments, setShowComments] = useState(false);
  const [comments, setComments] = useState(post.comments || []);
  const [newComment, setNewComment] = useState('');
  const [imageError, setImageError] = useState(false);

  const authorInitials = post.authorName.slice(0, 2).toUpperCase();
  const currentUserInitials = currentProfile.name
    ? currentProfile.name.slice(0, 2).toUpperCase()
    : 'ME';

  function handleToggleLike() {
    setIsLiked((prev) => !prev);
    setLikesCount((prev) => (isLiked ? Math.max(0, prev - 1) : prev + 1));
  }

  function handleAddComment(e: React.FormEvent) {
    e.preventDefault();
    if (!newComment.trim()) return;
    setComments((prev) => [
      ...prev,
      {
        id: `c_${Date.now()}`,
        authorName: currentProfile.name || 'You',
        authorAvatar: currentProfile.avatar_url,
        text: newComment.trim(),
        time: 'Just now',
        likes: 0,
      },
    ]);
    setNewComment('');
  }

  const totalComments = comments.length > 0 ? comments.length : post.commentsCount;

  return (
    <article className="rounded-xl border border-border bg-card overflow-hidden">
      {/* Header: X-style single baseline row */}
      <div className="flex items-center justify-between px-4 pt-3 pb-0">
        <div className="flex min-w-0 items-center gap-2">
          <Avatar className="size-10 shrink-0">
            {post.authorAvatar && (
              <AvatarImage src={resolveMediaUrl(post.authorAvatar)} alt={post.authorName} />
            )}
            <AvatarFallback className="font-semibold">{authorInitials}</AvatarFallback>
          </Avatar>
          <div className="flex min-w-0 items-baseline gap-1.5">
            <p className="truncate text-base font-bold leading-tight text-foreground">
              {post.authorName}
            </p>
            <span className="shrink-0 text-base leading-tight text-muted-foreground">
              · {post.time}
            </span>
          </div>
        </div>
        <Button type="button" variant="ghost" size="icon-xs" aria-label="More" className="text-muted-foreground -mr-1">
          <DotsThree className="size-4" weight="bold" />
        </Button>
      </div>

      {/* Content */}
      <p className="px-4 pt-1.5 pb-1.5 text-base/5 text-foreground">
        {post.content}
      </p>

      {/* Image */}
      {post.image && !imageError && (
        <div className="px-4 pb-1">
          <div className="overflow-hidden rounded-lg border border-border bg-muted">
            <img
              src={post.image}
              alt=""
              onError={() => setImageError(true)}
              className="w-full max-h-[420px] object-cover"
            />
          </div>
        </div>
      )}

      {/* Engagement summary */}
      {(likesCount > 0 || totalComments > 0 || post.sharesCount > 0) && (
        <div className="flex items-center justify-between px-4 py-1 text-xs text-muted-foreground">
          <div className="flex items-center gap-1">
            {likesCount > 0 && (
              <>
                <Heart className="size-3.5 text-rose-500" weight="fill" />
                <span>{likesCount}</span>
              </>
            )}
          </div>
          <div className="flex items-center gap-2.5">
            {totalComments > 0 && (
              <button
                type="button"
                onClick={() => setShowComments((p) => !p)}
                className="hover:underline cursor-pointer"
              >
                {totalComments} comment{totalComments !== 1 && 's'}
              </button>
            )}
            {post.sharesCount > 0 && <span>{post.sharesCount} shares</span>}
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="flex items-stretch border-t border-border mx-1">
        <button
          type="button"
          onClick={handleToggleLike}
          className={`flex flex-1 items-center justify-center gap-1 py-1 text-sm font-medium transition-colors cursor-pointer rounded-md m-0.5 ${
            isLiked
              ? 'text-rose-500'
              : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'
          }`}
        >
          <Heart className="size-3.5" weight={isLiked ? 'fill' : 'regular'} />
          Like
        </button>
        <button
          type="button"
          onClick={() => setShowComments((p) => !p)}
          className="flex flex-1 items-center justify-center gap-1 py-1 text-sm font-medium text-muted-foreground hover:bg-muted/60 hover:text-foreground transition-colors cursor-pointer rounded-md m-0.5"
        >
          <ChatCircle className="size-3.5" />
          Comment
        </button>
        <button
          type="button"
          className="flex flex-1 items-center justify-center gap-1 py-1 text-sm font-medium text-muted-foreground hover:bg-muted/60 hover:text-foreground transition-colors cursor-pointer rounded-md m-0.5"
        >
          <ShareFat className="size-3.5" />
          Share
        </button>
      </div>

      {/* Comments */}
      {showComments && (
        <div className="border-t border-border px-3.5 py-2 space-y-2">
          {comments.map((c) => (
            <div key={c.id} className="flex items-start gap-1.5">
              <Avatar className="size-6 shrink-0">
                {c.authorAvatar && <AvatarImage src={resolveMediaUrl(c.authorAvatar)} alt={c.authorName} />}
                <AvatarFallback className="font-semibold">
                  {c.authorName.slice(0, 2).toUpperCase()}
                </AvatarFallback>
              </Avatar>
              <div>
                <div className="inline-block rounded-xl bg-muted px-3 py-1.5">
                  <p className="text-base font-bold text-foreground leading-tight">{c.authorName}</p>
                  <p className="text-base/5 text-foreground/90">{c.text}</p>
                </div>
                <div className="flex gap-2.5 pl-2.5 mt-0.5 text-xs text-muted-foreground">
                  <button type="button" className="font-medium hover:underline cursor-pointer">Like</button>
                  <button type="button" className="font-medium hover:underline cursor-pointer">Reply</button>
                  <span>{c.time}</span>
                </div>
              </div>
            </div>
          ))}

          <form onSubmit={handleAddComment} className="flex items-center gap-1.5">
            <Avatar className="size-6 shrink-0">
              {currentProfile.avatar_url && (
                <AvatarImage src={resolveMediaUrl(currentProfile.avatar_url)} alt={currentProfile.name} />
              )}
              <AvatarFallback className="font-semibold">{currentUserInitials}</AvatarFallback>
            </Avatar>
            <div className="relative flex-1">
              <input
                type="text"
                value={newComment}
                onChange={(e) => setNewComment(e.target.value)}
                placeholder="Write a comment..."
                className="w-full rounded-full bg-muted px-3 py-1 pr-8 text-base md:text-sm text-foreground placeholder:text-muted-foreground focus:outline-hidden focus:ring-1 focus:ring-ring"
              />
              <button
                type="submit"
                disabled={!newComment.trim()}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-primary disabled:opacity-30 cursor-pointer"
              >
                <PaperPlaneRight className="size-3" weight="fill" />
              </button>
            </div>
          </form>
        </div>
      )}
    </article>
  );
}
