<script lang="ts">
  import { ArrowLeft, Send } from "lucide-svelte";
  import { goto } from "$app/navigation";
  import { api } from "$lib/api";
  import Reveal from "$lib/components/landing/Reveal.svelte";

  let title = $state("");
  let abstract = $state("");
  let fieldOfStudy = $state("");
  let thesisType = $state("skripsi");
  let loading = $state(false);
  let error = $state<string | null>(null);

  const wordCount = (value: string) => value.trim().split(/\s+/).filter(Boolean).length;

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    error = null;

    if (wordCount(title) < 10) {
      error = "Judul minimal 10 kata.";
      return;
    }
    if (wordCount(abstract) < 100) {
      error = "Abstrak minimal 100 kata.";
      return;
    }

    loading = true;
    try {
      const response = await api.api.v1.theses.$post({
        json: {
          title: title.trim(),
          abstract: abstract.trim(),
          fieldOfStudy: fieldOfStudy.trim() || undefined,
          thesisType,
        },
      });
      if (!response.ok) {
        const body = await response.json().catch(() => null);
        throw new Error((body as any)?.error?.message ?? "Gagal mengajukan judul skripsi.");
      }
      await goto("/dashboard/thesis");
    } catch (err) {
      error = err instanceof Error ? err.message : "Gagal mengajukan judul skripsi.";
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Ajukan Judul — SIMTAS FILKOM</title>
  <meta name="description" content="Ajukan judul Tugas Akhir Skripsi baru di SIMTAS FILKOM." />
</svelte:head>

<div class="mx-auto w-full max-w-3xl space-y-6">
  <Reveal>
    <div>
      <a href="/dashboard/thesis" class="mb-4 inline-flex items-center gap-1.5 text-sm font-medium text-st-muted transition-colors hover:text-st-text">
        <ArrowLeft size={16} aria-hidden="true" /> Kembali ke Tugas Akhir Saya
      </a>
      <p class="landing-eyebrow">Pengajuan Tugas Akhir</p>
      <h1 class="mt-2 text-balance landing-heading text-2xl">Ajukan judul <span class="accent-text italic">skripsi</span> baru</h1>
      <p class="mt-2 text-sm text-st-muted">Lengkapi data berikut untuk mengirim pengajuan kepada Kaprodi.</p>
    </div>
  </Reveal>

  {#if error}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-4 py-3 text-sm text-danger-700">
      {error}
    </div>
  {/if}

  <Reveal>
    <form onsubmit={handleSubmit} class="space-y-5 rounded-2xl border border-st-stroke bg-st-surface p-6">
      <div>
        <label for="thesis-title" class="mb-1.5 block text-sm font-medium text-st-text">Judul skripsi <span class="text-danger-700">*</span></label>
        <input
          id="thesis-title"
          type="text"
          bind:value={title}
          required
          maxlength="500"
          placeholder="Masukkan judul Tugas Akhir minimal 10 kata"
          class="w-full rounded-md border border-st-stroke bg-st-bg px-3 py-2.5 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
        />
        <p class="mt-1 text-xs text-st-muted">wordCount(title)} kata · maksimal 500 karakter</p>
      </div>

      <div>
        <label for="thesis-abstract" class="mb-1.5 block text-sm font-medium text-st-text">Abstrak <span class="text-danger-700">*</span></label>
        <textarea
          id="thesis-abstract"
          bind:value={abstract}
          required
          rows={9}
          placeholder="Jelaskan latar belakang, tujuan, metode, dan hasil yang diharapkan. Minimal 100 kata."
          class="w-full rounded-md border border-st-stroke bg-st-bg px-3 py-2.5 text-sm leading-relaxed text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
        ></textarea>
        <p class="mt-1 text-xs text-st-muted">wordCount(abstract)} kata · minimal 100 kata</p>
      </div>

      <div class="grid grid-cols-1 gap-5 sm:grid-cols-2">
        <div>
          <label for="field-of-study" class="mb-1.5 block text-sm font-medium text-st-text">Bidang keahlian</label>
          <input
            id="field-of-study"
            type="text"
            bind:value={fieldOfStudy}
            maxlength="100"
            placeholder="Contoh: Rekayasa Perangkat Lunak"
            class="w-full rounded-md border border-st-stroke bg-st-bg px-3 py-2.5 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
          />
        </div>
        <div>
          <label for="thesis-type" class="mb-1.5 block text-sm font-medium text-st-text">Jenis Tugas Akhir <span class="text-danger-700">*</span></label>
          <select
            id="thesis-type"
            bind:value={thesisType}
            class="w-full rounded-md border border-st-stroke bg-st-bg px-3 py-2.5 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
          >
            <option value="skripsi">Skripsi</option>
            <option value="tugas_akhir">Tugas Akhir</option>
          </select>
        </div>
      </div>

      <div class="flex flex-wrap justify-end gap-3 border-t border-st-stroke pt-5">
        <a href="/dashboard/thesis" class="inline-flex h-10 items-center justify-center rounded-md border border-st-stroke bg-st-surface px-4 text-sm font-medium text-st-text transition-colors hover:bg-st-surface-hi">Batal</a>
        <button
          type="submit"
          disabled={loading}
          class="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary-700 disabled:opacity-50"
        >
          <Send size={16} aria-hidden="true" />
          {loading ? "Mengirim…" : "Kirim Pengajuan"}
        </button>
      </div>
    </form>
  </Reveal>
</div>
