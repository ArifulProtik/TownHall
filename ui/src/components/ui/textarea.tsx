import * as React from 'react';
import { cn } from 'cn';

function Textarea({
  className,
  ...props
}: React.ComponentProps<'textarea'>) {
  return (
    <textarea
      data-slot="textarea"
      className={cn(
        'flex min-h-[80px] w-full rounded-2xl border border-transparent bg-input/50 px-3 py-2 text-base transition-[color,background-color,border-color] outline-none placeholder:text-muted-foreground focus-visible:border-ring disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-destructive md:text-sm dark:aria-invalid:border-destructive/50 resize-y',
        className,
      )}
      {...props}
    />
  );
}

export { Textarea };
