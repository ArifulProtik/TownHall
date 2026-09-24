const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1';

function apiOrigin(): string {
  if (API_BASE_URL.startsWith('http')) {
    return new URL(API_BASE_URL).origin;
  }
  // Relative base means the SPA is served from (or proxied to) the API
  // origin — e.g. Vite dev with /api and /uploads proxied to :8080.
  return window.location.origin;
}

/**
 * Resolve a media URL (avatar, banner) for rendering.
 *
 * Absolute remote URLs (UploadThing/CDN), data: and blob: URLs pass
 * through untouched. Relative paths such as the `/uploads/...` URLs
 * returned by local storage are resolved against the API origin so
 * images load in every deployment shape (proxied dev, same-origin
 * prod, or split frontend/API origins).
 */
export function resolveMediaUrl(url?: string | null): string | undefined {
  if (!url) return undefined;
  if (/^(https?:|data:|blob:)/.test(url)) return url;
  if (url.startsWith('/')) return `${apiOrigin()}${url}`;
  return url;
}
