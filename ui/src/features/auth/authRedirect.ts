/** Returns a safe local redirect target, falling back to `/`. */
export function safeNextPath(raw: string | null): string {
  if (raw && raw.startsWith('/') && !raw.startsWith('//')) {
    return raw;
  }
  return '/';
}
