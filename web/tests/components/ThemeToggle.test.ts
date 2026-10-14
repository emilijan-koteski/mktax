import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it } from 'vitest';
import ThemeToggle from '../../src/components/ThemeToggle.svelte';

describe('ThemeToggle', () => {
  afterEach(() => {
    delete document.documentElement.dataset.theme;
    localStorage.clear();
  });

  it('switches between light and dark and remembers the choice', async () => {
    document.documentElement.dataset.theme = 'light';
    render(ThemeToggle);
    await userEvent.click(screen.getByRole('button', { name: 'Вклучи темна тема' }));
    expect(document.documentElement.dataset.theme).toBe('dark');
    expect(localStorage.getItem('theme')).toBe('dark');
    await userEvent.click(screen.getByRole('button', { name: 'Вклучи светла тема' }));
    expect(document.documentElement.dataset.theme).toBe('light');
    expect(localStorage.getItem('theme')).toBe('light');
  });
});
