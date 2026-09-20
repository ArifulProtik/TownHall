import type { ButtonHTMLAttributes } from 'react';

type M3ButtonVariant = 'filled' | 'tonal' | 'outlined' | 'text';

const variantStyles: Record<M3ButtonVariant, string> = {
  filled: 'bg-primary text-on-primary hover:shadow-m3-1',
  tonal: 'bg-secondary-container text-on-secondary-container',
  outlined: 'border border-outline text-primary',
  text: 'text-primary',
};

interface M3ButtonProps extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'className'> {
  variant?: M3ButtonVariant;
  loading?: boolean;
  className?: string;
}

export function M3Button({
  variant = 'filled',
  loading = false,
  className = '',
  disabled,
  children,
  ...rest
}: M3ButtonProps) {
  return (
    <button
      type="button"
      disabled={disabled ?? loading}
      className={`inline-flex h-10 items-center justify-center gap-2 rounded-m3-md px-6 text-sm font-medium transition-shadow disabled:opacity-50 ${variantStyles[variant]} ${className}`}
      {...rest}
    >
      {loading ? (
        <span
          aria-hidden="true"
          className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
        />
      ) : null}
      {children}
    </button>
  );
}
