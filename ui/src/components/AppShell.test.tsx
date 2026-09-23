import { describe, it, expect } from 'vitest';
import { screen } from '@testing-library/react';
import { renderWithProviders } from '@/test/test-utils';
import { AppShell } from './AppShell';

describe('AppShell', () => {
  it('renders primary sidebar for desktop and mobile navigation bar', () => {
    renderWithProviders(<AppShell />, { route: '/' });

    expect(screen.getByRole('navigation', { name: /primary navigation rail/i })).toBeInTheDocument();
    expect(screen.getByRole('navigation', { name: /mobile primary/i })).toBeInTheDocument();
    expect(screen.getByRole('banner')).toBeInTheDocument();
  });
});
