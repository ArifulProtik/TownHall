import { Navigate, useLocation } from 'react-router';
import type { ReactElement } from 'react';
import { useAppSelector } from '@/app/hooks';
import { selectIsAuthenticated, selectSessionChecked } from '@/features/auth/authSlice';
import { useGetMeQuery } from '@/features/auth/authApi';

export function RequireAuth({ children }: { children: ReactElement }) {
  const isAuthenticated = useAppSelector(selectIsAuthenticated);
  const sessionChecked = useAppSelector(selectSessionChecked);
  const location = useLocation();
  const { data: user, isLoading } = useGetMeQuery(undefined, { skip: !isAuthenticated });

  // Wait for the boot-time refresh before deciding: otherwise a reload would
  // flash-redirect to login while the session is still being restored.
  if (!sessionChecked) {
    return null;
  }
  if (!isAuthenticated) {
    return <Navigate to={`/login?next=${encodeURIComponent(location.pathname)}`} replace />;
  }
  if (isLoading || !user) {
    return null;
  }

  const hasUsername = Boolean(user.username && user.username.trim());
  const isOnboardingRoute = location.pathname === '/onboarding';

  if (!hasUsername && !isOnboardingRoute) {
    return <Navigate to="/onboarding" replace />;
  }

  if (hasUsername && isOnboardingRoute) {
    return <Navigate to="/" replace />;
  }

  return children;
}

