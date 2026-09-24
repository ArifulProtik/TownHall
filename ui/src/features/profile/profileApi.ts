import { baseApi } from '@/lib/baseApi';

export interface ProfileResponse {
  id: string;
  name: string;
  email?: string;
  username?: string | null;
  provider: string;
  email_verified: boolean;
  bio?: string;
  avatar_url?: string;
  banner_url?: string;
  location?: string;
  website?: string;
  created_at: string;
  followers_count: number;
  following_count: number;
  posts_count: number;
  is_self: boolean;
}

export interface UpdateProfileRequest {
  name?: string;
  bio?: string;
  avatar_url?: string;
  banner_url?: string;
  location?: string;
  website?: string;
}

export interface UploadResponse {
  url: string;
  key?: string;
  name?: string;
  size?: number;
}

export const profileApi = baseApi.injectEndpoints({
  endpoints: (build) => ({
    getProfile: build.query<ProfileResponse, string>({
      query: (handle) => ({
        url: `/users/${encodeURIComponent(handle)}`,
        method: 'GET',
      }),
      providesTags: (_result, _error, handle) => [{ type: 'User', id: handle }],
    }),
    updateProfile: build.mutation<ProfileResponse, UpdateProfileRequest>({
      query: (body) => ({
        url: '/profile',
        method: 'PATCH',
        body,
      }),
      invalidatesTags: ['User'],
    }),
    uploadFile: build.mutation<UploadResponse, FormData>({
      query: (body) => ({
        url: '/profile/upload',
        method: 'POST',
        body,
      }),
    }),
  }),
  overrideExisting: false,
});

export const {
  useGetProfileQuery,
  useUpdateProfileMutation,
  useUploadFileMutation,
} = profileApi;
