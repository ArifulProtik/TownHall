import { screen } from '@testing-library/react';
import { expect, test } from 'vitest';
import { createMemoryRouter, RouterProvider } from 'react-router';
import { Provider } from 'react-redux';
import { render } from '@testing-library/react';
import ProfilePage from '@/features/profile/ProfilePage';
import { createAppStore } from '@/app/store';

test('renders the profile header and timeline from mocks', () => {
  const store = createAppStore();
  const router = createMemoryRouter([{ path: '/u/:handle', element: <ProfilePage /> }], {
    initialEntries: ['/u/joe'],
  });
  render(
    <Provider store={store}>
      <RouterProvider router={router} />
    </Provider>,
  );
  expect(screen.getByRole('heading', { name: 'Joe' })).toBeInTheDocument();
  expect(screen.getByText('@joe')).toBeInTheDocument();
  expect(screen.getByText('Timeline')).toBeInTheDocument();
});
