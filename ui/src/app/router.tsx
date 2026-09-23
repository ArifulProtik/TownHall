import { lazy, Suspense, type ComponentType } from 'react';
import { createBrowserRouter, type RouteObject } from 'react-router';
import App from '@/App';
import HomePage from '@/pages/HomePage';
import { AuthLayout } from '@/features/auth/AuthLayout';
import { RequireAuth } from '@/features/auth/RequireAuth';
import { RequireGuest } from '@/features/auth/RequireGuest';
import { AppShell } from '@/components/AppShell';
import { PageLoadingSkeleton } from '@/components/ui/PageLoadingSkeleton';

const LoginPage = lazy(() => import('@/features/auth/LoginPage'));
const SignupPage = lazy(() => import('@/features/auth/SignupPage'));
const OnboardingPage = lazy(() => import('@/features/auth/OnboardingPage'));
const SpacesPage = lazy(() => import('@/features/spaces/SpacesPage'));
const FeedPage = lazy(() => import('@/features/feed/FeedPage'));
const ProfilePage = lazy(() => import('@/features/profile/ProfilePage'));
const MessagesPage = lazy(() => import('@/features/messages/MessagesPage'));

function withSuspense(Component: ComponentType) {
  return (
    <Suspense fallback={<PageLoadingSkeleton />}>
      <Component />
    </Suspense>
  );
}

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
          { path: 'dms', element: withSuspense(MessagesPage) },
          { path: 'spaces', element: withSuspense(SpacesPage) },
          { path: 'feed', element: withSuspense(FeedPage) },
          { path: 'u/:handle', element: withSuspense(ProfilePage) },
        ],
      },
      {
        element: (
          <RequireAuth>
            <AuthLayout />
          </RequireAuth>
        ),
        children: [{ path: 'onboarding', element: withSuspense(OnboardingPage) }],
      },
      {
        element: (
          <RequireGuest>
            <AuthLayout />
          </RequireGuest>
        ),
        children: [
          { path: 'login', element: withSuspense(LoginPage) },
          { path: 'signup', element: withSuspense(SignupPage) },
        ],
      },
    ],
  },
];

export const router = createBrowserRouter(routes);
