/* eslint-disable react-refresh/only-export-components -- ThemeToggle module intentionally colocates initTheme/useTheme with the component */
import { Moon, Sun } from '@phosphor-icons/react';
import { useState } from 'react';
import { Button } from '@/components/ui/button';

const STORAGE_KEY = 'townhall-theme';

export function initTheme(): void {
  const saved = localStorage.getItem(STORAGE_KEY);
  const dark = saved ? saved === 'dark' : false;
  document.documentElement.classList.toggle('dark', dark);
}

export function useTheme(): { dark: boolean; toggle: () => void } {
  const [dark, setDark] = useState(() =>
    document.documentElement.classList.contains('dark'),
  );
  function toggle() {
    setDark((prev) => {
      const next = !prev;
      document.documentElement.classList.toggle('dark', next);
      localStorage.setItem(STORAGE_KEY, next ? 'dark' : 'light');
      return next;
    });
  }
  return { dark, toggle };
}

export function ThemeToggle() {
  const { dark, toggle } = useTheme();
  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      onClick={toggle}
      aria-label={dark ? 'Switch to light mode' : 'Switch to dark mode'}
      aria-pressed={dark}
    >
      {dark ? <Sun aria-hidden="true" /> : <Moon aria-hidden="true" />}
    </Button>
  );
}
