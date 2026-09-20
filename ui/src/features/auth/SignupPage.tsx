import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router';
import { useSignupMutation } from '@/features/auth/authApi';
import { getAuthErrorMessage } from '@/features/auth/authErrors';
import { TextField } from '@/components/TextField';

function validEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export default function SignupPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [formError, setFormError] = useState('');
  const [signup, { isLoading }] = useSignupMutation();
  const navigate = useNavigate();

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (name.trim().length < 2) {
      setFormError('Name must be at least 2 characters.');
      return;
    }
    if (!validEmail(email)) {
      setFormError('Enter a valid email address.');
      return;
    }
    if (password.length < 8) {
      setFormError('Password must be at least 8 characters.');
      return;
    }
    setFormError('');
    try {
      await signup({ name: name.trim(), email: email.trim(), password }).unwrap();
      navigate('/login', { replace: true });
    } catch (err) {
      setFormError(getAuthErrorMessage(err));
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <form
        onSubmit={onSubmit}
        noValidate
        className="w-full max-w-sm space-y-4 rounded-lg bg-white p-6 shadow"
      >
        <h1 className="text-xl font-bold text-gray-900">Create your account</h1>
        {formError ? (
          <p role="alert" className="text-sm text-red-600">
            {formError}
          </p>
        ) : null}
        <TextField id="name" label="Name" value={name} onChange={setName} autoComplete="name" />
        <TextField
          id="email"
          label="Email"
          type="email"
          value={email}
          onChange={setEmail}
          autoComplete="email"
        />
        <TextField
          id="password"
          label="Password"
          type="password"
          value={password}
          onChange={setPassword}
          autoComplete="new-password"
        />
        <button
          type="submit"
          disabled={isLoading}
          className="w-full rounded-md bg-gray-900 px-3 py-2 text-sm font-medium text-white disabled:opacity-50"
        >
          {isLoading ? 'Signing up…' : 'Sign up'}
        </button>
        <p className="text-sm text-gray-600">
          Have an account?{' '}
          <Link to="/login" className="underline">
            Log in
          </Link>
        </p>
      </form>
    </main>
  );
}
