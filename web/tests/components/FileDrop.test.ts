import { render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import FileDrop from '../../src/components/FileDrop.svelte';

describe('FileDrop', () => {
  it('passes picked files to the callback through a labelled input', async () => {
    const onfiles = vi.fn();
    render(FileDrop, { onfiles });
    const input = screen.getByLabelText(/изберете датотеки/i) as HTMLInputElement;
    expect(input.multiple).toBe(true);
    expect(input.accept).toContain('.csv');
    const file = new File(['Action\n'], '2025.csv', { type: 'text/csv' });
    await userEvent.upload(input, file);
    expect(onfiles).toHaveBeenCalledTimes(1);
    expect(onfiles.mock.calls[0][0][0].name).toBe('2025.csv');
  });

  it('accepts dropped files', () => {
    const onfiles = vi.fn();
    render(FileDrop, { onfiles });
    const zone = screen.getByTestId('dropzone');
    const file = new File(['Action\n'], 'a.csv', { type: 'text/csv' });
    const event = new Event('drop', { bubbles: true, cancelable: true }) as DragEvent;
    Object.defineProperty(event, 'dataTransfer', { value: { files: [file] } });
    zone.dispatchEvent(event);
    expect(onfiles).toHaveBeenCalledWith([file]);
  });
});
