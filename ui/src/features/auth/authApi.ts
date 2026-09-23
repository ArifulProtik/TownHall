import { baseApi } from '@/lib/baseApi';
import { setCredentials } from '@/features/auth/authSlice';

export interface UserResponse {
  id: string;
  name: string;
  email: string;
  username?: string | null;
  provider: string;
  email_verified: boolean;
  created_at: string;
}

export interface TokenResponse {
  access_token: string;
  expires_in: number;
}

export interface SignupRequest {
  name: string;
  email: string;
  password: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface ApiErrorBody {
  error: string;
  code?: string;
  fields?: Record<string, string>;
}

export interface StatusResponse {
  status: string;
}

export interface CheckUsernameResponse {
  available: boolean;
  reason?: string;
}

export interface SetupUsernameRequest {
  username: string;
}

export const authApi = baseApi.injectEndpoints({
  endpoints: (build) => ({
    signup: build.mutation<UserResponse, SignupRequest>({
      query: (body) => ({ url: '/auth/signup', method: 'POST', body }),
    }),
    login: build.mutation<TokenResponse, LoginRequest>({
      query: (body) => ({ url: '/auth/login', method: 'POST', body }),
      async onQueryStarted(_arg, { dispatch, queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;
          dispatch(setCredentials({ accessToken: data.access_token }));
        } catch {
          // error is surfaced to the caller via .unwrap()
        }
      },
    }),
    refresh: build.query<TokenResponse, void>({
      query: () => ({ url: '/auth/refresh', method: 'POST' }),
      async onQueryStarted(_arg, { dispatch, queryFulfilled }) {
        try {
          const { data } = await queryFulfilled;
          dispatch(setCredentials({ accessToken: data.access_token }));
        } catch {
          // No usable session (missing/expired cookie); base query already cleared credentials.
        }
      },
    }),
    logout: build.mutation<StatusResponse, void>({
      query: () => ({ url: '/auth/logout', method: 'POST' }),
    }),
    logoutAll: build.mutation<StatusResponse, void>({
      query: () => ({ url: '/auth/logout-all', method: 'POST' }),
    }),
    getMe: build.query<UserResponse, void>({
      query: () => ({ url: '/auth/me', method: 'GET' }),
      providesTags: ['User'],
    }),
    checkUsername: build.query<CheckUsernameResponse, string>({
      query: (username) => ({
        url: `/auth/check-username?username=${encodeURIComponent(username)}`,
        method: 'GET',
      }),
    }),
    setupUsername: build.mutation<UserResponse, SetupUsernameRequest>({
      query: (body) => ({ url: '/auth/onboarding', method: 'POST', body }),
      async onQueryStarted(_arg, { dispatch, queryFulfilled }) {
        try {
          const { data: updatedUser } = await queryFulfilled;
          dispatch(
            authApi.util.updateQueryData('getMe', undefined, () => updatedUser),
          );
        } catch {
          // error is surfaced to the caller via .unwrap()
        }
      },
    }),
  }),
  overrideExisting: false,
});

export const {
  useSignupMutation,
  useLoginMutation,
  useRefreshQuery,
  useLogoutMutation,
  useLogoutAllMutation,
  useGetMeQuery,
  useCheckUsernameQuery,
  useSetupUsernameMutation,
} = authApi;

