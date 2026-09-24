import { useState } from 'react';
import {
  CalendarBlank,
  Envelope,
  Globe,
  MapPin,
  ShieldCheck,
  Briefcase,
  MagnifyingGlass,
} from '@phosphor-icons/react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { CreatePostPrompt } from '@/features/profile/components/CreatePostPrompt';
import { ProfileIntroCard } from '@/features/profile/components/ProfileIntroCard';
import { FollowList } from '@/features/social/FollowList';
import {
  TimelinePostCard,
  type TimelinePost,
} from '@/features/profile/components/TimelinePostCard';
import type { ProfileResponse } from '@/features/profile/profileApi';

interface ProfileTabsProps {
  profile: ProfileResponse;
  activeTab: string;
  onEditProfile: () => void;
  onSelectTab: (tab: string) => void;
}

export function ProfileTabs({ profile, activeTab, onEditProfile, onSelectTab }: ProfileTabsProps) {
  const [posts, setPosts] = useState<TimelinePost[]>([
    {
      id: 'p1',
      authorName: profile.name || 'Member',
      authorHandle: profile.username ? `@${profile.username}` : '@member',
      authorAvatar: profile.avatar_url,
      time: '2 hours ago',
      content:
        profile.bio ||
        'Welcome to my TownHall space! Looking forward to collaborating on community tools and modern web apps.',
      likesCount: 38,
      commentsCount: 6,
      sharesCount: 2,
      comments: [
        {
          id: 'c1',
          authorName: 'Marcus Vance',
          authorAvatar:
            'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&w=150&q=80',
          text: 'Great having you here! The new space architecture looks fantastic.',
          time: '1h ago',
          likes: 3,
        },
      ],
    },
    {
      id: 'p2',
      authorName: profile.name || 'Member',
      authorHandle: profile.username ? `@${profile.username}` : '@member',
      authorAvatar: profile.avatar_url,
      time: 'Yesterday at 4:15 PM',
      content:
        'Just wrapped up an incredible town hall session discussing modern web architecture and real-time community tools. Check out our design previews from the session!',
      image:
        'https://images.unsplash.com/photo-1550745165-9bc0b252726f?auto=format&fit=crop&w=800&q=80',
      likesCount: 84,
      commentsCount: 15,
      sharesCount: 7,
      comments: [
        {
          id: 'c2',
          authorName: 'Sarah Chen',
          authorAvatar:
            'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=150&q=80',
          text: 'The recording and notes from this session were super insightful!',
          time: '18h ago',
          likes: 5,
        },
      ],
    },
    {
      id: 'p3',
      authorName: profile.name || 'Member',
      authorHandle: profile.username ? `@${profile.username}` : '@member',
      authorAvatar: profile.avatar_url,
      time: '3 days ago',
      content:
        'Always inspired by how fast this community is growing. Building in public has never been more rewarding! 🚀',
      likesCount: 129,
      commentsCount: 24,
      sharesCount: 11,
    },
  ]);

  const galleryPhotos = [
    'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&w=600&q=80',
    'https://images.unsplash.com/photo-1579783900882-c0d3dad7b119?auto=format&fit=crop&w=600&q=80',
    'https://images.unsplash.com/photo-1607604276583-eef5d076aa5f?auto=format&fit=crop&w=600&q=80',
    'https://images.unsplash.com/photo-1550745165-9bc0b252726f?auto=format&fit=crop&w=600&q=80',
    'https://images.unsplash.com/photo-1526374965328-7f61d4dc18c5?auto=format&fit=crop&w=600&q=80',
    'https://images.unsplash.com/photo-1504384308090-c894fdcc538d?auto=format&fit=crop&w=600&q=80',
    'https://images.unsplash.com/photo-1451187580459-43490279c0fa?auto=format&fit=crop&w=600&q=80',
    'https://images.unsplash.com/photo-1518770660439-4636190af475?auto=format&fit=crop&w=600&q=80',
    'https://images.unsplash.com/photo-1531482615713-2afd69097998?auto=format&fit=crop&w=600&q=80',
  ];

  function handleCreatePost(newContent: string, newImage?: string) {    const newPost: TimelinePost = {
      id: `p_${Date.now()}`,
      authorName: profile.name || 'Member',
      authorHandle: profile.username ? `@${profile.username}` : '@member',
      authorAvatar: profile.avatar_url,
      time: 'Just now',
      content: newContent,
      image: newImage,
      likesCount: 0,
      commentsCount: 0,
      sharesCount: 0,
    };
    setPosts((prev) => [newPost, ...prev]);
  }

  return (
    <div className="mt-3">
      {/* ────── Posts Tab ────── */}
      {activeTab === 'posts' && (
        <div className="grid items-start gap-3 lg:grid-cols-[300px_minmax(0,1fr)]">
          <aside className="order-2 min-w-0 lg:sticky lg:top-3 lg:order-1">
            <ProfileIntroCard profile={profile} onNavigateTab={onSelectTab} />
          </aside>
          <div className="order-1 w-full min-w-0 space-y-3 lg:order-2">
            {profile.is_self && (
              <CreatePostPrompt profile={profile} onPostCreated={handleCreatePost} />
            )}
            {posts.map((post) => (
              <TimelinePostCard key={post.id} post={post} currentProfile={profile} />
            ))}
          </div>
        </div>
      )}

      {/* ────── About Tab ────── */}
      {activeTab === 'about' && (
        <div className="space-y-3">
          <div className="flex items-center justify-between gap-2">
            <h2 className="text-foreground text-xl font-extrabold tracking-tight">
              About {profile.name}
            </h2>
            {profile.is_self && (
              <Button onClick={onEditProfile} variant="outline" size="sm" className="font-medium">
                Edit profile
              </Button>
            )}
          </div>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {/* Facts */}
            <section className="border-border bg-card rounded-2xl border p-4">
              <div className="text-foreground/80 space-y-2.5 text-sm">
                <div className="flex items-center gap-2">
                  <Briefcase className="text-muted-foreground/80 size-4 shrink-0" />
                  <span>
                    Community Member at{' '}
                    <strong className="text-foreground font-semibold">TownHall</strong>
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <CalendarBlank className="text-muted-foreground/80 size-4 shrink-0" />
                  <span>
                    Joined{' '}
                    {new Date(profile.created_at).toLocaleDateString('en-US', {
                      month: 'long',
                      day: 'numeric',
                      year: 'numeric',
                    })}
                  </span>
                </div>
                {profile.location && (
                  <div className="flex items-center gap-2">
                    <MapPin className="text-muted-foreground/80 size-4 shrink-0" />
                    <span>{profile.location}</span>
                  </div>
                )}
                {profile.email_verified && (
                  <div className="flex items-center gap-2">
                    <ShieldCheck className="size-4 shrink-0 text-emerald-500" weight="fill" />
                    <span className="font-medium text-emerald-600 dark:text-emerald-400">
                      Verified identity
                    </span>
                  </div>
                )}
              </div>
            </section>

            {/* Bio */}
            <section className="border-border bg-card rounded-2xl border p-4">
              {profile.bio ? (
                <p className="text-foreground/85 text-base/5">{profile.bio}</p>
              ) : profile.is_self ? (
                <Button
                  variant="secondary"
                  onClick={onEditProfile}
                  className="w-full text-sm font-medium"
                >
                  Add bio
                </Button>
              ) : (
                <p className="text-muted-foreground text-sm">No bio yet.</p>
              )}
            </section>

            {/* Contact */}
            <section className="border-border bg-card rounded-2xl border p-4">
              {profile.email || profile.website ? (
                <div className="space-y-2.5 text-sm">
                  {profile.email && (
                    <div className="flex items-center gap-2">
                      <Envelope className="text-muted-foreground/80 size-4 shrink-0" />
                      <span className="text-foreground truncate">{profile.email}</span>
                    </div>
                  )}
                  {profile.website && (
                    <div className="flex items-center gap-2">
                      <Globe className="text-muted-foreground/80 size-4 shrink-0" />
                      <a
                        href={
                          profile.website.startsWith('http')
                            ? profile.website
                            : `https://${profile.website}`
                        }
                        target="_blank"
                        rel="noreferrer noopener"
                        className="text-primary truncate font-medium hover:underline"
                      >
                        {profile.website}
                      </a>
                    </div>
                  )}
                </div>
              ) : (
                <p className="text-muted-foreground text-sm">No contact info yet.</p>
              )}
            </section>
          </div>
        </div>
      )}

      {/* ────── Friends Tab ────── */}
      {activeTab === 'community' && (
        <CommunityTab handle={profile.username || profile.id} />
      )}

      {/* ────── Photos Tab ────── */}
      {activeTab === 'photos' && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-foreground text-xl font-extrabold tracking-tight">Photos</h2>
            <Badge variant="secondary">{galleryPhotos.length} photos</Badge>
          </div>

          <div className="grid grid-cols-3 gap-1.5 lg:grid-cols-4">
            {galleryPhotos.map((src, i) => (
              <div
                key={i}
                className="group bg-muted aspect-square cursor-pointer overflow-hidden rounded-lg"
              >
                <img
                  src={src}
                  alt=""
                  className="size-full object-cover transition-transform duration-200 group-hover:scale-105"
                />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function CommunityTab({ handle }: { handle: string }) {
  const [search, setSearch] = useState('');

  return (
    <div className="space-y-3">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <h2 className="text-foreground text-xl font-extrabold tracking-tight">Friends</h2>
        <div className="relative w-full sm:w-48">
          <MagnifyingGlass className="text-muted-foreground absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="bg-muted text-foreground placeholder:text-muted-foreground focus:ring-ring w-full rounded-lg py-1.5 pr-3 pl-8 text-base focus:ring-1 focus:outline-hidden md:text-sm"
          />
        </div>
      </div>

      <div className="border-border bg-card rounded-2xl border px-4 py-1">
        <FollowList handle={handle} tab="friends" filter={search} />
      </div>
    </div>
  );
}
