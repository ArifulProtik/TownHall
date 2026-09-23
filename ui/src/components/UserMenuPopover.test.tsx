import { describe, it, expect } from 'vitest';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { renderWithProviders } from '@/test/test-utils';
import { UserMenuPopover } from './UserMenuPopover';

describe('UserMenuPopover', () => {
  it('renders user avatar trigger button with online indicator', () => {
    renderWithProviders(<UserMenuPopover />);
    const trigger = screen.getByRole('button', { name: /user settings/i });
    expect(trigger).toBeInTheDocument();
  });

  it('opens menu with user profile, theme toggle and logout actions on click', async () => {
    const user = userEvent.setup();
    renderWithProviders(<UserMenuPopover />);

    const trigger = screen.getByRole('button', { name: /user settings/i });
    await user.click(trigger);

    expect(screen.getByText(/my profile/i)).toBeInTheDocument();
    expect(screen.getByText(/log out/i)).toBeInTheDocument();
    expect(screen.getByText(/theme/i)).toBeInTheDocument();
  });
});
