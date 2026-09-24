import { baseApi } from '@/lib/baseApi';

export interface NotificationActor {
  id: string;
  name: string;
  username?: string | null;
  avatar_url?: string;
}

export interface NotificationItem {
  id: string;
  type: string;
  actor: NotificationActor;
  entity_type: string;
  entity_id: string;
  data?: string;
  read_at?: string | null;
  created_at: string;
}

export interface NotificationListResponse {
  notifications: NotificationItem[];
  next_cursor?: string;
  has_more: boolean;
}

export interface UnreadCountResponse {
  count: number;
}

export const notificationsApi = baseApi.injectEndpoints({
  endpoints: (build) => ({
    getNotifications: build.query<NotificationListResponse, { limit?: number; cursor?: string } | void>({
      query: (args) => {
        const params = new URLSearchParams();
        if (args && args.limit) params.set('limit', String(args.limit));
        if (args && args.cursor) params.set('cursor', args.cursor);
        const qs = params.toString();
        return { url: `/notifications${qs ? `?${qs}` : ''}`, method: 'GET' as const };
      },
      providesTags: ['Notification'],
    }),
    getUnreadCount: build.query<UnreadCountResponse, void>({
      query: () => ({ url: '/notifications/unread-count', method: 'GET' as const }),
      providesTags: ['Notification'],
    }),
    markNotificationRead: build.mutation<{ status: string }, string>({
      query: (id) => ({ url: `/notifications/${encodeURIComponent(id)}/read`, method: 'POST' as const }),
      invalidatesTags: ['Notification'],
    }),
    markAllNotificationsRead: build.mutation<{ status: string }, void>({
      query: () => ({ url: '/notifications/read-all', method: 'POST' as const }),
      invalidatesTags: ['Notification'],
    }),
  }),
  overrideExisting: false,
});

export const {
  useGetNotificationsQuery,
  useGetUnreadCountQuery,
  useMarkNotificationReadMutation,
  useMarkAllNotificationsReadMutation,
} = notificationsApi;
