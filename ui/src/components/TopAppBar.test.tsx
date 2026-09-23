import { describe, it, expect } from 'vitest';
import { screen } from '@testing-library/react';
import { renderWithProviders } from '@/test/test-utils';
import { TopAppBar } from './TopAppBar';

describe('TopAppBar', () => {
  it('renders default route title and notification button', () => {
    renderWithProviders(<TopAppBar />, { route: '/' });

    expect(screen.getByRole('heading', { name: 'Home' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /notifications/i })).toBeInTheDocument();
  });

  it('renders route-specific title for /spaces', () => {
    renderWithProviders(<TopAppBar />, { route: '/spaces' });

    expect(screen.getByRole('heading', { name: 'Spaces' })).toBeInTheDocument();
  });

  it('renders route-specific title for /dms', () => {
    renderWithProviders(<TopAppBar />, { route: '/dms' });

    expect(screen.getByRole('heading', { name: 'Direct Messages' })).toBeInTheDocument();
  });

  it('allows overriding title with explicit prop', () => {
    renderWithProviders(<TopAppBar title="Custom Page" />, { route: '/' });

    expect(screen.getByRole('heading', { name: 'Custom Page' })).toBeInTheDocument();
  });
});
