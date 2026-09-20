import { M3Card } from '@/components/M3Card';
import { Avatar } from '@/components/Avatar';
import { mockSpaces } from '@/features/spaces/mocks';

export default function SpacesPage() {
  return (
    <div className="space-y-4">
      <h2 className="text-lg font-medium text-on-surface">Your spaces</h2>
      <div className="grid gap-4 sm:grid-cols-2">
        {mockSpaces.map((space) => (
          <M3Card key={space.id} className="space-y-2 p-4">
            <div className="flex items-center gap-3">
              <Avatar initials={space.name.slice(0, 2)} />
              <div>
                <p className="font-medium text-on-surface">{space.name}</p>
                <p className="text-sm text-on-surface-variant">
                  {space.online} online · {space.members} members
                </p>
              </div>
            </div>
            <div className="flex flex-wrap gap-2">
              {space.rooms.map((room) => (
                <span
                  key={room}
                  className="rounded-m3-md bg-secondary-container px-3 py-1 text-xs text-on-secondary-container"
                >
                  {room}
                </span>
              ))}
            </div>
          </M3Card>
        ))}
      </div>
    </div>
  );
}
