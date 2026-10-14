import { render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import App from '../src/App.svelte';
import { installFakeEngine, ratesResponse } from './fakeEngine';

describe('App flow', () => {
  beforeEach(() => {
    installFakeEngine();
    vi.stubGlobal(
      'fetch',
      vi.fn(async (url: string) => new Response(JSON.stringify(ratesResponse(url.match(/rates\/(\d{4}-\d{2})/)![1], ['EUR', 'USD'])))),
    );
    vi.stubGlobal('navigator', { ...navigator, clipboard: { writeText: vi.fn(async () => {}) } });
  });
  afterEach(() => vi.unstubAllGlobals());

  it('goes from upload to report', async () => {
    render(App);
    await waitFor(() => expect(screen.getByRole('status')).toHaveTextContent(/Подготвено/));
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Даночко');
    const input = screen.getByLabelText(/изберете датотеки/i);
    await userEvent.upload(input, new File(['Action\n'], '2025.csv', { type: 'text/csv' }));
    await waitFor(() => expect(screen.getByText(/10 трговски трансакции/)).toBeInTheDocument());
    await userEvent.click(screen.getByRole('button', { name: 'Пресметај' }));
    await waitFor(() => expect(screen.getByRole('status')).toHaveTextContent(/Извештајот е подготвен/));
    expect(screen.getByRole('region', { name: /јануари 2026/ })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Зачувај CSV/ })).toBeInTheDocument();
    expect(screen.getByText(/Не е даночен совет/)).toBeInTheDocument();
  });
});
