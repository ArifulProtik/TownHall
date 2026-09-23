import { useEffect } from 'react';
import { Outlet } from 'react-router';
import { useAppDispatch } from '@/app/hooks';
import { useRefreshQuery } from '@/features/auth/authApi';
import { markSessionChecked } from '@/features/auth/authSlice';

export default function App() {
  const dispatch = useAppDispatch();
  // Restores the session from the HttpOnly refresh cookie on boot. A query
  // (not a mutation) dedupes the in-flight request, so StrictMode's
  // double-mount can't fire two rotations and trip reuse detection.
  const { isSuccess, isError } = useRefreshQuery();

  useEffect(() => {
    if (isSuccess || isError) {
      dispatch(markSessionChecked());
    }
  }, [isSuccess, isError, dispatch]);

  return <Outlet />;
}
