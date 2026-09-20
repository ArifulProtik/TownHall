import type { InputHTMLAttributes, Ref } from 'react';

interface M3TextFieldProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'className'> {
  label: string;
  error?: string;
  supportingText?: string;
  ref?: Ref<HTMLInputElement>;
}

export function M3TextField({ id, label, error, supportingText, ref, ...rest }: M3TextFieldProps) {
  const describedBy = error ? `${id}-error` : supportingText ? `${id}-support` : undefined;
  return (
    <div>
      <div className="rounded-t-m3-xs bg-surface-container-highest px-4 pb-1 pt-2 focus-within:border-b-2 focus-within:border-primary">
        <label htmlFor={id} className="block text-xs text-on-surface-variant">
          {label}
        </label>
        <input
          id={id}
          ref={ref}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy}
          className="block w-full bg-transparent py-1 text-sm text-on-surface focus:outline-none"
          {...rest}
        />
      </div>
      {error ? (
        <p id={`${id}-error`} className="mt-1 px-4 text-xs text-error">
          {error}
        </p>
      ) : supportingText ? (
        <p id={`${id}-support`} className="mt-1 px-4 text-xs text-on-surface-variant">
          {supportingText}
        </p>
      ) : null}
    </div>
  );
}
