import { Outlet, useLocation } from 'react-router';
import { AnimatePresence, motion, useReducedMotion } from 'motion/react';
import { ThemeToggle } from '@/components/ThemeToggle';

const blobTransition = { duration: 18, repeat: Infinity, ease: 'easeInOut' as const };

export function AuthLayout() {
  const reduceMotion = useReducedMotion();
  const location = useLocation();
  return (
    <main className="relative flex min-h-screen items-center justify-center overflow-hidden bg-surface px-4">
      <div aria-hidden="true" className="pointer-events-none absolute inset-0">
        <motion.div
          className="absolute -left-24 top-1/4 h-96 w-96 rounded-full bg-primary-container/50 blur-3xl"
          animate={reduceMotion ? undefined : { x: [0, 40, -20, 0], y: [0, -30, 20, 0] }}
          transition={blobTransition}
        />
        <motion.div
          className="absolute -right-24 bottom-1/4 h-96 w-96 rounded-full bg-tertiary-container/50 blur-3xl"
          animate={reduceMotion ? undefined : { x: [0, -30, 25, 0], y: [0, 25, -20, 0] }}
          transition={{ ...blobTransition, duration: 22 }}
        />
      </div>
      <div className="absolute right-4 top-4">
        <ThemeToggle />
      </div>
      <AnimatePresence mode="wait">
        <motion.div
          key={location.pathname}
          initial={reduceMotion ? { opacity: 0 } : { opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -12 }}
          transition={{ duration: 0.25 }}
          className="relative w-full max-w-sm"
        >
          <Outlet />
        </motion.div>
      </AnimatePresence>
    </main>
  );
}
