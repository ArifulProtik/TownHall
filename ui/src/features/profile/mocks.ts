// MOCK DATA — replace with RTK endpoints when profile backends land; do not build on these shapes.
import type { MockPost } from '@/features/feed/mocks';

export interface MockProfile {
  handle: string;
  name: string;
  bio: string;
  followers: number;
  following: number;
}

export const mockProfiles: MockProfile[] = [
  {
    handle: 'joe',
    name: 'Joe',
    bio: 'Building TownHall one huddle at a time.',
    followers: 128,
    following: 96,
  },
  {
    handle: 'ava',
    name: 'Ava',
    bio: 'Game night organizer. Coffee first.',
    followers: 342,
    following: 210,
  },
];

export const mockTimeline: MockPost[] = [
  { id: 'p1', author: 'Joe', handle: 'joe', time: '2h', body: 'Shipping the new huddle space today. Drop by Coffee Huddle!', likes: 12, replies: 3 },
];
