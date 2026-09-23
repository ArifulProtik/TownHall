import { Stack } from '@phosphor-icons/react';
import { AnimatePresence, motion, useReducedMotion } from 'motion/react';
import { Outlet, useLocation } from 'react-router';
import { AuthHero } from '@/features/auth/AuthHero';
import { ThemeToggle } from '@/components/ThemeToggle';

export function AuthLayout() {
  const reduceMotion = useReducedMotion();
  const location = useLocation();

  return (
    <div className="flex min-h-screen bg-background text-foreground">
      <AuthHero />

      <div className="relative flex min-h-0 min-w-0 flex-1 flex-col">
        <header className="flex items-center justify-between px-6 pt-5 sm:px-8">
          <div className="flex items-center gap-2.5 lg:invisible">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg border border-border bg-card">
              <Stack className="h-4 w-4 text-primary" weight="duotone" aria-hidden="true" />
            </div>
            <span className="font-heading text-base font-semibold tracking-tight">TownHall</span>
          </div>
          <ThemeToggle />
        </header>

        <main className="flex min-h-0 flex-1 overflow-y-auto px-6 py-6 sm:px-8">
          <AnimatePresence mode="wait">
            <motion.div
              key={location.pathname}
              initial={reduceMotion ? { opacity: 1 } : { opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              exit={reduceMotion ? { opacity: 1 } : { opacity: 0, y: -8 }}
              transition={{ duration: 0.28, ease: [0.16, 1, 0.3, 1] }}
              className="m-auto w-full max-w-[440px] rounded-2xl border border-border bg-card p-6 shadow-sm sm:p-8"
            >
              <Outlet />
            </motion.div>
          </AnimatePresence>
        </main>

        <footer className="px-6 pb-5 text-center text-xs text-muted-foreground/60 sm:px-8">
          © {new Date().getFullYear()} TownHall
        </footer>
      </div>
    </div>
  );
}
