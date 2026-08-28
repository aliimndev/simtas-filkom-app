<script lang="ts">
  import { ArrowLeft, Users } from "lucide-svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { api } from "$lib/api";
  import { auth } from "$lib/auth.store";
  import Reveal from "$lib/components/landing/Reveal.svelte";

  const id = $derived($page.params.id ?? "");

  let thesis = $state<any>(null);
  let thesisLoading = $state(true);
  let thesisError = $state("");

  let lecturers = $state<any[]>([]);
  let lecturersLoading = $state(true);
  let lecturersError = $state("");

  let selected = $state<string[]>([]);
  let q = $state("");
  let submitError = $state("");
  let submitting = $state(false);

  let prevSupKey = $state<string | null>(null);

  const filtered = $derived.by(() => {
    const term = q.trim().toLowerCase();
    if (!term) return lecturers;
    return lecturers.filter((l) => {
      const name = (l.fullName ?? l.full_name ?? "").toLowerCase();
      const nidn = (l.nimNidn ?? l.nim_nidn ?? "").toLowerCase();
      return name.includes(term) || nidn.includes(term);
    });
  });

  function supKeyOf(t: any) {
    return (t?.supervisors ?? []).map((s: any) => s.id).sort().join(",");
  }

  async function loadThesis() {
    if (!id) return;
    thesisLoading = true;
    thesisError = "";
    try {
      const res = await api.api.v1.theses[":id"].$get({ param: { id } });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      thesis = (await res.json()) as any;
    } catch (e: any) {
      thesisError = "Gagal memuat detail skripsi.";
    } finally {
      thesisLoading = false;
    }
  }

  async function loadLecturers() {
    lecturersLoading = true;
    lecturersError = "";
    try {
      const query: Record<string, string> = { role: "DOSEN_PEMBIMBING", per_page: "100" };
      const res = await api.api.v1.users.$get({ query });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const json: any = await res.json();
      lecturers = json?.data ?? [];
    } catch (e: any) {
      lecturersError = "Gagal memuat daftar dosen pembimbing.";
    } finally {
      lecturersLoading = false;
    }
  }

  $effect(() => {
    id;
    loadThesis();
    loadLecturers();
  });

  $effect(() => {
    const key = supKeyOf(thesis);
    if (thesis && key !== prevSupKey) {
      prevSupKey = key;
      selected = (thesis.supervisors ?? []).map((s: any) => s.id);
    }
  });

  function toggle(lecturerId: string) {
    selected = selected.includes(lecturerId)
      ? selected.filter((x) => x !== lecturerId)
      : [...selected, lecturerId];
  }

  function isSupervisor(lecturerId: string) {
    return (thesis?.supervisors ?? []).some((s: any) => s.id === lecturerId);
  }

  function formatDate(s?: string) {
    if (!s) return "—";
    try {
      return new Date(s).toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
    } catch {
      return s;
    }
  }

  async function submit(e: Event) {
    e.preventDefault();
    if (!id || selected.length === 0 || submitting) return;
    submitting = true;
    submitError = "";
    try {
      const res = await api.api.v1.theses[":id"].supervisors.$post({
        json: { supervisorIds: selected },
        param: { id },
      });
      if (!res.ok) {
        const j: any = await res.json().catch(() => null);
        throw new Error(j?.error?.message ?? `HTTP ${res.status}`);
      }
      void $auth;
      await goto("/theses");
    } catch (e: any) {
      submitError = e?.message ?? "Gagal menyimpan pembimbing.";
    } finally {
      submitting = false;
    }
  }
</script>

<div class="mx-auto w-full max-w-3xl space-y-6">
  <Reveal>
    <div>
      <a
        href="/theses"
        class="mb-4 inline-flex items-center gap-1.5 text-sm font-medium text-st-muted transition-colors hover:text-st-text"
      >
        <ArrowLeft size={16} aria-hidden="true" /> Kembali ke Daftar
      </a>
      <h1 class="text-balance landing-heading text-2xl">
        Atur Dosen <span class="accent-text italic">Pembimbing</span>
      </h1>
      {#if thesis}
        <p class="mt-1 text-sm text-st-muted">
          {thesis.title} · {thesis.student?.fullName ?? thesis.student?.full_name ?? "—"} ·
          {formatDate(thesis.submittedAt ?? thesis.submitted_at)}
        </p>
      {/if}
    </div>
  </Reveal>

  {#if submitError}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {submitError}
    </div>
  {/if}

  <Reveal>
    <section class="rounded-2xl border border-st-stroke bg-st-surface p-6">
      <h2 class="flex items-center gap-2 text-base font-semibold text-st-text">
        <Users size={18} class="text-primary" aria-hidden="true" /> Pilih Pembimbing (maks. 2)
      </h2>

      <form onsubmit={submit} class="mt-4 space-y-4">
        <div>
          <label for="lecturer-search" class="text-xs font-medium uppercase tracking-[0.18em] text-st-muted">
            Cari Dosen
          </label>
          <input
            id="lecturer-search"
            type="search"
            placeholder="Nama dosen…"
            bind:value={q}
            class="mt-1 h-10 w-full rounded-md border border-st-stroke bg-st-bg px-3 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>

        {#if thesisError}
          <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
            {thesisError}
          </div>
        {/if}
        {#if lecturersError}
          <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
            {lecturersError}
          </div>
        {/if}

        {#if thesisLoading || lecturersLoading}
          <div class="rounded-2xl border border-st-stroke bg-st-bg py-10 text-center text-sm text-st-muted">
            Memuat data…
          </div>
        {:else if filtered.length === 0}
          <div class="rounded-2xl border border-st-stroke bg-st-bg py-12 text-center">
            <Users size={36} class="mx-auto text-st-muted" aria-hidden="true" />
            <p class="mt-3 landing-heading text-lg">
              Tidak ada <span class="accent-text italic">dosen</span> ditemukan
            </p>
            <p class="mt-1 text-sm text-st-muted">
              {#if q}
                Coba kata kunci lain.
              {:else}
                Belum ada dosen pembimbing yang tersedia.
              {/if}
            </p>
          </div>
        {:else}
          <div class="grid grid-cols-1 gap-2 sm:grid-cols-2">
            {#each filtered as l (l.id)}
              {@const isSelected = selected.includes(l.id)}
              {@const already = isSupervisor(l.id)}
              {@const disabled = !isSelected && selected.length >= 2}
              <button
                type="button"
                onclick={() => {
                  if (isSelected) toggle(l.id);
                  else if (selected.length < 2) toggle(l.id);
                }}
                aria-pressed={isSelected}
                class="flex items-center justify-between gap-3 rounded-lg border p-3 text-left transition-colors {isSelected
                  ? 'border-primary bg-primary-50'
                  : 'border-st-stroke hover:bg-st-surface-hi'} {disabled ? 'cursor-not-allowed opacity-50' : ''}"
              >
                <div class="min-w-0">
                  <p class="truncate text-sm font-medium text-st-text">
                    {l.fullName ?? l.full_name ?? "—"}
                  </p>
                  <p class="truncate text-xs text-st-muted">
                    {l.nimNidn ?? l.nim_nidn ?? ""}
                  </p>
                </div>
                {#if already}
                  <span class="rounded-full bg-primary-50 px-2 py-0.5 text-xs font-medium text-primary">
                    Pembimbing
                  </span>
                {:else if isSelected}
                  <span class="rounded-full bg-success-50 px-2 py-0.5 text-xs font-medium text-success">
                    Dipilih
                  </span>
                {/if}
              </button>
            {/each}
          </div>
        {/if}

        <div class="flex gap-3 pt-2">
          <button
            type="submit"
            disabled={selected.length === 0 || submitting}
            class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50"
          >
            {submitting ? "Menyimpan…" : "Simpan Pembimbing"}
          </button>
          <button
            type="button"
            onclick={() => goto("/theses")}
            class="inline-flex h-10 items-center justify-center rounded-md border border-st-stroke bg-st-surface px-4 text-sm font-medium text-st-text transition-colors hover:bg-st-surface-hi"
          >
            Batal
          </button>
        </div>
      </form>
    </section>
  </Reveal>
</div>
