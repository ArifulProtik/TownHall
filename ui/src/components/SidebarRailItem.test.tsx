import { describe, it, expect } from 'vitest';
import { screen } from '@testing-library/react';
import { renderWithProviders } from '@/test/test-utils';
import { SidebarRailItem } from './SidebarRailItem';
import { House } from '@phosphor-icons/react';

describe('SidebarRailItem', () => {
  it('renders link with aria-label and tooltip text', () => {
    renderWithProviders(
      <SidebarRailItem to="/test" label="Test Item" icon={<House data-testid="icon" />} />,
      { route: '/' }
    );

    const link = screen.getByRole('link', { name: /test item/i });
    expect(link).toHaveAttribute('href', '/test');
    expect(screen.getByTestId('icon')).toBeInTheDocument();
  });

  it('renders active indicator pill when route matches', () => {
    const { container } = renderWithProviders(
      <SidebarRailItem to="/test" label="Test Item" initials="TI" />,
      { route: '/test' }
    );

    const pill = container.querySelector('[data-slot="rail-pill"]');
    expect(pill).toBeInTheDocument();
    expect(pill).toHaveClass('h-10');
  });

  it('renders badge count when provided', () => {
    renderWithProviders(
      <SidebarRailItem to="/test" label="Test Item" initials="TI" badgeCount={3} />
    );

    expect(screen.getByText('3')).toBeInTheDocument();
  });
});
