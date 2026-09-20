import { expect, test } from 'vitest';
import { LoginSchema, SignupSchema } from '@/features/auth/authSchema';

test('login schema accepts valid input and rejects boundaries', () => {
  expect(LoginSchema.safeParse({ email: 'joe@example.com', password: 'password123' }).success).toBe(
    true,
  );
  expect(LoginSchema.safeParse({ email: 'not-an-email', password: 'password123' }).success).toBe(
    false,
  );
  expect(LoginSchema.safeParse({ email: 'joe@example.com', password: 'short' }).success).toBe(
    false,
  );
});

test('signup schema enforces the name contract', () => {
  const base = { email: 'joe@example.com', password: 'password123' };
  expect(SignupSchema.safeParse({ ...base, name: 'J' }).success).toBe(false);
  expect(SignupSchema.safeParse({ ...base, name: '  ' }).success).toBe(false);
  const ok = SignupSchema.safeParse({ ...base, name: '  Joe  ' });
  expect(ok.success).toBe(true);
  if (ok.success) expect(ok.data.name).toBe('Joe');
});
