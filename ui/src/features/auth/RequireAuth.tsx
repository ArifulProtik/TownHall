import { Navigate, useLocation } from 'react-router';
import type { ReactElement } from 'react';
import { useAppSelector } from '@/app/hooks';
import { selectIsAuthenticated } from '@/features/auth/authSlice';

export function RequireAuth({ children }: { children: ReactElement }) {
  const isAuthenticated = useAppSelector(selectIsAuthenticated);
  const location = useLocation();
  if (!isAuthenticated) {
    return <Navigate to={`/login?next=${encodeURIComponent(location.pathname)}`} replace />;
  }
  return children;
}
