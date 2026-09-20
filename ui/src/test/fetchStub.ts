import { vi } from 'vitest';

// Shared fetch stub for RTK Query tests.
// RTK Query 2.x builds `new Request(relativeUrl)` internally, and undici's
// Request (used outside the browser) rejects relative URLs. Browsers accept
// them, so resolve against a dummy origin only in tests.
// Identical semantics to the shim in Task 5 test files (left untouched).

export function jsonResponse(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

const NativeRequest = globalThis.Request;

export class TestRequest extends NativeRequest {
  constructor(input: RequestInfo | URL, init?: RequestInit) {
    super(
      typeof input === 'string' && input.startsWith('/')
        ? new URL(input, 'http://localhost')
        : (input as RequestInfo),
      init,
    );
  }
}

export function stubFetch(handler: (url: string) => Response | Promise<Response>): void {
  vi.stubGlobal('Request', TestRequest);
  vi.stubGlobal(
    'fetch',
    vi.fn(async (input: RequestInfo | URL) => {
      // RTK Query 2.x passes a single Request object to fetch.
      const url = input instanceof Request ? input.url : String(input);
      return handler(url);
    }),
  );
}
