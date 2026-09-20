import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, test, vi } from 'vitest';
import { Provider } from 'react-redux';
import { createMemoryRouter, RouterProvider } from 'react-router';
import SignupPage from '@/features/auth/SignupPage';
import LoginPage from '@/features/auth/LoginPage';
import { createAppStore } from '@/app/store';
import { renderWithProviders } from '@/test/test-utils';
import { jsonResponse, stubFetch } from '@/test/fetchStub';

afterEach(() => {
  vi.unstubAllGlobals();
});

test('successful signup navigates to login', async () => {
  stubFetch((url) => {
    if (url.endsWith('/auth/signup')) {
      return jsonResponse(
        {
          id: '1',
          name: 'Joe',
          email: 'joe@example.com',
          provider: 'email',
          email_verified: false,
          created_at: '2026-09-20T00:00:00Z',
        },
        201,
      );
    }
    return jsonResponse({ error: 'not found' }, 404);
  });
  const user = userEvent.setup();
  const store = createAppStore();
  const testRouter = createMemoryRouter(
    [
      { path: '/signup', element: <SignupPage /> },
      { path: '/login', element: <LoginPage /> },
    ],
    { initialEntries: ['/signup'] },
  );
  render(
    <Provider store={store}>
      <RouterProvider router={testRouter} />
    </Provider>,
  );

  await user.type(screen.getByLabelText('Name'), 'Joe');
  await user.type(screen.getByLabelText('Email'), 'joe@example.com');
  await user.type(screen.getByLabelText('Password'), 'password123');
  await user.click(screen.getByRole('button', { name: 'Sign up' }));

  expect(testRouter.state.location.pathname).toBe('/login');
});

test('409 shows already registered', async () => {
  stubFetch(() => jsonResponse({ error: 'email already registered' }, 409));
  const user = userEvent.setup();
  renderWithProviders(<SignupPage />, { route: '/signup' });

  await user.type(screen.getByLabelText('Name'), 'Joe');
  await user.type(screen.getByLabelText('Email'), 'joe@example.com');
  await user.type(screen.getByLabelText('Password'), 'password123');
  await user.click(screen.getByRole('button', { name: 'Sign up' }));

  expect(await screen.findByRole('alert')).toHaveTextContent('This email is already registered.');
});
