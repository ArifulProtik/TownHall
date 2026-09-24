import { useState } from 'react';
import { useParams, Link } from 'react-router';
import { House, WarningCircle } from '@phosphor-icons/react';
import { useGetProfileQuery } from '@/features/profile/profileApi';
import { DocumentTitle } from '@/components/DocumentTitle';
import { ProfileHeader } from '@/features/profile/components/ProfileHeader';
import { ProfileTabs } from '@/features/profile/components/ProfileTabs';
import { EditProfileModal } from '@/features/profile/components/EditProfileModal';
import { Button } from '@/components/ui/button';

export default function ProfilePage() {
  const { handle = 'me' } = useParams<{ handle?: string }>();
  // Remount per handle so tab, header, and edit state never leak across profiles.
  return <ProfileView key={handle} handle={handle} />;
}

function ProfileView({ handle }: { handle: string }) {
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [activeTab, setActiveTab] = useState('posts');

  const {
    data: profile,
    currentData,
    isFetching,
    isError,
    error,
  } = useGetProfileQuery(handle || 'me');

  // Handle-derived title renders on first paint (no flash); upgrades to the
  // display name once the profile loads.
  const pageTitle = profile
    ? `${profile.name} | TownHall`
    : handle === 'me'
      ? 'Profile | TownHall'
      : `@${handle} | TownHall`;

  // Skeleton on first load only; background refetches keep showing content.
  if (isFetching && !currentData) {
    return (
      <div className="mx-auto w-full max-w-5xl animate-pulse">
        <DocumentTitle title={pageTitle} />
        {/* Hero skeleton */}
        <div className="overflow-hidden rounded-2xl border border-border bg-card">
          <div className="h-40 sm:h-52 w-full bg-muted/50" />
          <div className="px-4">
            <div className="flex items-end justify-between -mt-12 sm:-mt-14">
              <div className="size-24 sm:size-28 rounded-full border-4 border-card bg-muted" />
              <div className="h-8 w-24 rounded-lg bg-muted" />
            </div>
            <div className="mt-2 space-y-1">
              <div className="h-5 w-40 rounded bg-muted" />
              <div className="h-3 w-24 rounded bg-muted/70" />
              <div className="h-3 w-64 rounded bg-muted/50" />
            </div>
            <div className="mt-3 -mx-4 flex gap-6 border-t border-border px-6 pt-3">
              <div className="h-6 w-16 rounded bg-muted/60" />
              <div className="h-6 w-16 rounded bg-muted/40" />
              <div className="h-6 w-16 rounded bg-muted/40" />
              <div className="h-6 w-16 rounded bg-muted/40" />
            </div>
          </div>
        </div>
        {/* Post skeletons */}
        <div className="mt-3 grid gap-3 lg:grid-cols-[300px_minmax(0,1fr)] items-start">
          <div className="hidden lg:block rounded-xl border border-border bg-card p-4 space-y-2">
            <div className="h-4 w-24 rounded bg-muted" />
            <div className="h-3 w-full rounded bg-muted/60" />
            <div className="h-3 w-3/4 rounded bg-muted/50" />
          </div>
          <div className="w-full min-w-0 space-y-3">
          {[1, 2].map((n) => (
            <div key={n} className="rounded-xl border border-border bg-card p-3.5 space-y-2.5">
              <div className="flex items-center gap-2">
                <div className="size-8 rounded-full bg-muted" />
                <div className="space-y-1">
                  <div className="h-3 w-24 rounded bg-muted" />
                  <div className="h-2.5 w-16 rounded bg-muted/60" />
                </div>
              </div>
              <div className="space-y-1.5">
                <div className="h-3 w-full rounded bg-muted/60" />
                <div className="h-3 w-3/4 rounded bg-muted/50" />
              </div>
            </div>
          ))}
          </div>
        </div>
      </div>
    );
  }

  if (isError || !profile) {
    const apiError = error as { status?: number; data?: { error?: string } };
    const isNotFound = apiError?.status === 404;

    return (
      <div className="mx-auto flex w-full max-w-md flex-col items-center justify-center py-16 text-center">
        <DocumentTitle title={pageTitle} />
        <div className="flex size-12 items-center justify-center rounded-xl bg-destructive/10 text-destructive mb-3">
          <WarningCircle className="size-6" />
        </div>
        <h2 className="text-xl font-extrabold font-heading text-foreground">
          {isNotFound ? 'Profile not found' : 'Unable to load profile'}
        </h2>
        <p className="mt-1.5 text-sm text-muted-foreground max-w-xs">
          {isNotFound
            ? `The user @${handle} doesn't exist or may have changed their username.`
            : 'Something went wrong while fetching this profile. Please try again.'}
        </p>
        <Link to="/" className="mt-4">
          <Button variant="outline" size="sm" className="gap-1.5">
            <House className="size-3.5" />
            <span>Back to Home</span>
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="mx-auto w-full max-w-5xl pb-8">
      <DocumentTitle title={pageTitle} />
      <ProfileHeader
        profile={profile}
        onEditProfile={() => setIsEditOpen(true)}
        activeTab={activeTab}
        onSelectTab={setActiveTab}
      />

      <ProfileTabs
        profile={profile}
        activeTab={activeTab}
        onEditProfile={() => setIsEditOpen(true)}
        onSelectTab={setActiveTab}
      />

      {profile.is_self && (
        <EditProfileModal
          open={isEditOpen}
          onOpenChange={setIsEditOpen}
          profile={profile}
        />
      )}
    </div>
  );
}
