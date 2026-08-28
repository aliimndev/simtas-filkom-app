<script lang="ts">
  import { CalendarPlus, CheckCircle2, CalendarRange } from "lucide-svelte";
  import { api } from "$lib/api";
  import { auth } from "$lib/auth.store";
  import Reveal from "$lib/components/landing/Reveal.svelte";

  type AcademicYear = {
    id: string;
    name: string;
    semester: "ganjil" | "genap";
    startDate: string;
    endDate: string;
    isActive: boolean;
  };

  let list = $state<AcademicYear[]>([]);
  let loading = $state(true);
  let error = $state("");

  let showForm = $state(false);
  let form = $state({ name: "", semester: "ganjil" as "ganjil" | "genap", startDate: "", endDate: "" });
  let createError = $state("");
  let creating = $state(false);
  let activatingId = $state<string | null>(null);

  async function load() {
    loading = true;
    error = "";
    try {
      const res = await api.api.v1.academicYears.$get({});
      if (!res.ok) {
        error = "Gagal memuat daftar tahun akademik.";
        list = [];
        return;
      }
      const json: any = await res.json();
      list = Array.isArray(json) ? json : (json?.data ?? []);
    } catch (e) {
      error = "Gagal memuat daftar tahun akademik.";
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    auth; // track auth state for reactive load on login
    load();
  });

  async function submitCreate(e: Event) {
    e.preventDefault();
    createError = "";
    creating = true;
    try {
      const res = await api.api.v1.academicYears.$post({
        json: {
          name: form.name,
          semester: form.semester,
          startDate: form.startDate,
          endDate: form.endDate,
        },
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        createError = (body as any)?.error?.message ?? "Gagal membuat tahun akademik.";
        return;
      }
      showForm = false;
      form = { name: "", semester: "ganjil", startDate: "", endDate: "" };
      await load();
    } catch (e: any) {
      createError = e?.message ?? "Gagal membuat tahun akademik.";
    } finally {
      creating = false;
    }
  }

  async function activate(id: string) {
    activatingId = id;
    try {
      const res = await (api.api.v1.academicYears as any)[":id"].activate.$patch({
        param: { id },
      });
      if (res.ok) await load();
      else error = "Gagal mengaktifkan tahun akademik.";
    } catch (e) {
      error = "Gagal mengaktifkan tahun akademik.";
    } finally {
      activatingId = null;
    }
  }

  function formatDate(s?: string) {
    if (!s) return "—";
    try {
      return new Date(s).toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
    } catch {
      return s;
    }
  }

  const semesterLabel = (s: string) => (s === "ganjil" ? "Ganjil" : s === "genap" ? "Genap" : "—");
</script>

<div class="space-y-6">
  <Reveal>
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <p class="landing-eyebrow">Tahun Akademik</p>
        <h1 class="mt-2 text-balance landing-heading text-2xl">
          Kelola periode <span class="accent-text italic">tahun akademik</span> aktif
        </h1>
      </div>
      <button
        type="button"
        onclick={() => (showForm = !showForm)}
        class="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
      >
        <CalendarPlus size={16} aria-hidden="true" />
        Tambah Tahun Akademik
      </button>
    </div>
  </Reveal>

  {#if showForm}
    <Reveal>
      <div class="rounded-2xl border border-st-stroke bg-st-surface p-6">
        {#if createError}
          <div role="alert" class="mb-4 rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
            {createError}
          </div>
        {/if}
        <form class="grid grid-cols-1 gap-4 sm:grid-cols-2" onsubmit={submitCreate}>
          <div class="sm:col-span-2">
            <label for="ay-name" class="mb-1 block text-sm font-medium text-st-text">
              Nama <span class="text-danger-700">*</span>
            </label>
            <input
              id="ay-name"
              type="text"
              required
              placeholder="mis. 2026/2027 Ganjil"
              bind:value={form.name}
              class="w-full rounded-md border border-st-stroke bg-st-bg px-3 py-2 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
          <div>
            <label for="ay-semester" class="mb-1 block text-sm font-medium text-st-text">
              Semester <span class="text-danger-700">*</span>
            </label>
            <select
              id="ay-semester"
              required
              bind:value={form.semester}
              class="w-full rounded-md border border-st-stroke bg-st-bg px-3 py-2 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
            >
              <option value="ganjil">Ganjil</option>
              <option value="genap">Genap</option>
            </select>
          </div>
          <div></div>
          <div>
            <label for="ay-start" class="mb-1 block text-sm font-medium text-st-text">
              Mulai <span class="text-danger-700">*</span>
            </label>
            <input
              id="ay-start"
              type="date"
              required
              bind:value={form.startDate}
              class="w-full rounded-md border border-st-stroke bg-st-bg px-3 py-2 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
          <div>
            <label for="ay-end" class="mb-1 block text-sm font-medium text-st-text">
              Selesai <span class="text-danger-700">*</span>
            </label>
            <input
              id="ay-end"
              type="date"
              required
              bind:value={form.endDate}
              class="w-full rounded-md border border-st-stroke bg-st-bg px-3 py-2 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
            />
          </div>
          <div class="flex gap-2 sm:col-span-2">
            <button
              type="submit"
              disabled={creating}
              class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary-700 disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              {creating ? "Menyimpan…" : "Simpan"}
            </button>
            <button
              type="button"
              onclick={() => (showForm = false)}
              class="inline-flex h-10 items-center justify-center rounded-md px-4 text-sm font-medium text-st-text hover:bg-st-surface-hi focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
              Batal
            </button>
          </div>
        </form>
      </div>
    </Reveal>
  {/if}

  {#if error}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {error}
    </div>
  {/if}

  {#if loading}
    <div class="rounded-2xl border border-st-stroke bg-st-surface py-12 text-center text-sm text-st-muted">
      Memuat tahun akademik…
    </div>
  {:else if list.length === 0}
    <div class="rounded-2xl border border-st-stroke bg-st-surface py-12 text-center">
      <CalendarRange size={36} class="mx-auto text-st-muted" aria-hidden="true" />
      <p class="mt-3 landing-heading text-lg">
        Belum ada <span class="accent-text italic">tahun akademik</span>
      </p>
      <p class="mt-1 text-sm text-st-muted">Tambahkan periode tahun akademik untuk memulai.</p>
    </div>
  {:else}
    <Reveal>
      <div class="space-y-3">
        {#each list as y (y.id)}
          <article class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-st-stroke bg-st-surface p-5">
            <div class="flex min-w-0 items-start gap-3">
              <div class="accent-ring flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-(--st-accent-from)/10 text-(--st-accent-to)">
                <CalendarRange size={20} aria-hidden="true" />
              </div>
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="font-medium text-st-text">{y.name}</p>
                  <span class="rounded-full border border-st-stroke bg-st-surface-hi px-2.5 py-0.5 text-[11px] font-medium text-st-muted">
                    {semesterLabel(y.semester)}
                  </span>
                  {#if y.isActive}
                    <span class="inline-flex items-center gap-1.5 rounded-full border border-success/20 bg-success-50 px-2.5 py-0.5 text-[11px] font-medium text-success">
                      <span class="h-1.5 w-1.5 rounded-full bg-success"></span>
                      Aktif
                    </span>
                  {:else}
                    <span class="inline-flex items-center gap-1.5 rounded-full border border-st-stroke bg-st-surface-hi px-2.5 py-0.5 text-[11px] font-medium text-st-muted">
                      <span class="h-1.5 w-1.5 rounded-full bg-muted-foreground"></span>
                      Tidak aktif
                    </span>
                  {/if}
                </div>
                <p class="mt-1 text-xs text-st-muted">
                  {formatDate(y.startDate)} — {formatDate(y.endDate)}
                </p>
              </div>
            </div>
            {#if !y.isActive}
              <button
                type="button"
                disabled={activatingId === y.id}
                onclick={() => activate(y.id)}
                class="inline-flex h-8 items-center justify-center gap-1.5 rounded-md border border-st-stroke bg-st-surface px-3 text-sm font-medium text-st-text hover:bg-st-surface-hi disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                <CheckCircle2 size={14} aria-hidden="true" />
                {activatingId === y.id ? "Mengaktifkan…" : "Jadikan Aktif"}
              </button>
            {/if}
          </article>
        {/each}
      </div>
    </Reveal>
  {/if}
</div>
