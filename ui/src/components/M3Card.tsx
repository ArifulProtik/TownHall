import type { HTMLAttributes, ReactNode } from 'react';

interface M3CardProps extends Omit<HTMLAttributes<HTMLDivElement>, 'className'> {
  variant?: 'elevated' | 'filled';
  className?: string;
  children: ReactNode;
}

export function M3Card({ variant = 'elevated', className = '', children, ...rest }: M3CardProps) {
  const style =
    variant === 'elevated'
      ? 'bg-surface-container-low shadow-m3-1'
      : 'bg-surface-container-highest';
  return (
    <div className={`rounded-m3-md ${style} ${className}`} {...rest}>
      {children}
    </div>
  );
}
