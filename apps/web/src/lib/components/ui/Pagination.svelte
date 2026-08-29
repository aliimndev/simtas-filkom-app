<script lang="ts">
  let {
    total,
    page,
    pageSize,
    onPage,
  }: {
    total: number;
    page: number;
    pageSize: number;
    onPage: (next: number) => void;
  } = $props();

  const lastPage = $derived(Math.max(1, Math.ceil(total / pageSize)));
  const show = $derived(total > pageSize);
</script>

{#if show}
  <div class="flex flex-wrap items-center justify-between gap-3 pt-2">
    <p class="text-sm text-st-muted">Total {total} · Halaman {page}</p>
    <div class="flex gap-2">
      <button
        type="button"
        disabled={page <= 1}
        onclick={() => onPage(page - 1)}
        class="inline-flex h-9 items-center justify-center rounded-md border border-st-stroke bg-st-surface px-3 text-sm text-st-text transition-colors hover:bg-st-surface-hi disabled:opacity-50"
      >
        Sebelumnya
      </button>
      <button
        type="button"
        disabled={page >= lastPage}
        onclick={() => onPage(page + 1)}
        class="inline-flex h-9 items-center justify-center rounded-md border border-st-stroke bg-st-surface px-3 text-sm text-st-text transition-colors hover:bg-st-surface-hi disabled:opacity-50"
      >
        Berikutnya
      </button>
    </div>
  </div>
{/if}
