import { render, screen } from '@testing-library/react';
import { expect, test } from 'vitest';
import { M3Button } from '@/components/M3Button';

test('renders children and handles loading state', () => {
  const { rerender } = render(<M3Button>Save</M3Button>);
  expect(screen.getByRole('button', { name: 'Save' })).toBeEnabled();
  rerender(<M3Button loading>Save</M3Button>);
  expect(screen.getByRole('button', { name: 'Save' })).toBeDisabled();
});
