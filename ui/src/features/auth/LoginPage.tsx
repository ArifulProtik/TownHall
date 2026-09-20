import { Link, useNavigate, useSearchParams } from 'react-router';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useLoginMutation } from '@/features/auth/authApi';
import { getAuthErrorMessage } from '@/features/auth/authErrors';
import { LoginSchema, type LoginFormValues } from '@/features/auth/authSchema';
import { M3Button } from '@/components/M3Button';
import { M3Card } from '@/components/M3Card';
import { M3TextField } from '@/components/M3TextField';

export default function LoginPage() {
  const [formError, setFormError] = useState('');
  const [login, { isLoading }] = useLoginMutation();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({ resolver: zodResolver(LoginSchema), mode: 'onTouched' });

  async function onValid(values: LoginFormValues) {
    setFormError('');
    try {
      await login({ email: values.email.trim(), password: values.password }).unwrap();
      navigate(params.get('next') ?? '/', { replace: true });
    } catch (err) {
      setFormError(getAuthErrorMessage(err));
    }
  }

  return (
    <M3Card className="w-full max-w-sm space-y-4 p-6">
      <h1 className="text-xl font-medium text-on-surface">Log in to TownHall</h1>
      {formError ? (
        <p role="alert" className="text-sm text-error">
          {formError}
        </p>
      ) : null}
      <form onSubmit={handleSubmit(onValid)} noValidate className="space-y-4">
        <M3TextField
          id="email"
          label="Email"
          type="email"
          autoComplete="email"
          error={errors.email?.message}
          {...register('email')}
        />
        <M3TextField
          id="password"
          label="Password"
          type="password"
          autoComplete="current-password"
          error={errors.password?.message}
          {...register('password')}
        />
        <M3Button type="submit" loading={isLoading} className="w-full">
          {isLoading ? 'Logging in…' : 'Log in'}
        </M3Button>
      </form>
      <p className="text-sm text-on-surface-variant">
        No account?{' '}
        <Link to="/signup" className="text-primary underline">
          Sign up
        </Link>
      </p>
    </M3Card>
  );
}
