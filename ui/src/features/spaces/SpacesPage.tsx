import { Card, CardContent } from '@/components/ui/card';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { mockSpaces } from '@/features/spaces/mocks';

export default function SpacesPage() {
  return (
    <div className="space-y-4">
      <h2 className="text-lg font-medium text-foreground">Your spaces</h2>
      <div className="grid gap-4 sm:grid-cols-2">
        {mockSpaces.map((space) => (
          <Card key={space.id}>
            <CardContent className="space-y-2">
              <div className="flex items-center gap-3">
                <Avatar>
                  <AvatarFallback>{space.name.slice(0, 2)}</AvatarFallback>
                </Avatar>
                <div>
                  <p className="font-medium text-foreground">{space.name}</p>
                  <p className="text-sm text-muted-foreground">
                    {space.online} online · {space.members} members
                  </p>
                </div>
              </div>
              <div className="flex flex-wrap gap-2">
                {space.rooms.map((room) => (
                  <Badge key={room} variant="secondary">
                    {room}
                  </Badge>
                ))}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
