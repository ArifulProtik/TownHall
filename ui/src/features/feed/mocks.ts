// MOCK DATA — replace with RTK endpoints when feed backends land; do not build on these shapes.
export interface MockPost {
  id: string;
  author: string;
  handle: string;
  time: string;
  body: string;
  likes: number;
  replies: number;
}

export const mockPosts: MockPost[] = [
  {
    id: 'p1',
    author: 'Joe',
    handle: 'joe',
    time: '2h',
    body: 'Shipping the new huddle space today. Drop by Coffee Huddle!',
    likes: 12,
    replies: 3,
  },
  {
    id: 'p2',
    author: 'Ava',
    handle: 'ava',
    time: '5h',
    body: 'Game night was legendary. Same time next week?',
    likes: 30,
    replies: 11,
  },
];
