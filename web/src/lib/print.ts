import { tick } from 'svelte';
import { app } from './state.svelte';

export async function printReport(title: string): Promise<void> {
  const previous = document.title;
  document.title = title;
  app.printMode = true;
  await tick();
  window.print();
  app.printMode = false;
  document.title = previous;
}
