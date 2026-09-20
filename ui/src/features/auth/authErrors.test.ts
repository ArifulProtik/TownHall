import { expect, test } from 'vitest';
import { getAuthErrorMessage } from '@/features/auth/authErrors';

test('maps 401 to invalid credentials', () => {
  expect(getAuthErrorMessage({ status: 401, data: { error: 'x' } })).toBe(
    'Invalid email or password.',
  );
});

test('maps 409 to already registered', () => {
  expect(getAuthErrorMessage({ status: 409, data: { error: 'x' } })).toBe(
    'This email is already registered.',
  );
});

test('maps 429 with Retry-After seconds', () => {
  const headers = new Headers({ 'Retry-After': '34' });
  const err = { status: 429, data: { error: 'x' }, meta: { response: { headers } } };
  expect(getAuthErrorMessage(err as never)).toBe('Too many attempts. Try again in 34s.');
});

test('maps 429 without header generically', () => {
  expect(getAuthErrorMessage({ status: 429, data: { error: 'x' } })).toBe(
    'Too many attempts. Try again later.',
  );
});

test('maps validation fields to the first entry', () => {
  expect(
    getAuthErrorMessage({
      status: 400,
      data: { error: 'validation failed', fields: { email: 'email' } },
    }),
  ).toBe('email: email');
});

test('falls back for unknown shapes', () => {
  expect(getAuthErrorMessage(new Error('boom'))).toBe('Something went wrong. Please try again.');
  expect(getAuthErrorMessage(undefined)).toBe('Something went wrong. Please try again.');
});
