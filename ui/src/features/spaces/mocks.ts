// MOCK DATA — replace with RTK endpoints when space backends land; do not build on these shapes.
export interface MockSpace {
  id: string;
  name: string;
  members: number;
  online: number;
  rooms: string[];
}

export const mockSpaces: MockSpace[] = [
  { id: 'coffee-huddle', name: 'Coffee Huddle', members: 24, online: 6, rooms: ['lobby', 'focus-desk', 'rooftop'] },
  { id: 'game-night', name: 'Game Night', members: 58, online: 13, rooms: ['arcade', 'voice-1', 'voice-2'] },
];
