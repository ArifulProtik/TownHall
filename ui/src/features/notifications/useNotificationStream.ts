import { useEffect, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { baseApi } from '@/lib/baseApi';
import { selectAccessToken } from '@/features/auth/authSlice';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';

interface StreamHandlers {
  onEvent?: (id: string) => void;
}

// Header-auth SSE via fetch (EventSource can't send Authorization).
// Replays missed events with Last-Event-ID, reconnects with backoff.
export function useNotificationStream(handlers?: StreamHandlers) {
  const token = useSelector(selectAccessToken);
  const dispatch = useDispatch();
  const lastIdRef = useRef<string>('');
  const handlersRef = useRef(handlers);
  handlersRef.current = handlers;

  useEffect(() => {
    if (!token) return;
    let cancelled = false;
    let attempts = 0;
    let aborter: AbortController | null = null;

    const connect = async () => {
      aborter = new AbortController();
      try {
        const res = await fetch(`${API_BASE_URL}/stream`, {
          headers: {
            Authorization: `Bearer ${token}`,
            Accept: 'text/event-stream',
            ...(lastIdRef.current ? { 'Last-Event-ID': lastIdRef.current } : {}),
          },
          credentials: 'include',
          signal: aborter.signal,
        });
        if (!res.ok || !res.body) throw new Error(`stream ${res.status}`);
        attempts = 0;
        const reader = res.body.getReader();
        const decoder = new TextDecoder();
        let buf = '';
        let eventId = '';
        const flush = (chunk: string) => {
          buf += chunk;
          let idx: number;
          while ((idx = buf.indexOf('\n\n')) >= 0) {
            const raw = buf.slice(0, idx);
            buf = buf.slice(idx + 2);
            let data = '';
            for (const line of raw.split('\n')) {
              if (line.startsWith('id:')) eventId = line.slice(3).trim();
              else if (line.startsWith('data:')) data += line.slice(5).trim();
            }
            if (data.startsWith('{')) {
              if (eventId) lastIdRef.current = eventId;
              dispatch(baseApi.util.invalidateTags(['Notification']));
              handlersRef.current?.onEvent?.(eventId);
            }
          }
        };
        for (;;) {
          const { done, value } = await reader.read();
          if (done || cancelled) break;
          flush(decoder.decode(value, { stream: true }));
        }
      } catch {
        if (cancelled) return;
      }
      // Reached on clean completion (server/proxy closed the stream) as
      // well as read errors: both schedule the backoff retry. Teardown
      // returns before this via `cancelled`.
      if (cancelled) return;
      attempts += 1;
      const backoff = Math.min(1000 * 2 ** attempts, 15000);
      await new Promise((r) => setTimeout(r, backoff));
      if (!cancelled) void connect();
    };

    void connect();
    return () => {
      cancelled = true;
      aborter?.abort();
    };
  }, [token, dispatch]);
}
