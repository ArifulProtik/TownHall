const sizes = {
  sm: 'h-8 w-8 text-xs',
  md: 'h-10 w-10 text-sm',
  lg: 'h-16 w-16 text-xl',
} as const;

export function Avatar({
  initials,
  size = 'md',
  className = '',
}: {
  initials: string;
  size?: keyof typeof sizes;
  className?: string;
}) {
  return (
    <span
      aria-hidden="true"
      className={`inline-flex items-center justify-center rounded-full bg-primary-container font-medium text-on-primary-container ${sizes[size]} ${className}`}
    >
      {initials}
    </span>
  );
}
