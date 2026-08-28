<script lang="ts">
  import { ArrowLeft, CheckCircle2, XCircle } from "lucide-svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { api } from "$lib/api";
  import { auth } from "$lib/auth.store";
  import Reveal from "$lib/components/landing/Reveal.svelte";
  import StatusBadge from "$lib/components/dashboard/StatusBadge.svelte";
  import { thesisStatusProps } from "$lib/components/dashboard/thesis-status";

  const id = $derived($page.params.id ?? "");

  let thesis = $state<any>(null);
  let loading = $state(true);
  let error = $state("");

  let notes = $state("");
  let submitError = $state("");
  let submitting = $state(false);

  async function load() {
    if (!id) return;
    loading = true;
    error = "";
    try {
      const res = await api.api.v1.theses[":id"].$get({ param: { id } });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      thesis = (await res.json()) as any;
    } catch (e: any) {
      error = "Gagal memuat detail skripsi.";
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    id;
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

  async function decide(decision: "approved" | "rejected") {
    if (!id || submitting) return;
    submitting = true;
    submitError = "";
    try {
      const res = await api.api.v1.theses[":id"].$put({
        json: { decision, notes: notes || undefined },
        param: { id },
      });
      if (!res.ok) {
        const j: any = await res.json().catch(() => null);
        throw new Error(j?.error?.message ?? `HTTP ${res.status}`);
      }
      void $auth;
      await goto("/theses");
    } catch (e: any) {
      submitError = e?.message ?? "Gagal menyimpan keputusan review.";
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
      <p class="landing-eyebrow">Review Pengajuan</p>
      <h1 class="mt-2 text-balance landing-heading text-2xl">
        Tinjau <span class="accent-text italic">Skripsi</span> Mahasiswa
      </h1>
    </div>
  </Reveal>

  {#if loading}
    <div class="rounded-2xl border border-st-stroke bg-st-surface py-10 text-center text-sm text-st-muted">
      Memuat detail skripsi…
    </div>
  {:else if error || !thesis}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {error || "Skripsi tidak ditemukan."}
    </div>
  {:else}
    {@const sp = thesisStatusProps(thesis.status ?? "")}
    <Reveal>
      <section class="rounded-2xl border border-st-stroke bg-st-surface p-6">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="min-w-0 space-y-1.5">
            <h2 class="text-lg leading-snug text-st-text">{thesis.title}</h2>
            <StatusBadge variant={sp.variant} label={sp.label} />
          </div>
        </div>

        <div class="mt-5 grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <p class="text-xs font-medium uppercase tracking-[0.18em] text-st-muted">Mahasiswa</p>
            <p class="mt-1 font-medium text-st-text">
              {thesis.student?.fullName ?? thesis.student?.full_name ?? "—"}
            </p>
            <p class="text-xs text-st-muted">{thesis.student?.nimNidn ?? thesis.student?.nim_nidn ?? ""}</p>
          </div>
          <div>
            <p class="text-xs font-medium uppercase tracking-[0.18em] text-st-muted">Diajukan</p>
            <p class="mt-1 font-medium text-st-text">
              {formatDate(thesis.submittedAt ?? thesis.submitted_at)}
            </p>
          </div>
          <div>
            <p class="text-xs font-medium uppercase tracking-[0.18em] text-st-muted">Bidang Keahlian</p>
            <p class="mt-1 font-medium text-st-text">{thesis.fieldOfStudy ?? thesis.field_of_study ?? "—"}</p>
          </div>
        </div>

        <div class="mt-5">
          <p class="text-xs font-medium uppercase tracking-[0.18em] text-st-muted">Abstrak</p>
          <p class="mt-1 text-sm leading-relaxed text-st-text">{thesis.abstract}</p>
        </div>
      </section>
    </Reveal>

    <Reveal>
      <section class="rounded-2xl border border-st-stroke bg-st-surface p-6">
        <h2 class="text-base font-semibold text-st-text">Keputusan Review</h2>

        {#if submitError}
          <div
            role="alert"
            class="mt-4 rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700"
          >
            {submitError}
          </div>
        {/if}

        <form
          onsubmit={(e) => {
            e.preventDefault();
          }}
          class="mt-4 space-y-4"
        >
          <div>
            <label for="notes" class="text-xs font-medium uppercase tracking-[0.18em] text-st-muted">
              Catatan untuk Mahasiswa
            </label>
            <textarea
              id="notes"
              rows={4}
              placeholder="Catatan revisi atau alasan penolakan…"
              bind:value={notes}
              class="mt-1 w-full rounded-md border border-st-stroke bg-st-bg px-3 py-2 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
            ></textarea>
          </div>

          <div class="flex flex-wrap gap-3">
            <button
              type="button"
              disabled={submitting}
              onclick={() => decide("approved")}
              class="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-success px-4 text-sm font-medium text-white transition-colors hover:bg-success/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50"
            >
              <CheckCircle2 size={16} aria-hidden="true" />
              {submitting ? "Menyimpan…" : "Setujui"}
            </button>
            <button
              type="button"
              disabled={submitting}
              onclick={() => decide("rejected")}
              class="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-danger px-4 text-sm font-medium text-white transition-colors hover:bg-danger/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50"
            >
              <XCircle size={16} aria-hidden="true" />
              {submitting ? "Menyimpan…" : "Tolak"}
            </button>
            <button
              type="button"
              disabled={submitting}
              onclick={() => goto("/theses")}
              class="inline-flex h-10 items-center justify-center rounded-md border border-st-stroke bg-st-surface px-4 text-sm font-medium text-st-text transition-colors hover:bg-st-surface-hi"
            >
              Batal
            </button>
          </div>
        </form>
      </section>
    </Reveal>
  {/if}
</div>
