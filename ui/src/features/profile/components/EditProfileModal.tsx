import * as React from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar';
import { resolveMediaUrl } from '@/lib/media';
import { ImageUploadButton } from '@/features/profile/components/ImageUploadButton';
import { useUpdateProfileMutation, type ProfileResponse } from '@/features/profile/profileApi';

interface EditProfileModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  profile: ProfileResponse;
}

function EditProfileForm({ profile, onClose }: { profile: ProfileResponse; onClose: () => void }) {
  const [updateProfile, { isLoading }] = useUpdateProfileMutation();

  const [name, setName] = React.useState(profile.name || '');
  const [bio, setBio] = React.useState(profile.bio || '');
  const [location, setLocation] = React.useState(profile.location || '');
  const [website, setWebsite] = React.useState(profile.website || '');
  const [avatarUrl, setAvatarUrl] = React.useState(profile.avatar_url || '');
  const [bannerUrl, setBannerUrl] = React.useState(profile.banner_url || '');
  const [bannerPreviewError, setBannerPreviewError] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);

  const initials = profile.name ? profile.name.slice(0, 2).toUpperCase() : 'TH';

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim() || name.length < 2) {
      setError('Name must be at least 2 characters');
      return;
    }
    if (bio.length > 280) {
      setError('Bio must not exceed 280 characters');
      return;
    }

    setError(null);
    try {
      await updateProfile({
        name: name.trim(),
        bio: bio.trim(),
        location: location.trim(),
        website: website.trim(),
        avatar_url: avatarUrl.trim(),
        banner_url: bannerUrl.trim(),
      }).unwrap();
      onClose();
    } catch (err: unknown) {
      const apiErr = err as { data?: { error?: string } };
      setError(apiErr?.data?.error || 'Failed to update profile. Please try again.');
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex min-h-0 flex-1 flex-col">
      {error && (
        <div className="px-5 pt-4 sm:px-6">
          <div className="bg-destructive/10 text-destructive rounded-lg p-2.5 text-xs font-medium">
            {error}
          </div>
        </div>
      )}

      <div className="min-h-0 flex-1 p-5 sm:p-6 md:columns-2 md:gap-6 md:[column-fill:auto]">
        {/* Identity block — media plus name stay together */}
        <div className="mb-5 break-inside-avoid space-y-4">
          <div>
            <div className="border-border relative h-28 w-full overflow-hidden rounded-xl border bg-gradient-to-br from-[#2f6bff] via-[#6d5cff] to-[#a855f7]">
              {bannerUrl && !bannerPreviewError ? (
                <img
                  src={resolveMediaUrl(bannerUrl)}
                  alt="Banner preview"
                  onError={() => setBannerPreviewError(true)}
                  className="size-full object-cover"
                />
              ) : (
                <>
                  <div className="absolute inset-0 bg-[radial-gradient(420px_circle_at_15%_0%,rgba(255,255,255,0.32),transparent_62%)]" />
                  <div className="absolute inset-0 bg-[radial-gradient(400px_circle_at_88%_108%,rgba(255,186,120,0.38),transparent_65%)]" />
                </>
              )}
              <div className="absolute inset-0 bg-gradient-to-t from-black/50 via-black/5 to-transparent" />
              <div className="absolute top-2.5 right-2.5">
                <ImageUploadButton
                  label="Cover"
                  onUploaded={(url) => setBannerUrl(url)}
                  className="px-2.5 py-1 text-xs"
                />
              </div>
            </div>

            <div className="relative -mt-10 ml-4 flex items-end gap-2.5">
              <div className="relative size-20 shrink-0">
                <div className="border-card bg-muted size-full overflow-hidden rounded-full border-4">
                  <Avatar className="size-full">
                    {avatarUrl && <AvatarImage src={resolveMediaUrl(avatarUrl)} alt={name} />}
                    <AvatarFallback className="text-lg font-bold">{initials}</AvatarFallback>
                  </Avatar>
                </div>
                <div className="absolute -right-1 -bottom-1">
                  <ImageUploadButton
                    iconOnly
                    label="Change Avatar"
                    onUploaded={(url) => setAvatarUrl(url)}
                    className="size-7"
                  />
                </div>
              </div>
              <p className="text-muted-foreground pb-0.5 text-xs">
                Recommended: 400×400px (JPG, PNG, WebP)
              </p>
            </div>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="name" className="text-xs font-semibold">
              Name
            </Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Your name"
              maxLength={100}
              required
            />
          </div>
        </div>

        {/* Detail blocks flow after — second column only if the first fills up */}
        <div className="mb-5 break-inside-avoid space-y-1.5">
          <div className="flex items-center justify-between">
            <Label htmlFor="bio" className="text-xs font-semibold">
              Bio
            </Label>
            <span
              className={`text-xs ${
                bio.length > 260
                  ? bio.length > 280
                    ? 'text-destructive font-bold'
                    : 'font-medium text-amber-500'
                  : 'text-muted-foreground'
              }`}
            >
              {bio.length} / 280
            </span>
          </div>
          <Textarea
            id="bio"
            value={bio}
            onChange={(e) => setBio(e.target.value)}
            placeholder="Tell the world a little about yourself"
            rows={3}
            maxLength={280}
            className="resize-none"
          />
        </div>

        <div className="mb-5 break-inside-avoid space-y-1.5">
          <Label htmlFor="location" className="text-xs font-semibold">
            Location
          </Label>
          <Input
            id="location"
            value={location}
            onChange={(e) => setLocation(e.target.value)}
            placeholder="City, Country"
            maxLength={100}
          />
        </div>

        <div className="mb-5 break-inside-avoid space-y-1.5 last:mb-0">
          <Label htmlFor="website" className="text-xs font-semibold">
            Website
          </Label>
          <Input
            id="website"
            type="url"
            value={website}
            onChange={(e) => setWebsite(e.target.value)}
            placeholder="https://example.com"
            maxLength={200}
          />
        </div>
      </div>

      <DialogFooter className="border-border mt-auto border-t px-5 py-3 sm:px-6">
        <Button type="button" variant="outline" onClick={onClose} disabled={isLoading}>
          Cancel
        </Button>
        <Button type="submit" disabled={isLoading}>
          {isLoading ? 'Saving...' : 'Save changes'}
        </Button>
      </DialogFooter>
    </form>
  );
}

export function EditProfileModal({ open, onOpenChange, profile }: EditProfileModalProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[90vh] flex-col gap-0 overflow-y-auto rounded-2xl p-0 sm:max-w-4xl md:h-[640px] lg:max-w-5xl">
        <DialogHeader className="border-border space-y-0.5 border-b px-5 py-3 sm:px-6">
          <DialogTitle className="font-heading text-xl font-extrabold">Edit profile</DialogTitle>
          <DialogDescription>Update your public profile details and photos.</DialogDescription>
        </DialogHeader>

        {open && (
          <EditProfileForm key={profile.id} profile={profile} onClose={() => onOpenChange(false)} />
        )}
      </DialogContent>
    </Dialog>
  );
}
