import {
  createApi,
  fetchBaseQuery,
  type BaseQueryFn,
  type FetchArgs,
  type FetchBaseQueryError,
} from '@reduxjs/toolkit/query/react';
import { Mutex } from 'async-mutex';
import type { RootState } from '@/app/store';
import { clearCredentials, setCredentials } from '@/features/auth/authSlice';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';

const rawBaseQuery = fetchBaseQuery({
  baseUrl: API_BASE_URL,
  credentials: 'include',
  prepareHeaders: (headers, { getState }) => {
    const token = (getState() as RootState).auth.accessToken;
    if (token) {
      headers.set('Authorization', `Bearer ${token}`);
    }
    return headers;
  },
});

const mutex = new Mutex();

const refreshUrl = '/auth/refresh';

function requestUrl(args: string | FetchArgs): string {
  return typeof args === 'string' ? args : args.url;
}

export const baseQueryWithReauth: BaseQueryFn<string | FetchArgs, unknown, FetchBaseQueryError> =
  async (args, api, extraOptions) => {
    await mutex.waitForUnlock();
    let result = await rawBaseQuery(args, api, extraOptions);
    if (result.error?.status === 401 && requestUrl(args) !== refreshUrl) {
      if (!mutex.isLocked()) {
        const release = await mutex.acquire();
        try {
          const refresh = await rawBaseQuery(
            { url: refreshUrl, method: 'POST' },
            api,
            extraOptions,
          );
          if (refresh.data) {
            const data = refresh.data as { access_token: string };
            api.dispatch(setCredentials({ accessToken: data.access_token }));
            result = await rawBaseQuery(args, api, extraOptions);
          } else {
            api.dispatch(clearCredentials());
          }
        } finally {
          release();
        }
      } else {
        await mutex.waitForUnlock();
        result = await rawBaseQuery(args, api, extraOptions);
      }
    }
    return result;
  };

export const baseApi = createApi({
  reducerPath: 'api',
  baseQuery: baseQueryWithReauth,
  tagTypes: ['Auth', 'User', 'Follow', 'Post', 'Space', 'Message', 'Notification'],
  endpoints: () => ({}),
});

