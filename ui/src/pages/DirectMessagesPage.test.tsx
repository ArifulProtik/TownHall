import { describe, it, expect } from 'vitest';
import { screen } from '@testing-library/react';
import { renderWithProviders } from '@/test/test-utils';
import DirectMessagesPage from './DirectMessagesPage';

describe('DirectMessagesPage', () => {
  it('renders direct messages placeholder heading and description', () => {
    renderWithProviders(<DirectMessagesPage />);
    expect(screen.getByRole('heading', { name: /direct messages/i })).toBeInTheDocument();
    expect(screen.getByText(/start a conversation or message your friends/i)).toBeInTheDocument();
  });
});
