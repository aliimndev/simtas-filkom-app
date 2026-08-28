<script lang="ts">
  import { BookOpen } from "lucide-svelte";
  import { api } from "$lib/api";
  import StatusBadge from "$lib/components/dashboard/StatusBadge.svelte";
  import { thesisStatusProps } from "$lib/components/dashboard/thesis-status";

  let q = $state("");
  let status = $state("");
  let page = $state(1);
  let list = $state<any[]>([]);
  let total = $state(0);
  let loading = $state(true);
  let error = $state("");

  async function load() {
    loading = true;
    error = "";
    try {
      const query: Record<string, string> = { page: String(page), per_page: "10" };
      if (q) query.q = q;
      if (status) query.status = status;
      const res = await api.api.v1.theses.$get({ query });
      const json: any = res.ok ? await res.json() : null;
      list = json?.data ?? [];
      total = json?.meta?.total ?? 0;
    } catch (e) {
      error = "Gagal memuat daftar skripsi.";
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    q; status; page;
    load();
  });

  function formatDate(s?: string) {
    if (!s) return "—";
    try {
      return new Date(s).toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
    } catch {
      return s;
    }
  }
</script>

<div class="space-y-6">
  <div>
    <p class="landing-eyebrow">Daftar Tugas Akhir</p>
    <h1 class="mt-2 text-balance landing-heading text-2xl">
      Kelola pengajuan dan status <span class="accent-text italic">skripsi</span> mahasiswa
    </h1>
  </div>

  <div class="flex flex-wrap items-center gap-3">
    <label for="thesis-search" class="sr-only">Cari judul atau mahasiswa</label>
    <input
      id="thesis-search"
      type="search"
      placeholder="Cari judul / mahasiswa…"
      bind:value={q}
      oninput={() => (page = 1)}
      class="w-full max-w-xs rounded-md border border-st-stroke bg-st-surface px-3 py-2 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
    />
    <label for="thesis-status" class="sr-only">Filter status skripsi</label>
    <select
      id="thesis-status"
      bind:value={status}
      onchange={() => (page = 1)}
      class="w-44 rounded-md border border-st-stroke bg-st-surface px-3 py-2 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
    >
      <option value="">Semua status</option>
      <option value="submitted">Diajukan</option>
      <option value="approved">Disetujui</option>
      <option value="in_progress">Bimbingan</option>
      <option value="seminar_ready">Siap Seminar</option>
      <option value="defense_ready">Siap Sidang</option>
      <option value="graduated">Lulus</option>
      <option value="cancelled">Dibatalkan</option>
    </select>
  </div>

  {#if error}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {error}
    </div>
  {/if}

  {#if loading}
    <div class="py-10 text-center text-sm text-st-muted">Memuat daftar skripsi…</div>
  {:else}
    <div class="space-y-3">
      {#each list as t (t.id)}
        {@const sp = thesisStatusProps(t.status ?? "")}
        <article class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-st-stroke bg-st-surface p-5">
          <div class="flex min-w-0 items-center gap-3">
            <div class="accent-ring flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-(--st-accent-from)/10 text-(--st-accent-to)">
              <BookOpen size={20} />
            </div>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <p class="font-medium text-st-text">{t.student?.fullName ?? t.student?.full_name ?? "—"}</p>
                <StatusBadge variant={sp.variant} label={sp.label} />
              </div>
              <p class="mt-0.5 line-clamp-1 text-sm text-st-muted">{t.title}</p>
              <p class="mt-1 text-xs text-st-muted">
                Diajukan {formatDate(t.submittedAt ?? t.submitted_at)}
                {#if t.supervisors?.length}
                  · Pembimbing: {(t.supervisors ?? []).map((s: any) => s.fullName ?? s.full_name).join(", ")}
                {/if}
              </p>
            </div>
          </div>
          <div class="flex gap-2">
            {#if t.status === "submitted"}
              <a href="/dashboard/theses/{t.id}/review" class="inline-flex h-8 items-center justify-center rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground hover:bg-primary-700">Review</a>
            {/if}
            {#if t.status === "approved" || t.status === "in_progress"}
              <a href="/dashboard/theses/{t.id}/assign" class="inline-flex h-8 items-center justify-center rounded-md border border-st-stroke bg-st-surface px-3 text-sm font-medium text-st-text hover:bg-st-surface-hi">Atur Pembimbing</a>
            {/if}
          </div>
        </article>
      {/each}
      {#if list.length === 0}
        <div class="rounded-2xl border border-st-stroke bg-st-surface py-12 text-center">
          <BookOpen size={36} class="mx-auto text-st-muted" />
          <p class="mt-3 landing-heading text-lg">Tidak ada <span class="accent-text italic">skripsi</span> ditemukan</p>
          <p class="mt-1 text-sm text-st-muted">Coba ubah filter atau kata kunci pencarian Anda.</p>
        </div>
      {/if}
    </div>
  {/if}

  {#if total > 10}
    <div class="flex items-center justify-between pt-2">
      <p class="text-sm text-st-muted">Total {total} skripsi · Halaman {page}</p>
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
          disabled={page >= Math.ceil(total / 10)}
          onclick={() => (page += 1)}
          class="inline-flex h-8 items-center rounded-md border border-st-stroke bg-st-surface px-3 text-sm text-st-text hover:bg-st-surface-hi disabled:opacity-50"
        >
          Berikutnya
        </button>
      </div>
    </div>
  {/if}
</div>
