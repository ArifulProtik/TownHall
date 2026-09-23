import { render, screen } from '@testing-library/react';
import { afterEach, expect, test, vi } from 'vitest';
import { Provider } from 'react-redux';
import { createMemoryRouter, RouterProvider } from 'react-router';
import { RequireGuest } from '@/features/auth/RequireGuest';
import { AuthLayout } from '@/features/auth/AuthLayout';
import LoginPage from '@/features/auth/LoginPage';
import { createAppStore } from '@/app/store';
import { markSessionChecked, setCredentials } from '@/features/auth/authSlice';

afterEach(() => {
  vi.unstubAllGlobals();
});

function renderGuestAt(path: string, { authenticated }: { authenticated: boolean }) {
  const store = createAppStore();
  // The boot-time refresh is settled before the test starts: a returning
  // session would have populated the token by now.
  if (authenticated) {
    store.dispatch(setCredentials({ accessToken: 'tok' }));
  }
  store.dispatch(markSessionChecked());
  const router = createMemoryRouter(
    [
      {
        element: (
          <RequireGuest>
            <AuthLayout />
          </RequireGuest>
        ),
        children: [{ path: '/login', element: <LoginPage /> }],
      },
      { path: '/', element: <div>Home</div> },
      { path: '/spaces', element: <div>Spaces</div> },
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

test('renders login for anonymous users after the session check', () => {
  renderGuestAt('/login', { authenticated: false });
  expect(screen.getByRole('heading', { name: 'Log in' })).toBeInTheDocument();
});

test('redirects authenticated users away from login', async () => {
  const { router } = renderGuestAt('/login', { authenticated: true });
  expect(await screen.findByText('Home')).toBeInTheDocument();
  expect(router.state.location.pathname).toBe('/');
});

test('redirects authenticated users to the next parameter', async () => {
  const { router } = renderGuestAt('/login?next=/spaces', { authenticated: true });
  expect(await screen.findByText('Spaces')).toBeInTheDocument();
  expect(router.state.location.pathname).toBe('/spaces');
});

test('renders nothing before the session check settles', () => {
  const store = createAppStore();
  const router = createMemoryRouter(
    [
      {
        element: (
          <RequireGuest>
            <AuthLayout />
          </RequireGuest>
        ),
        children: [{ path: '/login', element: <LoginPage /> }],
      },
    ],
    { initialEntries: ['/login'] },
  );
  const { container } = render(
    <Provider store={store}>
      <RouterProvider router={router} />
    </Provider>,
  );
  expect(container).toBeEmptyDOMElement();
});
