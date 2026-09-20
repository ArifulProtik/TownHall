/* eslint-disable react-refresh/only-export-components -- ThemeToggle module intentionally colocates initTheme/useTheme with the component (consumed by Tasks 3-5) */
import { useState } from 'react';

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
    <button
      type="button"
      onClick={toggle}
      aria-label={dark ? 'Switch to light mode' : 'Switch to dark mode'}
      aria-pressed={dark}
      className="inline-flex h-10 w-10 items-center justify-center rounded-full text-on-surface-variant hover:bg-surface-container-highest"
    >
      <svg viewBox="0 0 24 24" aria-hidden="true" className="h-5 w-5" fill="currentColor">
        {dark ? (
          <path d="M12 17a5 5 0 1 0 0-10 5 5 0 0 0 0 10Zm0-15v3m0 15v3M2 12h3m15 0h3M4.9 4.9l2.1 2.1m11.4 11.4 2.1 2.1m0-14.9-2.1 2.1M7 17l-2.1 2.1" />
        ) : (
          <path d="M20.4 14.2A8.5 8.5 0 0 1 9.8 3.6a8.5 8.5 0 1 0 10.6 10.6Z" />
        )}
      </svg>
    </button>
  );
}
