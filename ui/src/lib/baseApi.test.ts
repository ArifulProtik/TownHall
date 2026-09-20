import { afterEach, expect, test, vi } from 'vitest';
import { createAppStore } from '@/app/store';
import { setCredentials } from '@/features/auth/authSlice';
import { baseQueryWithReauth } from '@/lib/baseApi';

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

test('retries once after a successful refresh and stores the new token', async () => {
  const calls: { url: string; auth: string | null }[] = [];
  vi.stubGlobal('Request', TestRequest);
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      // RTK Query 2.x passes a single Request object to fetch.
      const url = input instanceof Request ? input.url : String(input);
      const headers =
        input instanceof Request ? new Headers(input.headers) : new Headers(init?.headers);
      calls.push({ url, auth: headers.get('authorization') });
      if (url.endsWith('/auth/me')) {
        if (headers.get('authorization') === 'Bearer new-token') return json({ ok: true });
        return json({ error: 'unauthorized' }, 401);
      }
      if (url.endsWith('/auth/refresh'))
        return json({ access_token: 'new-token', expires_in: 900 });
      return json({ error: 'not found' }, 404);
    }),
  );

  const store = createAppStore();
  store.dispatch(setCredentials({ accessToken: 'expired-token' }));
  const api = { dispatch: store.dispatch, getState: store.getState } as never;
  const result = await baseQueryWithReauth('/auth/me', api, {});

  expect(result.error).toBeUndefined();
  expect(calls.map((c) => c.url)).toEqual([
    expect.stringContaining('/auth/me'),
    expect.stringContaining('/auth/refresh'),
    expect.stringContaining('/auth/me'),
  ]);
  expect(calls[2]?.auth).toBe('Bearer new-token');
  expect(store.getState().auth.accessToken).toBe('new-token');
});

test('failed refresh clears the stored token', async () => {
  vi.stubGlobal('Request', TestRequest);
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: RequestInfo | URL) => {
      const url = input instanceof Request ? input.url : String(input);
      if (url.endsWith('/auth/me')) return json({ error: 'unauthorized' }, 401);
      if (url.endsWith('/auth/refresh')) return json({ error: 'unauthorized' }, 401);
      return json({ error: 'not found' }, 404);
    }),
  );

  const store = createAppStore();
  store.dispatch(setCredentials({ accessToken: 'expired-token' }));
  const api = { dispatch: store.dispatch, getState: store.getState } as never;
  const result = await baseQueryWithReauth('/auth/me', api, {});

  expect(result.error).toBeDefined();
  expect(store.getState().auth.accessToken).toBeNull();
});
