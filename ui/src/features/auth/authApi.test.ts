import { afterEach, expect, test, vi } from 'vitest';
import { createAppStore } from '@/app/store';
import { authApi } from '@/features/auth/authApi';
import { selectAccessToken, selectIsAuthenticated } from '@/features/auth/authSlice';

const json = (data: unknown, status = 200) =>
  new Response(JSON.stringify(data), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });

// RTK Query 2.x builds `new Request(relativeUrl)` internally, and undici's
// Request (used outside the browser) rejects relative URLs. Browsers accept
// them, so resolve against a dummy origin only in tests.
const NativeRequest = globalThis.Request;
class TestRequest extends NativeRequest {
  constructor(input: RequestInfo | URL, init?: RequestInit) {
    super(
      typeof input === 'string' && input.startsWith('/')
        ? new URL(input, 'http://localhost')
        : (input as RequestInfo),
      init,
    );
  }
}

afterEach(() => {
  vi.unstubAllGlobals();
});

test('login mutation fulfillment stores the token', async () => {
  vi.stubGlobal('Request', TestRequest);
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: RequestInfo | URL) => {
      // RTK Query 2.x passes a single Request object to fetch.
      const url = input instanceof Request ? input.url : String(input);
      if (url.endsWith('/auth/login')) return json({ access_token: 'login-token', expires_in: 900 });
      return json({ error: 'not found' }, 404);
    }),
  );

  const store = createAppStore();
  expect(selectIsAuthenticated(store.getState())).toBe(false);

  const data = await store
    .dispatch(authApi.endpoints.login.initiate({ email: 'a@example.com', password: 'secret' }))
    .unwrap();

  expect(data.access_token).toBe('login-token');
  expect(selectAccessToken(store.getState())).toBe('login-token');
  expect(selectIsAuthenticated(store.getState())).toBe(true);
});

test('refresh query fulfillment restores the session', async () => {
  vi.stubGlobal('Request', TestRequest);
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: RequestInfo | URL) => {
      const url = input instanceof Request ? input.url : String(input);
      if (url.endsWith('/auth/refresh'))
        return json({ access_token: 'restored-token', expires_in: 900 });
      return json({ error: 'not found' }, 404);
    }),
  );

  const store = createAppStore();
  expect(selectIsAuthenticated(store.getState())).toBe(false);

  const data = await store.dispatch(authApi.endpoints.refresh.initiate()).unwrap();

  expect(data.access_token).toBe('restored-token');
  expect(selectAccessToken(store.getState())).toBe('restored-token');
  expect(selectIsAuthenticated(store.getState())).toBe(true);
});
