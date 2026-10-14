import { tick } from 'svelte';
import { app } from './state.svelte';

export async function printTables(title: string): Promise<void> {
  const previous = document.title;
  document.title = title;
  app.printLayout = 'tables';
  app.printMode = true;
  await tick();
  window.print();
  app.printMode = false;
  app.printLayout = 'cards';
  document.title = previous;
}
