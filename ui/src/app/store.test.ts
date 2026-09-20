import { expect, test } from 'vitest';
import { clearCredentials, selectIsAuthenticated, setCredentials } from '@/features/auth/authSlice';
import { createAppStore } from '@/app/store';

test('auth slice sets and clears the token', () => {
  const store = createAppStore();
  expect(selectIsAuthenticated(store.getState())).toBe(false);
  store.dispatch(setCredentials({ accessToken: 'abc' }));
  expect(selectIsAuthenticated(store.getState())).toBe(true);
  store.dispatch(clearCredentials());
  expect(selectIsAuthenticated(store.getState())).toBe(false);
});
