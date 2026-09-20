import type { FetchBaseQueryError, FetchBaseQueryMeta } from '@reduxjs/toolkit/query';
import type { ApiErrorBody } from '@/features/auth/authApi';

// RTK 2.x does not declare `meta` on the FetchBaseQueryError union, but
// rejected query values carry the base-query meta at runtime.
type QueryErrorWithMeta = FetchBaseQueryError & { meta?: FetchBaseQueryMeta };

function retryAfterSeconds(error: FetchBaseQueryError): number | null {
  const headers = (error as QueryErrorWithMeta).meta?.response?.headers;
  const raw = headers?.get('Retry-After');
  if (!raw) return null;
  const secs = Number.parseInt(raw, 10);
  return Number.isFinite(secs) && secs > 0 ? secs : null;
}

export function getAuthErrorMessage(err: unknown): string {
  if (err && typeof err === 'object' && 'status' in err) {
    const error = err as FetchBaseQueryError;
    const data = error.data as ApiErrorBody | undefined;
    if (error.status === 401) return 'Invalid email or password.';
    if (error.status === 409) return 'This email is already registered.';
    if (error.status === 429) {
      const secs = retryAfterSeconds(error);
      return secs
        ? `Too many attempts. Try again in ${secs}s.`
        : 'Too many attempts. Try again later.';
    }
    if (data?.fields) {
      const first = Object.entries(data.fields)[0];
      if (first) return `${first[0]}: ${first[1]}`;
    }
    if (data?.error) return data.error;
    return 'Something went wrong. Please try again.';
  }
  return 'Something went wrong. Please try again.';
}
