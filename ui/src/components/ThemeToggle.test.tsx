import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, expect, test } from 'vitest';
import { initTheme, ThemeToggle } from '@/components/ThemeToggle';

beforeEach(() => {
  localStorage.clear();
  document.documentElement.classList.remove('dark');
});

test('initTheme defaults to light without a saved preference', () => {
  initTheme();
  expect(document.documentElement.classList.contains('dark')).toBe(false);
});

test('toggle flips the dark class and persists it', async () => {
  const user = userEvent.setup();
  render(<ThemeToggle />);
  await user.click(screen.getByRole('button', { name: 'Switch to dark mode' }));
  expect(document.documentElement.classList.contains('dark')).toBe(true);
  expect(localStorage.getItem('townhall-theme')).toBe('dark');
});
