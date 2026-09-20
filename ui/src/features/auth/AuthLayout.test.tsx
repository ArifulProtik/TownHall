import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { expect, test } from 'vitest';
import { createMemoryRouter, RouterProvider } from 'react-router';
import { Provider } from 'react-redux';
import { AuthLayout } from '@/features/auth/AuthLayout';
import LoginPage from '@/features/auth/LoginPage';
import SignupPage from '@/features/auth/SignupPage';
import { createAppStore } from '@/app/store';

function renderAuthAt(path: string) {
  const store = createAppStore();
  const router = createMemoryRouter(
    [
      {
        element: <AuthLayout />,
        children: [
          { path: '/login', element: <LoginPage /> },
          { path: '/signup', element: <SignupPage /> },
        ],
      },
    ],
    { initialEntries: [path] },
  );
  render(
    <Provider store={store}>
      <RouterProvider router={router} />
    </Provider>,
  );
  return { store, router };
}

test('renders the login card inside the animated layout', () => {
  renderAuthAt('/login');
  expect(screen.getByRole('heading', { name: 'Log in to TownHall' })).toBeInTheDocument();
});

test('navigates login to signup with the layout intact', async () => {
  const user = userEvent.setup();
  const { router } = renderAuthAt('/login');
  await user.click(screen.getByRole('link', { name: 'Sign up' }));
  expect(await screen.findByRole('heading', { name: 'Create your account' })).toBeInTheDocument();
  expect(router.state.location.pathname).toBe('/signup');
});
