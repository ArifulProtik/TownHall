import { Card, CardContent } from '@/components/ui/card';
import { ChatCircleDots } from '@phosphor-icons/react';

export default function MessagesPage() {
  return (
    <div className="flex h-full min-h-[400px] items-center justify-center p-6">
      <Card className="max-w-md text-center">
        <CardContent className="space-y-4 pt-6">
          <div className="mx-auto flex size-14 items-center justify-center rounded-2xl bg-primary/10 text-primary">
            <ChatCircleDots className="size-8" weight="duotone" />
          </div>
          <h2 className="text-xl font-bold text-card-foreground">Direct Messages</h2>
          <p className="text-sm text-muted-foreground">
            Start a conversation or message your friends directly in TownHall.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
