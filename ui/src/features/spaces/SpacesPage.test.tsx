import { screen } from '@testing-library/react';
import { expect, test } from 'vitest';
import SpacesPage from '@/features/spaces/SpacesPage';
import { mockSpaces } from '@/features/spaces/mocks';
import { renderWithProviders } from '@/test/test-utils';

test('renders all mock spaces with rooms', () => {
  renderWithProviders(<SpacesPage />, { route: '/spaces' });
  for (const space of mockSpaces) {
    expect(screen.getByText(space.name)).toBeInTheDocument();
  }
  expect(screen.getByText('lobby')).toBeInTheDocument();
});
