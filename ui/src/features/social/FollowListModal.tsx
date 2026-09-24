import * as React from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { FollowList, type FollowListTab } from '@/features/social/FollowList';

export type { FollowListTab };

interface FollowListModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  handle: string;
  initialTab: FollowListTab;
}

const TAB_LABELS: Record<FollowListTab, string> = {
  followers: 'Followers',
  following: 'Following',
  friends: 'Friends',
};

export function FollowListModal({ open, onOpenChange, handle, initialTab }: FollowListModalProps) {
  // Parent remounts per opened tab (key), so the initializer is enough.
  const [tab, setTab] = React.useState<FollowListTab>(initialTab);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{TAB_LABELS[tab]}</DialogTitle>
        </DialogHeader>
        <div className="flex gap-1 border-b border-border">
          {(Object.keys(TAB_LABELS) as FollowListTab[]).map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => setTab(t)}
              className={`px-3 py-2 text-sm transition-colors cursor-pointer ${
                tab === t
                  ? 'font-bold text-foreground border-b-2 border-primary -mb-px'
                  : 'font-medium text-muted-foreground hover:text-foreground'
              }`}
            >
              {TAB_LABELS[t]}
            </button>
          ))}
        </div>
        {open && (
          <FollowList
            key={`${handle}-${tab}`}
            handle={handle}
            tab={tab}
            onNavigate={() => onOpenChange(false)}
          />
        )}
      </DialogContent>
    </Dialog>
  );
}
