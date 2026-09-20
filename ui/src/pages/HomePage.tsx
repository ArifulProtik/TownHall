import { Link, useNavigate } from 'react-router';
import { authApi, useLogoutMutation } from '@/features/auth/authApi';
import { useAppDispatch } from '@/app/hooks';
import { clearCredentials } from '@/features/auth/authSlice';

export default function HomePage() {
  const dispatch = useAppDispatch();
  const navigate = useNavigate();
  const [logout] = useLogoutMutation();

  async function onLogout() {
    try {
      await logout().unwrap();
    } finally {
      dispatch(clearCredentials());
      dispatch(authApi.util.resetApiState());
      navigate('/login', { replace: true });
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-sm space-y-4 rounded-lg bg-white p-6 text-center shadow">
        <h1 className="text-xl font-bold text-gray-900">Welcome to TownHall</h1>
        <p className="text-sm text-on-surface-variant">You are logged in.</p>
        <div className="flex justify-center gap-4 text-sm">
          <Link to="/spaces" className="text-primary underline">
            Spaces
          </Link>
          <Link to="/feed" className="text-primary underline">
            Feed
          </Link>
        </div>
        <button
          type="button"
          onClick={onLogout}
          className="w-full rounded-md bg-gray-900 px-3 py-2 text-sm font-medium text-white"
        >
          Log out
        </button>
      </div>
    </main>
  );
}
