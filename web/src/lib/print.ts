import { tick } from 'svelte';
import { app, type PrintLayout } from './state.svelte';

export async function printReport(title: string, layout: PrintLayout = 'cards'): Promise<void> {
  const previous = document.title;
  document.title = title;
  app.printLayout = layout;
  app.printMode = true;
  await tick();
  window.print();
  app.printMode = false;
  app.printLayout = 'cards';
  document.title = previous;
}
