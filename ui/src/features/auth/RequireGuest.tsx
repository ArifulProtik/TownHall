import { Navigate, useLocation } from 'react-router';
import type { ReactElement } from 'react';
import { useAppSelector } from '@/app/hooks';
import { selectIsAuthenticated, selectSessionChecked } from '@/features/auth/authSlice';
import { safeNextPath } from '@/features/auth/authRedirect';

export function RequireGuest({ children }: { children: ReactElement }) {
  const isAuthenticated = useAppSelector(selectIsAuthenticated);
  const sessionChecked = useAppSelector(selectSessionChecked);
  const location = useLocation();
  // Wait for the boot-time refresh: redirecting before the session is
  // restored would bounce a returning user away from login needlessly.
  if (!sessionChecked) {
    return null;
  }
  if (isAuthenticated) {
    const next = safeNextPath(new URLSearchParams(location.search).get('next'));
    return <Navigate to={next} replace />;
  }
  return children;
}
