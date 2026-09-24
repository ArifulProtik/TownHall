import * as React from 'react';
import { Link } from 'react-router';
import {
  Gear,
  PencilSimple,
  ShareNetwork,
  PaperPlaneTilt,
  Camera,
  UserPlus,
  UserCheck,
  MapPin,
  LinkSimple,
  CalendarBlank,
} from '@phosphor-icons/react';
import { Button } from '@/components/ui/button';
import { resolveMediaUrl } from '@/lib/media';
import type { ProfileResponse } from '@/features/profile/profileApi';

interface ProfileHeaderProps {
  profile: ProfileResponse;
  onEditProfile: () => void;
  activeTab: string;
  onSelectTab: (tab: string) => void;
}

export function ProfileHeader({
  profile,
  onEditProfile,
  activeTab,
  onSelectTab,
}: ProfileHeaderProps) {
  const [following, setFollowing] = React.useState(false);
  const [copied, setCopied] = React.useState(false);
  const [bannerError, setBannerError] = React.useState(false);

  const initials = profile.name ? profile.name.slice(0, 2).toUpperCase() : 'TH';

  const joinDate = profile.created_at
    ? new Date(profile.created_at).toLocaleDateString('en-US', { month: 'short', year: 'numeric' })
    : 'Recently';

  function handleShare() {
    if (navigator.clipboard) {
      navigator.clipboard.writeText(window.location.href);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  }

  const tabs = [
    { id: 'posts', label: 'Posts' },
    { id: 'about', label: 'About' },
    { id: 'community', label: 'Friends' },
    { id: 'photos', label: 'Photos' },
  ];

  return (
    <div className="overflow-hidden rounded-2xl border border-border bg-card">
      {/* Cover */}
      <div className="relative h-40 sm:h-52 w-full overflow-hidden bg-gradient-to-br from-[#2f6bff] via-[#6d5cff] to-[#a855f7]">
        {profile.banner_url && !bannerError ? (
          <img
            src={resolveMediaUrl(profile.banner_url)}
            alt=""
            onError={() => setBannerError(true)}
            className="size-full object-cover"
          />
        ) : (
          <>
            <div className="absolute inset-0 bg-gradient-to-br from-[#2f6bff] via-[#6d5cff] to-[#a855f7]" />
            <div className="absolute inset-0 bg-[radial-gradient(560px_circle_at_15%_0%,rgba(255,255,255,0.32),transparent_62%)]" />
            <div className="absolute inset-0 bg-[radial-gradient(520px_circle_at_88%_108%,rgba(255,186,120,0.38),transparent_65%)]" />
            <div className="absolute inset-0 bg-gradient-to-t from-black/50 via-black/5 to-transparent" />
          </>
        )}
        {profile.banner_url && !bannerError && (
          <div className="absolute inset-0 bg-gradient-to-t from-black/50 via-black/5 to-transparent" />
        )}

        {profile.is_self && (
          <button
            type="button"
            onClick={onEditProfile}
            className="absolute top-3 right-3 flex items-center gap-1.5 rounded-full bg-black/45 px-2.5 py-1 text-xs font-medium text-white backdrop-blur-md hover:bg-black/60 transition-colors cursor-pointer"
          >
            <Camera className="size-3.5" />
            <span className="hidden sm:inline">Edit cover</span>
          </button>
        )}
      </div>

      {/* Profile Info — overlapping the cover bottom */}
      <div className="relative px-4">
        <div className="-mt-12 sm:-mt-14 flex flex-col gap-2.5">
          {/* Top row: avatar + actions */}
          <div className="flex items-end justify-between gap-3">
            {/* Avatar: circle like every other avatar on the page */}
            <div className="relative">
              <div className="flex size-24 sm:size-28 items-center justify-center overflow-hidden rounded-full border-4 border-card bg-muted">
                {profile.avatar_url ? (
                  <img
                    src={resolveMediaUrl(profile.avatar_url)}
                    alt={profile.name}
                    className="size-full object-cover"
                  />
                ) : (
                  <span className="text-xl sm:text-2xl font-bold font-heading text-muted-foreground">
                    {initials}
                  </span>
                )}
              </div>
              <span
                aria-hidden="true"
                className="absolute bottom-1.5 right-1.5 size-4 rounded-full bg-emerald-500 ring-2 ring-card"
              />
              {profile.is_self && (
                <button
                  type="button"
                  onClick={onEditProfile}
                  aria-label="Edit profile photo"
                  className="absolute -bottom-1 -right-1 flex size-7 items-center justify-center rounded-full bg-primary text-primary-foreground shadow hover:bg-primary/90 transition-colors cursor-pointer"
                >
                  <Camera className="size-3.5" />
                </button>
              )}
            </div>

            {/* Action buttons */}
            <div className="flex items-center gap-1.5 pb-0.5">
              {profile.is_self ? (
                <>
                  <Button onClick={onEditProfile} size="sm" className="gap-1.5 font-medium">
                    <PencilSimple className="size-3.5" />
                    <span>Edit profile</span>
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="icon-sm"
                    aria-label="Share profile"
                    onClick={handleShare}
                    className="relative"
                  >
                    <ShareNetwork className="size-3.5" />
                    {copied && (
                      <span className="absolute -top-8 whitespace-nowrap rounded bg-foreground px-2 py-0.5 text-xs text-background shadow">
                        Copied!
                      </span>
                    )}
                  </Button>
                  <Link to="/settings">
                    <Button type="button" variant="outline" size="icon-sm" aria-label="Settings">
                      <Gear className="size-3.5" />
                    </Button>
                  </Link>
                </>
              ) : (
                <>
                  <Button
                    size="sm"
                    variant={following ? 'outline' : 'default'}
                    onClick={() => setFollowing(!following)}
                    className="gap-1.5 font-medium"
                  >
                    {following ? (
                      <><UserCheck className="size-3.5" /><span>Following</span></>
                    ) : (
                      <><UserPlus className="size-3.5" /><span>Follow</span></>
                    )}
                  </Button>
                  <Link to="/dms">
                    <Button variant="outline" size="sm" className="gap-1.5 font-medium">
                      <PaperPlaneTilt className="size-3.5" />
                      <span>Message</span>
                    </Button>
                  </Link>
                </>
              )}
            </div>
          </div>

          {/* Name + bio + meta + stats, each on its own line like X */}
          <div className="space-y-1.5">
            <div className="space-y-0">
              <h1 className="text-xl font-extrabold font-heading tracking-[-0.02em] text-foreground leading-tight">
                {profile.name}
              </h1>
              <p className="text-base leading-snug text-muted-foreground">
                @{profile.username || 'member'}
              </p>
            </div>

            {profile.bio && (
              <p className="text-base/5 text-foreground/85 max-w-2xl">
                {profile.bio}
              </p>
            )}

            <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-base text-muted-foreground pt-0.5">
              {profile.location && (
                <span className="inline-flex items-center gap-1">
                  <MapPin className="size-4" weight="fill" />
                  {profile.location}
                </span>
              )}
              {profile.website && (
                <a
                  href={profile.website.startsWith('http') ? profile.website : `https://${profile.website}`}
                  target="_blank"
                  rel="noreferrer noopener"
                  className="inline-flex items-center gap-1 text-primary hover:underline font-medium"
                >
                  <LinkSimple className="size-4" />
                  {profile.website.replace(/^https?:\/\//, '')}
                </a>
              )}
              <span className="inline-flex items-center gap-1">
                <CalendarBlank className="size-4" />
                Joined {joinDate}
              </span>
            </div>

            <div className="flex flex-wrap items-center gap-x-4 text-base">
              <span>
                <strong className="text-foreground font-bold">{profile.following_count || 142}</strong>{' '}
                <span className="text-muted-foreground">Following</span>
              </span>
              <span>
                <strong className="text-foreground font-bold">{profile.followers_count || 489}</strong>{' '}
                <span className="text-muted-foreground">Followers</span>
              </span>
            </div>
          </div>
        </div>

        {/* Tab row: X-style compact tabs, left-aligned, flush to card edge */}
        <nav className="mt-3 -mx-4 flex items-center gap-1 overflow-x-auto border-t border-border px-2 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
          {tabs.map((tab) => {
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                type="button"
                onClick={() => onSelectTab(tab.id)}
                aria-current={isActive ? 'page' : undefined}
                className={`relative whitespace-nowrap px-4 py-3 text-base transition-colors cursor-pointer ${
                  isActive
                    ? 'font-bold text-foreground'
                    : 'font-medium text-muted-foreground hover:text-foreground'
                }`}
              >
                {tab.label}
                {isActive && (
                  <span className="absolute inset-x-4 bottom-0 h-1 rounded-full bg-primary" />
                )}
              </button>
            );
          })}
        </nav>
      </div>
    </div>
  );
}
