import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, test, vi } from 'vitest';
import { Provider } from 'react-redux';
import { createMemoryRouter, RouterProvider } from 'react-router';
import LoginPage from '@/features/auth/LoginPage';
import { routes } from '@/app/router';
import { createAppStore } from '@/app/store';
import { renderWithProviders } from '@/test/test-utils';
import { TestRequest, jsonResponse, stubFetch } from '@/test/fetchStub';

afterEach(() => {
  vi.unstubAllGlobals();
});

test('schema validation blocks a bad email without network', async () => {
  vi.stubGlobal('Request', TestRequest);
  const fetchMock = vi.fn(async () => jsonResponse({}));
  vi.stubGlobal('fetch', fetchMock);
  const user = userEvent.setup();
  renderWithProviders(<LoginPage />, { route: '/login' });

  await user.type(screen.getByLabelText('Email'), 'not-an-email');
  await user.type(screen.getByLabelText('Password'), 'password123');
  await user.click(screen.getByRole('button', { name: 'Log in' }));

  expect(await screen.findByText('Enter a valid email address.')).toBeInTheDocument();
  expect(fetchMock).not.toHaveBeenCalled();
});

test('successful login navigates home', async () => {
  stubFetch((url) => {
    if (url.endsWith('/auth/login')) return jsonResponse({ access_token: 'tok', expires_in: 900 });
    return jsonResponse({ error: 'not found' }, 404);
  });
  const user = userEvent.setup();
  const store = createAppStore();
  const testRouter = createMemoryRouter(routes, { initialEntries: ['/login'] });
  render(
    <Provider store={store}>
      <RouterProvider router={testRouter} />
    </Provider>,
  );

  await user.type(screen.getByLabelText('Email'), 'joe@example.com');
  await user.type(screen.getByLabelText('Password'), 'password123');
  await user.click(screen.getByRole('button', { name: 'Log in' }));

  expect(testRouter.state.location.pathname).toBe('/');
});

test('401 shows invalid credentials', async () => {
  stubFetch(() => jsonResponse({ error: 'invalid credentials' }, 401));
  const user = userEvent.setup();
  renderWithProviders(<LoginPage />, { route: '/login' });

  await user.type(screen.getByLabelText('Email'), 'joe@example.com');
  await user.type(screen.getByLabelText('Password'), 'password123');
  await user.click(screen.getByRole('button', { name: 'Log in' }));

  expect(await screen.findByRole('alert')).toHaveTextContent('Invalid email or password.');
});
