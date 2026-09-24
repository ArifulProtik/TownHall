import { useState } from 'react';
import {
  Image as ImageIcon,
  Smiley,
  GlobeHemisphereWest,
  CaretDown,
  X,
  Users,
  MapPin,
} from '@phosphor-icons/react';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import type { ProfileResponse } from '@/features/profile/profileApi';
import { resolveMediaUrl } from '@/lib/media';

interface CreatePostPromptProps {
  profile: ProfileResponse;
  onPostCreated?: (content: string, image?: string) => void;
}

export function CreatePostPrompt({ profile, onPostCreated }: CreatePostPromptProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [content, setContent] = useState('');
  const [selectedImage, setSelectedImage] = useState<string | null>(null);

  const firstName = profile.name ? profile.name.split(' ')[0] : 'Member';
  const initials = profile.name ? profile.name.slice(0, 2).toUpperCase() : 'TH';

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!content.trim() && !selectedImage) return;
    onPostCreated?.(content.trim(), selectedImage || undefined);
    setContent('');
    setSelectedImage(null);
    setIsOpen(false);
  }

  return (
    <>
      {/* Compact inline trigger */}
      <div className="flex items-center gap-2.5 rounded-xl border border-border bg-card px-3 py-2">
        <Avatar className="size-10 shrink-0">
          {profile.avatar_url && (
            <AvatarImage src={resolveMediaUrl(profile.avatar_url)} alt={profile.name} />
          )}
          <AvatarFallback className="font-semibold">{initials}</AvatarFallback>
        </Avatar>
        <button
          type="button"
          onClick={() => setIsOpen(true)}
          className="flex-1 rounded-full bg-muted hover:bg-muted/80 px-4 py-2 text-left text-base text-muted-foreground transition-colors cursor-pointer"
        >
          Share something, {firstName}...
        </button>
        <button
          type="button"
          onClick={() => setIsOpen(true)}
          className="text-muted-foreground hover:text-primary transition-colors cursor-pointer"
        >
          <ImageIcon className="size-4" weight="fill" />
        </button>
      </div>

      {/* Create post modal */}
      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="sm:max-w-lg p-0 overflow-hidden rounded-2xl">
          <DialogHeader className="px-4 pt-3 pb-2 border-b border-border">
            <DialogTitle className="text-xl font-extrabold text-center">
              Create post
            </DialogTitle>
          </DialogHeader>

          <form onSubmit={handleSubmit} className="p-4 space-y-3">
            <div className="flex items-center gap-2">
              <Avatar className="size-8">
                {profile.avatar_url && (
                  <AvatarImage src={resolveMediaUrl(profile.avatar_url)} alt={profile.name} />
                )}
                <AvatarFallback className="font-semibold">{initials}</AvatarFallback>
              </Avatar>
              <div>
                <p className="text-sm font-semibold text-foreground">{profile.name}</p>
                <div className="inline-flex items-center gap-0.5 rounded bg-muted px-1.5 py-0.5 text-xs font-medium text-muted-foreground">
                  <GlobeHemisphereWest className="size-2.5" />
                  Public
                  <CaretDown className="size-2" />
                </div>
              </div>
            </div>

            <textarea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder={`What's on your mind, ${firstName}?`}
              rows={3}
              autoFocus
              className="w-full resize-none border-none bg-transparent text-base md:text-sm text-foreground placeholder:text-muted-foreground/60 focus:outline-hidden leading-relaxed"
            />

            {selectedImage && (
              <div className="relative rounded-lg overflow-hidden border border-border bg-muted max-h-44">
                <img src={selectedImage} alt="" className="w-full max-h-44 object-cover" />
                <button
                  type="button"
                  onClick={() => setSelectedImage(null)}
                  className="absolute top-1.5 right-1.5 flex size-6 items-center justify-center rounded-full bg-black/50 text-white hover:bg-black/70 transition-colors cursor-pointer"
                >
                  <X className="size-3" />
                </button>
              </div>
            )}

            <div className="flex items-center justify-between rounded-lg border border-border px-3 py-2">
              <span className="text-sm font-medium text-foreground">Add to post</span>
              <div className="flex items-center gap-0.5">
                <Button type="button" variant="ghost" size="icon-xs" aria-label="Photo" className="text-emerald-500 rounded-full">
                  <ImageIcon className="size-4" weight="fill" />
                </Button>
                <Button type="button" variant="ghost" size="icon-xs" aria-label="Tag" className="text-blue-500 rounded-full">
                  <Users className="size-4" weight="fill" />
                </Button>
                <Button type="button" variant="ghost" size="icon-xs" aria-label="Feeling" className="text-amber-500 rounded-full">
                  <Smiley className="size-4" weight="fill" />
                </Button>
                <Button type="button" variant="ghost" size="icon-xs" aria-label="Check in" className="text-rose-500 rounded-full">
                  <MapPin className="size-4" weight="fill" />
                </Button>
              </div>
            </div>

            <Button
              type="submit"
              disabled={!content.trim() && !selectedImage}
              className="w-full font-semibold text-sm"
              size="sm"
            >
              Post
            </Button>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
}
