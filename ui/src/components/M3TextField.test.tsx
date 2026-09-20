import { render, screen } from '@testing-library/react';
import { expect, test } from 'vitest';
import { M3TextField } from '@/components/M3TextField';

test('shows label, error text, and aria wiring', () => {
  render(<M3TextField id="email" label="Email" error="Required" />);
  expect(screen.getByLabelText('Email')).toHaveAttribute('aria-invalid', 'true');
  expect(screen.getByText('Required')).toBeInTheDocument();
});

test('shows supporting text when no error', () => {
  render(<M3TextField id="nick" label="Nick" supportingText="Visible to everyone" />);
  expect(screen.getByText('Visible to everyone')).toBeInTheDocument();
});
