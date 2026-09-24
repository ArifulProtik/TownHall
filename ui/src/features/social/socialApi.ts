import { baseApi } from '@/lib/baseApi';

export interface FollowStatus {
  is_following: boolean;
  is_followed_by: boolean;
  is_friend: boolean;
  is_self: boolean;
  followers_count: number;
  following_count: number;
}

export interface FollowListUser {
  id: string;
  name: string;
  username?: string | null;
}

export interface FollowListResponse {
  users: FollowListUser[];
  next_cursor?: string;
  has_more: boolean;
}

export interface FollowListArgs {
  handle: string;
  limit?: number;
  cursor?: string;
}

function listQuery(path: (handle: string) => string) {
  return (args: FollowListArgs) => {
    const params = new URLSearchParams();
    if (args.limit) params.set('limit', String(args.limit));
    if (args.cursor) params.set('cursor', args.cursor);
    const qs = params.toString();
    return {
      url: `${path(encodeURIComponent(args.handle))}${qs ? `?${qs}` : ''}`,
      method: 'GET' as const,
    };
  };
}

export const socialApi = baseApi.injectEndpoints({
  endpoints: (build) => ({
    getFollowStatus: build.query<FollowStatus, string>({
      query: (handle) => ({
        url: `/users/${encodeURIComponent(handle)}/follow/status`,
        method: 'GET',
      }),
      providesTags: (_result, _error, handle) => [
        { type: 'Follow', id: handle },
      ],
    }),
    followUser: build.mutation<FollowStatus, string>({
      query: (handle) => ({
        url: `/users/${encodeURIComponent(handle)}/follow`,
        method: 'POST',
      }),
      invalidatesTags: (_result, _error, handle) => [
        { type: 'Follow', id: handle },
        'Follow',
        'User',
      ],
    }),
    unfollowUser: build.mutation<FollowStatus, string>({
      query: (handle) => ({
        url: `/users/${encodeURIComponent(handle)}/follow`,
        method: 'DELETE',
      }),
      invalidatesTags: (_result, _error, handle) => [
        { type: 'Follow', id: handle },
        'Follow',
        'User',
      ],
    }),
    getFollowers: build.query<FollowListResponse, FollowListArgs>({
      query: listQuery((h) => `/users/${h}/followers`),
      providesTags: (_result, _error, args) => [
        { type: 'Follow', id: `${args.handle}-followers` },
      ],
    }),
    getFollowing: build.query<FollowListResponse, FollowListArgs>({
      query: listQuery((h) => `/users/${h}/following`),
      providesTags: (_result, _error, args) => [
        { type: 'Follow', id: `${args.handle}-following` },
      ],
    }),
    getFriends: build.query<FollowListResponse, FollowListArgs>({
      query: listQuery((h) => `/users/${h}/friends`),
      providesTags: (_result, _error, args) => [
        { type: 'Follow', id: `${args.handle}-friends` },
      ],
    }),
  }),
  overrideExisting: false,
});

export const {
  useGetFollowStatusQuery,
  useFollowUserMutation,
  useUnfollowUserMutation,
  useGetFollowersQuery,
  useGetFollowingQuery,
  useGetFriendsQuery,
  useLazyGetFollowersQuery,
  useLazyGetFollowingQuery,
  useLazyGetFriendsQuery,
} = socialApi;
