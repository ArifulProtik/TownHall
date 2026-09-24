import { useLayoutEffect, useRef } from 'react';

export const APP_NAME = 'TownHall';

function toTitle(title: string | undefined, fallback: string): string {
  if (title && title.trim() !== '') {
    return title;
  }
  return fallback;
}

/**
 * Sets document.title, synchronously before paint (useLayoutEffect) so the
 * tab never flashes a stale value. Restores the previous title on unmount.
 */
export function useDocumentTitle(title: string | undefined, fallback = APP_NAME) {
  const previous = useRef<string | null>(null);

  useLayoutEffect(() => {
    if (previous.current === null) {
      previous.current = document.title;
    }
    document.title = toTitle(title, fallback);
  }, [title, fallback]);

  useLayoutEffect(() => {
    return () => {
      if (previous.current !== null) {
        document.title = previous.current;
      }
    };
  }, []);
}
