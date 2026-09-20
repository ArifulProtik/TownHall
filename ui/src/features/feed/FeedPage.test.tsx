import { screen } from '@testing-library/react';
import { expect, test } from 'vitest';
import FeedPage from '@/features/feed/FeedPage';
import { mockPosts } from '@/features/feed/mocks';
import { renderWithProviders } from '@/test/test-utils';

test('renders all mock posts', () => {
  renderWithProviders(<FeedPage />, { route: '/feed' });
  for (const post of mockPosts) {
    expect(screen.getByText(post.body)).toBeInTheDocument();
  }
});
