import {
  CircleNotch,
  Eye,
  EyeSlash,
  WarningCircle,
} from '@phosphor-icons/react';
import { Link, useLocation, useNavigate } from 'react-router';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { motion, useReducedMotion } from 'motion/react';
import { useLoginMutation } from '@/features/auth/authApi';
import { getAuthErrorMessage } from '@/features/auth/authErrors';
import { safeNextPath } from '@/features/auth/authRedirect';
import { LoginSchema, type LoginFormValues } from '@/features/auth/authSchema';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Field, FieldError, FieldLabel } from '@/components/ui/field';

const stagger = (index: number, reduce: boolean | null) =>
  reduce
    ? { initial: { opacity: 0 }, animate: { opacity: 1 }, transition: { duration: 0.2 } }
    : {
        initial: { opacity: 0, y: 10 },
        animate: { opacity: 1, y: 0 },
        transition: { delay: 0.05 * index, duration: 0.3, ease: [0.25, 0.1, 0.25, 1] as const },
      };

export default function LoginPage() {
  const reduceMotion = useReducedMotion();
  const [formError, setFormError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [login, { isLoading }] = useLoginMutation();
  const navigate = useNavigate();
  const location = useLocation();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({ resolver: zodResolver(LoginSchema), mode: 'onTouched' });

  async function onValid(values: LoginFormValues) {
    setFormError('');
    try {
      await login(values).unwrap();
      const next = safeNextPath(new URLSearchParams(location.search).get('next'));
      navigate(next, { replace: true });
    } catch (err) {
      setFormError(getAuthErrorMessage(err));
    }
  }

  return (
    <div className="space-y-5">
      <motion.div className="space-y-1.5" {...stagger(0, reduceMotion)}>
        <h1 className="font-heading text-2xl font-semibold tracking-tight">Log in</h1>
        <p className="text-sm text-muted-foreground">
          Enter your details to continue to TownHall.
        </p>
      </motion.div>

      {formError ? (
        <motion.div
          role="alert"
          initial={reduceMotion ? { opacity: 0 } : { opacity: 0, y: -6 }}
          animate={{ opacity: 1, y: 0 }}
          className="flex items-center gap-2 rounded-xl border border-destructive/20 bg-destructive/5 px-3 py-2.5 text-sm text-destructive"
        >
          <WarningCircle className="h-4 w-4 shrink-0" aria-hidden="true" />
          <span>{formError}</span>
        </motion.div>
      ) : null}

      <form onSubmit={handleSubmit(onValid)} noValidate className="space-y-4">
        <motion.div {...stagger(1, reduceMotion)}>
          <Field className="gap-1.5">
            <FieldLabel htmlFor="email">Email</FieldLabel>
            <Input
              id="email"
              type="email"
              autoComplete="email"
              placeholder="you@example.com"
              aria-invalid={errors.email ? true : undefined}
              {...register('email')}
            />
            {errors.email?.message ? (
              <FieldError className="text-xs">{errors.email.message}</FieldError>
            ) : null}
          </Field>
        </motion.div>

        <motion.div {...stagger(2, reduceMotion)}>
          <Field className="gap-1.5">
            <FieldLabel htmlFor="password">Password</FieldLabel>
            <div className="relative">
              <Input
                id="password"
                type={showPassword ? 'text' : 'password'}
                autoComplete="current-password"
                placeholder="••••••••"
                className="pr-10"
                aria-invalid={errors.password ? true : undefined}
                {...register('password')}
              />
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="absolute right-1 top-1/2 -translate-y-1/2 text-muted-foreground"
                onClick={() => setShowPassword((prev) => !prev)}
                aria-label={showPassword ? 'Hide password' : 'Show password'}
              >
                {showPassword ? (
                  <EyeSlash aria-hidden="true" />
                ) : (
                  <Eye aria-hidden="true" />
                )}
              </Button>
            </div>
            {errors.password?.message ? (
              <FieldError className="text-xs">{errors.password.message}</FieldError>
            ) : null}
          </Field>
        </motion.div>

        <motion.div className="space-y-4 pt-1" {...stagger(3, reduceMotion)}>
          <Button
            type="submit"
            disabled={isLoading}
            className="h-10 w-full text-sm transition-transform duration-200 active:scale-[0.98]"
          >
            {isLoading ? (
              <>
                <CircleNotch className="animate-spin" aria-hidden="true" />
                Logging in…
              </>
            ) : (
              'Log in'
            )}
          </Button>

          <p className="text-center text-sm text-muted-foreground">
            Don&apos;t have an account?{' '}
            <Link
              to="/signup"
              className="font-semibold text-foreground underline-offset-4 transition-colors hover:underline"
            >
              Sign up
            </Link>
          </p>
        </motion.div>
      </form>
    </div>
  );
}
