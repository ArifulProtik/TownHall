import type { ReactElement, ReactNode } from 'react';
import { render, type RenderOptions } from '@testing-library/react';
import { MemoryRouter } from 'react-router';
import { Provider } from 'react-redux';
import { createAppStore } from '@/app/store';

interface RenderWithProvidersOptions extends Omit<RenderOptions, 'wrapper'> {
  route?: string;
}

export function renderWithProviders(
  ui: ReactElement,
  { route = '/', ...options }: RenderWithProvidersOptions = {},
) {
  const store = createAppStore();
  const Wrapper = ({ children }: { children: ReactNode }) => (
    <Provider store={store}>
      <MemoryRouter initialEntries={[route]}>{children}</MemoryRouter>
    </Provider>
  );
  return { store, ...render(ui, { wrapper: Wrapper, ...options }) };
}
