import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, test, vi } from 'vitest';
import { Provider } from 'react-redux';
import { createMemoryRouter, RouterProvider } from 'react-router';
import LoginPage from '@/features/auth/LoginPage';
import { routes } from '@/app/router';
import { createAppStore } from '@/app/store';
import { markSessionChecked } from '@/features/auth/authSlice';
import { renderWithProviders } from '@/test/test-utils';
import { TestRequest, jsonResponse, stubFetch } from '@/test/fetchStub';

afterEach(() => {
  vi.unstubAllGlobals();
});

test('hides validation errors until a field is touched', () => {
  renderWithProviders(<LoginPage />, { route: '/login' });

  expect(screen.queryByText('Enter a valid email address.')).not.toBeInTheDocument();
  expect(
    screen.queryByText('Password must be at least 8 characters.'),
  ).not.toBeInTheDocument();
  expect(screen.getByLabelText('Email')).not.toHaveAttribute('aria-invalid');
  expect(screen.getByLabelText('Password')).not.toHaveAttribute('aria-invalid');
});

test('keeps field labels calm when validation fails', async () => {
  const user = userEvent.setup();
  renderWithProviders(<LoginPage />, { route: '/login' });

  await user.click(screen.getByRole('button', { name: 'Log in' }));

  const emailInput = screen.getByLabelText('Email');
  expect(emailInput).toHaveAttribute('aria-invalid', 'true');
  expect(emailInput.closest('[data-slot="field"]')).not.toHaveAttribute('data-invalid');
  expect(
    emailInput.closest('[data-slot="field"]')?.querySelector('label'),
  ).not.toHaveClass('text-destructive');
  expect(await screen.findByText('Enter a valid email address.')).toBeInTheDocument();
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
  store.dispatch(markSessionChecked());
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

test('successful login honors the next parameter', async () => {
  stubFetch((url) => {
    if (url.endsWith('/auth/login')) return jsonResponse({ access_token: 'tok', expires_in: 900 });
    return jsonResponse({ error: 'not found' }, 404);
  });
  const user = userEvent.setup();
  const store = createAppStore();
  store.dispatch(markSessionChecked());
  const testRouter = createMemoryRouter(routes, { initialEntries: ['/login?next=/spaces'] });
  render(
    <Provider store={store}>
      <RouterProvider router={testRouter} />
    </Provider>,
  );

  await user.type(screen.getByLabelText('Email'), 'joe@example.com');
  await user.type(screen.getByLabelText('Password'), 'password123');
  await user.click(screen.getByRole('button', { name: 'Log in' }));

  expect(testRouter.state.location.pathname).toBe('/spaces');
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
