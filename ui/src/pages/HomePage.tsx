import { Link, useNavigate } from 'react-router';
import { authApi, useLogoutMutation } from '@/features/auth/authApi';
import { useAppDispatch } from '@/app/hooks';
import { clearCredentials } from '@/features/auth/authSlice';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';

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
    <main className="flex min-h-screen items-center justify-center bg-muted px-4">
      <Card className="w-full max-w-sm">
        <CardContent className="space-y-4 text-center">
          <h1 className="text-xl font-bold text-card-foreground">Welcome to TownHall</h1>
          <p className="text-sm text-muted-foreground">You are logged in.</p>
          <div className="flex justify-center gap-4 text-sm">
            <Link to="/spaces" className="text-primary underline">
              Spaces
            </Link>
            <Link to="/feed" className="text-primary underline">
              Feed
            </Link>
          </div>
          <Button type="button" onClick={onLogout} className="w-full">
            Log out
          </Button>
        </CardContent>
      </Card>
    </main>
  );
}
