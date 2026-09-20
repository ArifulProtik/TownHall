import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

// RTL's auto-cleanup registers on a global afterEach, which doesn't exist
// when vitest runs with globals disabled. Clean up explicitly so renders
// from one test never leak into the next.
afterEach(() => {
  cleanup();
});
