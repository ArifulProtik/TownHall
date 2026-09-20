import { createBrowserRouter, type RouteObject } from 'react-router';
import App from '@/App';
import HomePage from '@/pages/HomePage';
import LoginPage from '@/features/auth/LoginPage';
import SignupPage from '@/features/auth/SignupPage';
import { RequireAuth } from '@/features/auth/RequireAuth';

export const routes: RouteObject[] = [
  {
    path: '/',
    element: <App />,
    children: [
      {
        index: true,
        element: (
          <RequireAuth>
            <HomePage />
          </RequireAuth>
        ),
      },
      { path: 'login', element: <LoginPage /> },
      { path: 'signup', element: <SignupPage /> },
    ],
  },
];

export const router = createBrowserRouter(routes);
