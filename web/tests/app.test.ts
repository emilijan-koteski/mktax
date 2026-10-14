import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import App from '../src/App.svelte';

describe('App', () => {
  it('shows the brand heading', () => {
    render(App);
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('Даночко');
  });
});
