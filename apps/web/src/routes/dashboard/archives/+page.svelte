<script lang="ts">
  import { FileArchive, Download, Loader2, Search } from "lucide-svelte";
  import { api } from "$lib/api";
  import Reveal from "$lib/components/landing/Reveal.svelte";

  let q = $state("");
  let page = $state(1);
  let list = $state<any[]>([]);
  let total = $state(0);
  let loading = $state(true);
  let error = $state("");
  let downloadingId = $state<string | null>(null);

  async function load() {
    loading = true;
    error = "";
    try {
      const query: Record<string, string> = {
        page: String(page),
        per_page: "12",
        status: "graduated",
      };
      if (q) query.q = q;
      const res = await api.api.v1.theses.$get({ query });
      const json: any = res.ok ? await res.json() : null;
      list = json?.data ?? [];
      total = json?.meta?.total ?? 0;
    } catch (e) {
      error = "Gagal memuat daftar arsip.";
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    q;
    page;
    load();
  });

  function formatDate(s?: string) {
    if (!s) return "—";
    try {
      return new Date(s).toLocaleDateString("id-ID", {
        day: "2-digit",
        month: "short",
        year: "numeric",
      });
    } catch {
      return s;
    }
  }

  async function handleDownload(id: string) {
    downloadingId = id;
    try {
      window.open(`/api/v1/archives/${id}/download`, "_blank");
    } finally {
      downloadingId = null;
    }
  }
</script>

<div class="space-y-6">
  <Reveal>
    <div>
      <p class="landing-eyebrow">Arsip Skripsi</p>
      <h1 class="mt-2 text-balance landing-heading text-2xl">
        Kumpulan skripsi yang telah <span class="accent-text italic">diarsipkan</span>
      </h1>
    </div>
  </Reveal>

  <Reveal delay={80}>
    <div class="relative w-full max-w-md">
      <Search
        size={16}
        class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-st-muted"
        aria-hidden="true"
      />
      <input
        type="search"
        placeholder="Cari judul / mahasiswa…"
        bind:value={q}
        oninput={() => (page = 1)}
        class="w-full rounded-md border border-st-stroke bg-st-surface py-2 pl-9 pr-3 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
      />
    </div>
  </Reveal>

  {#if error}
    <div
      role="alert"
      class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700"
    >
      {error}
    </div>
  {/if}

  {#if loading}
    <div class="py-10 text-center text-sm text-st-muted">Memuat arsip…</div>
  {:else}
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {#each list as a (a.id)}
        {@const studentName = a.student?.fullName ?? a.student?.full_name ?? "—"}
        {@const archivedAt = a.archivedAt ?? a.archived_at ?? a.graduatedAt ?? a.graduated_at}
        {@const year = a.graduationYear ?? a.graduation_year}
        <article
          class="group rounded-2xl border border-st-stroke bg-st-surface p-5 transition hover:border-(--st-accent-from)/40"
        >
          <div class="mb-3 flex items-center justify-between">
            <div
              class="accent-ring flex h-10 w-10 items-center justify-center rounded-xl bg-(--st-accent-from)/10 text-(--st-accent-to)"
            >
              <FileArchive size={20} aria-hidden="true" />
            </div>
            <button
              type="button"
              disabled={downloadingId === a.id}
              onclick={() => handleDownload(a.id)}
              aria-label="Unduh arsip"
              class="inline-flex h-8 w-8 items-center justify-center rounded-md text-st-muted hover:bg-st-surface-hi hover:text-st-text disabled:opacity-50"
            >
              {#if downloadingId === a.id}
                <Loader2 size={16} class="animate-spin" />
              {:else}
                <Download size={16} />
              {/if}
            </button>
          </div>
          <p class="line-clamp-2 font-medium leading-snug text-st-text">
            {a.title ?? "Skripsi"}
          </p>
          <p class="mt-1 text-sm text-st-muted">{studentName}</p>
          {#if year}
            <p class="mt-2 text-xs text-st-muted">
              Tahun {year} · Diarsipkan {formatDate(archivedAt)}
            </p>
          {:else}
            <p class="mt-2 text-xs text-st-muted">Diarsipkan {formatDate(archivedAt)}</p>
          {/if}
        </article>
      {/each}
      {#if list.length === 0}
        <div
          class="col-span-full rounded-2xl border border-st-stroke bg-st-surface py-12 text-center"
        >
          <FileArchive size={36} class="mx-auto text-st-muted" />
          <p class="mt-3 landing-heading text-lg">
            Tidak ada <span class="accent-text italic">arsip</span> ditemukan
          </p>
          <p class="mt-1 text-sm text-st-muted">Coba ubah kata kunci pencarian Anda.</p>
        </div>
      {/if}
    </div>
  {/if}

  {#if total > 12}
    <div class="flex items-center justify-between pt-2">
      <p class="text-sm text-st-muted">Total {total} arsip · Halaman {page}</p>
      <div class="flex gap-2">
        <button
          type="button"
          disabled={page <= 1}
          onclick={() => (page -= 1)}
          class="inline-flex h-8 items-center rounded-md border border-st-stroke bg-st-surface px-3 text-sm text-st-text hover:bg-st-surface-hi disabled:opacity-50"
        >
          Sebelumnya
        </button>
        <button
          type="button"
          disabled={page >= Math.ceil(total / 12)}
          onclick={() => (page += 1)}
          class="inline-flex h-8 items-center rounded-md border border-st-stroke bg-st-surface px-3 text-sm text-st-text hover:bg-st-surface-hi disabled:opacity-50"
        >
          Berikutnya
        </button>
      </div>
    </div>
  {/if}
</div>
