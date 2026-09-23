import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router';
import { motion, useReducedMotion } from 'motion/react';
import {
  At,
  CheckCircle,
  CircleNotch,
  SignOut,
  WarningCircle,
} from '@phosphor-icons/react';
import {
  useCheckUsernameQuery,
  useSetupUsernameMutation,
  useLogoutMutation,
  authApi,
} from '@/features/auth/authApi';
import { useAppDispatch } from '@/app/hooks';
import { clearCredentials } from '@/features/auth/authSlice';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Field, FieldLabel } from '@/components/ui/field';
import { getAuthErrorMessage } from '@/features/auth/authErrors';

const USERNAME_REGEX = /^[a-z0-9_]{3,30}$/;

const stagger = (index: number, reduce: boolean | null) =>
  reduce
    ? { initial: { opacity: 1 }, animate: { opacity: 1 }, transition: { duration: 0.2 } }
    : {
        initial: { opacity: 0, y: 10 },
        animate: { opacity: 1, y: 0 },
        transition: { delay: 0.05 * index, duration: 0.3, ease: [0.25, 0.1, 0.25, 1] as const },
      };

export default function OnboardingPage() {
  const reduceMotion = useReducedMotion();
  const dispatch = useAppDispatch();
  const navigate = useNavigate();

  const [username, setUsername] = useState('');
  const [debouncedUsername, setDebouncedUsername] = useState('');
  const [submitError, setSubmitError] = useState('');

  const [setupUsername, { isLoading: isSubmitting }] = useSetupUsernameMutation();
  const [logout, { isLoading: isLoggingOut }] = useLogoutMutation();

  // Debounce username input
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedUsername(username.trim());
    }, 300);
    return () => clearTimeout(timer);
  }, [username]);

  const normalized = username.trim().toLowerCase();
  const isValidFormat = USERNAME_REGEX.test(normalized);
  const isTooShort = normalized.length > 0 && normalized.length < 3;
  const hasInvalidChars =
    normalized.length >= 3 && !/^[a-z0-9_]+$/.test(normalized);

  const shouldCheck = USERNAME_REGEX.test(debouncedUsername);
  const {
    data: checkData,
    isFetching: isChecking,
  } = useCheckUsernameQuery(debouncedUsername, {
    skip: !shouldCheck,
  });

  const isAvailable = checkData?.available;
  const isTaken = checkData !== undefined && !checkData.available;

  const canSubmit =
    isValidFormat &&
    !isChecking &&
    isAvailable === true &&
    !isSubmitting;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit) return;
    setSubmitError('');

    try {
      await setupUsername({ username: normalized }).unwrap();
      navigate('/', { replace: true });
    } catch (err) {
      setSubmitError(getAuthErrorMessage(err));
    }
  }

  async function handleLogout() {
    try {
      await logout().unwrap();
    } finally {
      dispatch(clearCredentials());
      dispatch(authApi.util.resetApiState());
      navigate('/login', { replace: true });
    }
  }

  return (
    <div className="space-y-6">
      <motion.div className="space-y-1.5" {...stagger(0, reduceMotion)}>
        <h1 className="font-heading text-2xl font-semibold tracking-tight">
          Choose your username
        </h1>
        <p className="text-sm text-muted-foreground">
          Pick a unique handle to identify yourself across TownHall.
        </p>
      </motion.div>

      {submitError ? (
        <motion.div
          role="alert"
          initial={reduceMotion ? { opacity: 1 } : { opacity: 0, y: -6 }}
          animate={{ opacity: 1, y: 0 }}
          className="flex items-center gap-2 rounded-xl border border-destructive/20 bg-destructive/5 px-3 py-2.5 text-sm text-destructive"
        >
          <WarningCircle className="h-4 w-4 shrink-0" aria-hidden="true" />
          <span>{submitError}</span>
        </motion.div>
      ) : null}

      <form onSubmit={handleSubmit} noValidate className="space-y-5">
        <motion.div {...stagger(1, reduceMotion)}>
          <Field className="gap-1.5">
            <FieldLabel htmlFor="username">Username handle</FieldLabel>
            <div className="relative">
              <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">
                <At className="size-4" />
              </span>
              <Input
                id="username"
                autoComplete="off"
                placeholder="username"
                value={username}
                onChange={(e) => {
                  setUsername(e.target.value.toLowerCase().replace(/\s+/g, ''));
                  setSubmitError('');
                }}
                className="pl-9 pr-9"
                maxLength={30}
              />
              <div className="absolute right-3 top-1/2 -translate-y-1/2">
                {isChecking ? (
                  <CircleNotch className="size-4 animate-spin text-muted-foreground" />
                ) : isAvailable && isValidFormat ? (
                  <CheckCircle className="size-4 text-emerald-500" />
                ) : isTaken ? (
                  <WarningCircle className="size-4 text-destructive" />
                ) : null}
              </div>
            </div>

            <div className="min-h-5 pt-0.5 text-xs">
              {isChecking ? (
                <span className="text-muted-foreground">Checking availability…</span>
              ) : isTaken ? (
                <span className="text-destructive font-medium">
                  Username is already taken
                </span>
              ) : isAvailable && isValidFormat ? (
                <span className="text-emerald-500 font-medium">
                  Username is available
                </span>
              ) : isTooShort ? (
                <span className="text-muted-foreground">
                  Must be at least 3 characters
                </span>
              ) : hasInvalidChars ? (
                <span className="text-destructive">
                  Only lowercase letters, numbers, and underscores allowed
                </span>
              ) : (
                <span className="text-muted-foreground">
                  3–30 characters, letters, numbers, and underscores
                </span>
              )}
            </div>
          </Field>
        </motion.div>

        <motion.div className="space-y-4 pt-1" {...stagger(2, reduceMotion)}>
          <Button
            type="submit"
            disabled={!canSubmit}
            className="h-10 w-full text-sm transition-transform duration-200 active:scale-[0.98]"
          >
            {isSubmitting ? (
              <>
                <CircleNotch className="animate-spin" aria-hidden="true" />
                Setting up…
              </>
            ) : (
              'Continue to TownHall'
            )}
          </Button>

          <div className="flex justify-center pt-2">
            <button
              type="button"
              onClick={handleLogout}
              disabled={isLoggingOut}
              className="flex items-center gap-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground hover:underline cursor-pointer"
            >
              <SignOut className="size-3.5" />
              <span>Log out</span>
            </button>
          </div>
        </motion.div>
      </form>
    </div>
  );
}
