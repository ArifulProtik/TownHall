import { describe, it, expect } from 'vitest';
import { screen } from '@testing-library/react';
import { renderWithProviders } from '@/test/test-utils';
import { PrimarySidebar } from './PrimarySidebar';

describe('PrimarySidebar', () => {
  it('renders Home, Direct Messages, Spaces from mockSpaces, and Explore button', () => {
    renderWithProviders(<PrimarySidebar />, { route: '/' });

    expect(screen.getByRole('link', { name: /home/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /direct messages/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /explore spaces/i })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /coffee huddle/i })).toBeInTheDocument();
  });

  it('renders bottom user menu avatar button', () => {
    renderWithProviders(<PrimarySidebar />);
    expect(screen.getByRole('button', { name: /user settings/i })).toBeInTheDocument();
  });
});
