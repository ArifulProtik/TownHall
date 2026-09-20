import { Link, useNavigate } from 'react-router';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useSignupMutation } from '@/features/auth/authApi';
import { getAuthErrorMessage } from '@/features/auth/authErrors';
import { SignupSchema, type SignupFormValues } from '@/features/auth/authSchema';
import { M3Button } from '@/components/M3Button';
import { M3Card } from '@/components/M3Card';
import { M3TextField } from '@/components/M3TextField';

export default function SignupPage() {
  const [formError, setFormError] = useState('');
  const [signup, { isLoading }] = useSignupMutation();
  const navigate = useNavigate();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<SignupFormValues>({ resolver: zodResolver(SignupSchema), mode: 'onTouched' });

  async function onValid(values: SignupFormValues) {
    setFormError('');
    try {
      await signup(values).unwrap();
      navigate('/login', { replace: true });
    } catch (err) {
      setFormError(getAuthErrorMessage(err));
    }
  }

  return (
    <M3Card className="w-full max-w-sm space-y-4 p-6">
      <h1 className="text-xl font-medium text-on-surface">Create your account</h1>
      {formError ? (
        <p role="alert" className="text-sm text-error">
          {formError}
        </p>
      ) : null}
      <form onSubmit={handleSubmit(onValid)} noValidate className="space-y-4">
        <M3TextField
          id="name"
          label="Name"
          autoComplete="name"
          error={errors.name?.message}
          {...register('name')}
        />
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
          autoComplete="new-password"
          error={errors.password?.message}
          {...register('password')}
        />
        <M3Button type="submit" loading={isLoading} className="w-full">
          {isLoading ? 'Signing up…' : 'Sign up'}
        </M3Button>
      </form>
      <p className="text-sm text-on-surface-variant">
        Have an account?{' '}
        <Link to="/login" className="text-primary underline">
          Log in
        </Link>
      </p>
    </M3Card>
  );
}
