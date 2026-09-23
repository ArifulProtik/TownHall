import {
  motion,
  useMotionValue,
  useReducedMotion,
  useSpring,
  type MotionValue,
} from 'motion/react';
import { Stack } from '@phosphor-icons/react';
import { useCallback, useRef, type PointerEvent as ReactPointerEvent } from 'react';

const MESH = [
  {
    key: 'core',
    className: 'bg-primary/25',
    size: 'h-[36rem] w-[36rem]',
    from: { x: '-22%', y: '10%' },
    to: { x: '-4%', y: '28%' },
    duration: 22,
  },
  {
    key: 'halo',
    className: 'bg-primary/10',
    size: 'h-[28rem] w-[28rem]',
    from: { x: '18%', y: '40%' },
    to: { x: '32%', y: '22%' },
    duration: 26,
  },
  {
    key: 'depth',
    className: 'bg-secondary/15',
    size: 'h-[22rem] w-[22rem]',
    from: { x: '-8%', y: '-6%' },
    to: { x: '10%', y: '4%' },
    duration: 20,
  },
];

function MeshOrb({
  orb,
  reduceMotion,
  px,
  py,
}: {
  orb: (typeof MESH)[number];
  reduceMotion: boolean | null;
  px: MotionValue<number>;
  py: MotionValue<number>;
}) {
  return (
    <motion.div className="absolute inset-0" style={{ x: px, y: py }}>
      <motion.div
        className={`absolute rounded-full blur-3xl ${orb.size} ${orb.className}`}
        style={{ top: '20%', left: '18%' }}
        initial={orb.from}
        animate={
          reduceMotion
            ? orb.from
            : {
                x: [orb.from.x, orb.to.x, orb.from.x],
                y: [orb.from.y, orb.to.y, orb.from.y],
              }
        }
        transition={
          reduceMotion
            ? undefined
            : { duration: orb.duration, repeat: Infinity, ease: 'easeInOut' }
        }
      />
    </motion.div>
  );
}

function LiveIndicator({ reduceMotion }: { reduceMotion: boolean | null }) {
  return (
    <div className="flex items-center gap-2">
      <span className="relative flex h-2 w-2">
        {reduceMotion ? null : (
          <>
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-500 opacity-75" />
            <span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-500" />
          </>
        )}
        {reduceMotion ? (
          <span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-500" />
        ) : null}
      </span>
      <span className="text-xs font-medium text-muted-foreground">Real-time</span>
    </div>
  );
}

export function AuthHero() {
  const reduceMotion = useReducedMotion();
  const ref = useRef<HTMLElement>(null);
  const px = useSpring(useMotionValue(0), { stiffness: 40, damping: 20 });
  const py = useSpring(useMotionValue(0), { stiffness: 40, damping: 20 });

  const onPointerMove = useCallback(
    (event: ReactPointerEvent<HTMLElement>) => {
      if (reduceMotion || event.pointerType === 'touch') return;
      const rect = ref.current?.getBoundingClientRect();
      if (!rect) return;
      px.set(((event.clientX - rect.left) / rect.width - 0.5) * 12);
      py.set(((event.clientY - rect.top) / rect.height - 0.5) * 12);
    },
    [px, py, reduceMotion],
  );

  const onPointerLeave = useCallback(() => {
    px.set(0);
    py.set(0);
  }, [px, py]);

  return (
    <aside
      ref={ref}
      onPointerMove={onPointerMove}
      onPointerLeave={onPointerLeave}
      className="relative isolate hidden w-[45%] max-w-[600px] shrink-0 overflow-hidden bg-background lg:flex"
      aria-hidden="true"
    >
      <div className="absolute inset-0 bg-gradient-to-br from-background via-background to-muted" />

      {MESH.map((orb) => (
        <MeshOrb key={orb.key} orb={orb} reduceMotion={reduceMotion} px={px} py={py} />
      ))}

      <div className="absolute inset-0 bg-background/40" />

      <div className="relative z-10 flex h-full flex-col justify-between p-14">
        <motion.div
          className="flex h-12 w-12 items-center justify-center rounded-xl border border-border bg-card/60"
          initial={reduceMotion ? { opacity: 1 } : { opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
        >
          <Stack className="h-6 w-6 text-primary" weight="duotone" aria-hidden="true" />
        </motion.div>

        <div className="space-y-6">
          <motion.h1
            className="font-heading text-6xl font-semibold tracking-[-0.03em] text-foreground"
            initial={reduceMotion ? { opacity: 1 } : { opacity: 0, y: 24 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.05, ease: [0.16, 1, 0.3, 1] }}
          >
            TownHall
          </motion.h1>

          <motion.p
            className="max-w-[32ch] text-xl leading-relaxed text-muted-foreground"
            initial={reduceMotion ? { opacity: 1 } : { opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.1, ease: [0.16, 1, 0.3, 1] }}
          >
            Where communities come alive.
          </motion.p>

          <motion.div
            initial={reduceMotion ? { opacity: 1 } : { opacity: 0, y: 12 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.15, ease: [0.16, 1, 0.3, 1] }}
          >
            <LiveIndicator reduceMotion={reduceMotion} />
          </motion.div>
        </div>

        <motion.p
          className="text-xs text-muted-foreground/60"
          initial={reduceMotion ? { opacity: 1 } : { opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.6, delay: 0.25 }}
        >
          Gaming · Communities · Conversations
        </motion.p>
      </div>
    </aside>
  );
}
