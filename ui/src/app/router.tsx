import { createBrowserRouter, type RouteObject } from 'react-router';
import App from '@/App';
import HomePage from '@/pages/HomePage';
import LoginPage from '@/features/auth/LoginPage';
import SignupPage from '@/features/auth/SignupPage';
import { AuthLayout } from '@/features/auth/AuthLayout';
import { RequireAuth } from '@/features/auth/RequireAuth';
import { RequireGuest } from '@/features/auth/RequireGuest';
import { AppShell } from '@/components/AppShell';
import SpacesPage from '@/features/spaces/SpacesPage';
import FeedPage from '@/features/feed/FeedPage';
import ProfilePage from '@/features/profile/ProfilePage';

export const routes: RouteObject[] = [
  {
    path: '/',
    element: <App />,
    children: [
      {
        element: (
          <RequireAuth>
            <AppShell />
          </RequireAuth>
        ),
        children: [
          { index: true, element: <HomePage /> },
          { path: 'spaces', element: <SpacesPage /> },
          { path: 'feed', element: <FeedPage /> },
          { path: 'u/:handle', element: <ProfilePage /> },
        ],
      },
      {
        element: (
          <RequireGuest>
            <AuthLayout />
          </RequireGuest>
        ),
        children: [
          { path: 'login', element: <LoginPage /> },
          { path: 'signup', element: <SignupPage /> },
        ],
      },
    ],
  },
];

export const router = createBrowserRouter(routes);
