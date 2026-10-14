<script lang="ts">
  import Info from '@lucide/svelte/icons/info';
  import Lock from '@lucide/svelte/icons/lock';
  import Upload from '@lucide/svelte/icons/upload';
  import { m } from '../lib/i18n';

  let { onfiles, busy = false }: { onfiles: (files: File[]) => void; busy?: boolean } = $props();
  let over = $state(false);

  function pick(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    if (input.files?.length) onfiles([...input.files]);
    input.value = '';
  }

  function drop(e: DragEvent) {
    e.preventDefault();
    over = false;
    const files = [...(e.dataTransfer?.files ?? [])];
    if (files.length) onfiles(files);
  }
</script>

<div
  data-testid="dropzone"
  role="group"
  aria-label={m.steps.files}
  class="dropzone"
  data-over={over || undefined}
  ondragover={(e) => {
    e.preventDefault();
    over = true;
  }}
  ondragleave={() => (over = false)}
  ondrop={drop}
>
  <span class="dropzone-icon"><Upload /></span>
  <p class="max-w-md text-base font-medium">{m.files.dropHint}</p>
  <input id="csv-input" class="visually-hidden" type="file" multiple accept=".csv,text/csv" onchange={pick} disabled={busy} />
  <label for="csv-input" class="btn btn-primary" class:opacity-50={busy}>{m.files.choose}</label>
  <p class="text-caption font-medium text-success"><Lock class="mr-1.5 inline size-4 align-[-3px]" />{m.files.neverLeaves}</p>
</div>

<ul class="mt-4 space-y-2 text-caption text-ink-muted">
  <li class="flex gap-2"><Info class="mt-0.5 size-4 shrink-0 text-link" /><span>{m.files.oneFilePerYear}</span></li>
  <li class="flex gap-2"><Info class="mt-0.5 size-4 shrink-0 text-link" /><span>{m.files.howToExport}</span></li>
</ul>
