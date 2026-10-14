<script lang="ts">
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
  class="rounded-xs border-2 border-dashed px-6 py-8 text-center {over ? 'border-accent bg-accent-soft' : 'border-line bg-sheet'}"
  ondragover={(e) => {
    e.preventDefault();
    over = true;
  }}
  ondragleave={() => (over = false)}
  ondrop={drop}
>
  <p class="text-base">
    {m.files.dropHint}
    <label for="csv-input" class="cursor-pointer font-medium text-accent underline">{m.files.choose}</label>
  </p>
  <input id="csv-input" class="visually-hidden" type="file" multiple accept=".csv,text/csv" onchange={pick} disabled={busy} />
  <p class="mx-auto mt-3 max-w-prose text-sm text-muted">{m.files.oneFilePerYear}</p>
  <p class="mx-auto mt-1 max-w-prose text-sm text-muted">{m.files.howToExport}</p>
  <p class="mt-3 text-sm font-medium text-ok">{m.files.neverLeaves}</p>
</div>
