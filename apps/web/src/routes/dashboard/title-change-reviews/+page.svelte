<script lang="ts">
  import { ArrowRight, CheckCircle2, Inbox, PencilLine, XCircle } from "lucide-svelte";
  import { auth } from "$lib/auth.store";
  import { api } from "$lib/api";
  import StatusBadge from "$lib/components/ui/StatusBadge.svelte";
  import Dialog from "$lib/components/ui/Dialog.svelte";
  import { titleChangeStatusProps } from "$lib/constants/statuses";
  import { formatDate } from "$lib/utils/format";
  import { apiErrorMessage } from "$lib/utils/errors";

  const user = $derived($auth.user);
  const role = $derived(String(user?.role ?? ""));
  const allowed = $derived(role === "DOSEN_PEMBIMBING" || role === "dosen_pembimbing");

  type TcrItem = {
    id: string;
    thesisId: string;
    previousTitle: string;
    requestedTitle: string;
    reason?: string | null;
    status: "PENDING" | "APPROVED" | "REJECTED" | "CANCELLED";
    requestedById: string;
    createdAt: string;
    reviewNotes?: string | null;
  };

  let list = $state<TcrItem[]>([]);
  let loading = $state(true);
  let error = $state("");

  let reviewTarget = $state<TcrItem | null>(null);
  let reviewOpen = $state(false);
  let notes = $state("");
  let notesTouched = $state(false);
  let submitting = $state(false);
  let pendingDecision = $state<"approve" | "reject" | null>(null);
  let submitError = $state("");
  let toast = $state<{ kind: "success" | "danger"; title: string; description: string } | null>(null);

  async function load() {
    loading = true;
    error = "";
    try {
      const res = await api.api.v1.titleChangeRequests.$get({});
      const json: any = res.ok ? await res.json() : null;
      const all: TcrItem[] = (json?.data ?? []) as TcrItem[];
      list = all.filter((r) => r.status === "PENDING");
    } catch (e) {
      error = "Gagal memuat antrian review.";
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    allowed;
    if (allowed) load();
  });

  function openReview(r: TcrItem) {
    reviewTarget = r;
    notes = "";
    notesTouched = false;
    submitError = "";
    pendingDecision = null;
    reviewOpen = true;
  }

  function closeReview(force = false) {
    if (submitting && !force) return;
    reviewTarget = null;
    notes = "";
    notesTouched = false;
    submitError = "";
    pendingDecision = null;
    reviewOpen = false;
  }

  async function runReview(decision: "approve" | "reject") {
    if (!reviewTarget) return;
    if (decision === "reject" && !notes.trim()) {
      notesTouched = true;
      return;
    }
    submitting = true;
    pendingDecision = decision;
    submitError = "";
    try {
      const res = await api.api.v1.titleChangeRequests[":id"].review.$patch({
        param: { id: reviewTarget.id },
        json: {
          decision: decision === "approve" ? "APPROVED" : "REJECTED",
          reviewNotes: notes.trim() || undefined,
        },
      });
      if (!res.ok) throw await res.json().catch(() => ({}));
      toast = {
        kind: "success",
        title: decision === "approve" ? "Perubahan judul disetujui" : "Perubahan judul ditolak",
        description:
          decision === "approve"
            ? "Judul skripsi telah diperbarui dan mahasiswa akan menerima notifikasi email."
            : "Mahasiswa akan menerima notifikasi email beserta catatan Anda.",
      };
      closeReview(true);
      await load();
      setTimeout(() => (toast = null), 4000);
    } catch (e: any) {
      submitError = apiErrorMessage(e, "Gagal memproses review.");
    } finally {
      submitting = false;
    }
  }

  const notesInvalid = $derived(notesTouched && !notes.trim());
</script>

<div class="space-y-6">
  <div>
    <p class="landing-eyebrow">Review Perubahan Judul</p>
    <h1 class="mt-2 text-balance landing-heading text-2xl">
      Permintaan perubahan <span class="accent-text italic">judul</span> dari mahasiswa bimbingan
    </h1>
    <p class="mt-1.5 text-sm text-st-muted">Permintaan yang menunggu keputusan Anda</p>
  </div>

  {#if !allowed}
    <div role="alert" class="rounded-2xl border border-st-stroke bg-st-surface p-6 text-sm text-st-muted">
      Halaman ini hanya tersedia untuk peran Dosen Pembimbing.
    </div>
  {:else if error}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {error}
    </div>
  {:else if loading}
    <div class="rounded-2xl border border-st-stroke bg-st-surface p-8 text-center text-sm text-st-muted">
      Memuat permintaan…
    </div>
  {:else if list.length === 0}
    <div class="rounded-2xl border border-st-stroke bg-st-surface p-10 text-center">
      <Inbox size={36} class="mx-auto text-st-muted" />
      <p class="mt-4 landing-heading text-lg">Tidak ada <span class="accent-text italic">permintaan</span> menunggu</p>
      <p class="mt-1 text-sm text-st-muted">
        Belum ada pengajuan perubahan judul dari mahasiswa bimbingan Anda.
      </p>
    </div>
  {:else}
    <div class="space-y-3">
      {#each list as r (r.id)}
        {@const rb = titleChangeStatusProps(r.status)}
        <article class="rounded-2xl border border-st-stroke bg-st-surface p-5">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <p class="font-medium text-st-text">Mahasiswa</p>
                <StatusBadge variant={rb.variant} label={rb.label} />
              </div>
              <p class="mt-1 text-xs text-st-muted">Diajukan {formatDate(r.createdAt)}</p>

              <div class="mt-3 flex flex-col gap-2 rounded-lg border border-st-stroke p-3 text-sm sm:flex-row sm:items-center">
                <div class="min-w-0 flex-1">
                  <p class="text-xs font-medium text-st-muted">Judul Saat Ini</p>
                  <p class="mt-0.5 line-through decoration-st-muted/40 text-st-text">{r.previousTitle}</p>
                </div>
                <ArrowRight size={16} class="hidden h-4 w-4 shrink-0 text-st-muted sm:block" />
                <div class="min-w-0 flex-1">
                  <p class="text-xs font-medium text-primary">Judul Baru</p>
                  <p class="mt-0.5 text-st-text">{r.requestedTitle}</p>
                </div>
              </div>

              {#if r.reason}
                <p class="mt-2 text-sm text-st-muted">
                  <span class="font-medium text-st-text">Alasan: </span>
                  {r.reason}
                </p>
              {/if}
            </div>
            <button
              type="button"
              onclick={() => openReview(r)}
              class="inline-flex h-8 items-center justify-center gap-1.5 rounded-md border border-st-stroke bg-st-surface px-3 text-sm font-medium text-st-text hover:bg-st-surface-hi"
            >
              <PencilLine size={14} /> Review
            </button>
          </div>
        </article>
      {/each}
    </div>
  {/if}
</div>

<Dialog bind:open={reviewOpen} labelledBy="review-title" initialFocus="#review-notes">
  <h2 id="review-title" class="landing-heading text-lg">
    Konfirmasi Perubahan Judul
  </h2>
  <p class="mt-1 text-sm text-st-muted">Tinjau pengajuan mahasiswa sebelum mengambil keputusan.</p>

  <div class="mt-4 grid grid-cols-1 gap-3 text-sm sm:grid-cols-2">
    <div class="rounded-lg bg-st-bg p-3">
      <p class="text-xs font-medium text-st-muted">Judul Sebelumnya</p>
      <p class="mt-0.5 leading-snug text-st-text">{reviewTarget?.previousTitle}</p>
    </div>
    <div class="rounded-lg bg-primary-50 p-3">
      <p class="text-xs font-medium text-primary">Judul Baru</p>
      <p class="mt-0.5 leading-snug text-st-text">{reviewTarget?.requestedTitle}</p>
    </div>
  </div>

  {#if reviewTarget?.reason}
    <div class="mt-3 rounded-lg border border-st-stroke p-3 text-sm">
      <p class="text-xs font-medium text-st-muted">Alasan Mahasiswa</p>
      <p class="mt-0.5 text-st-text">{reviewTarget.reason}</p>
    </div>
  {/if}

  <div class="mt-4">
    <label for="review-notes" class="text-sm font-medium text-st-text">Catatan Pembimbing</label>
    <textarea
      id="review-notes"
      rows="3"
      placeholder="Catatan untuk mahasiswa (wajib jika menolak)"
      bind:value={notes}
      oninput={() => {
        if (notesTouched) notesTouched = false;
      }}
      class="mt-1 w-full rounded-md border bg-st-bg px-3 py-2 text-sm text-st-text outline-none focus-visible:ring-2 focus-visible:ring-ring {notesInvalid ? 'border-danger' : 'border-st-stroke'}"
    ></textarea>
    {#if notesInvalid}
      <p class="mt-1 text-xs text-danger-700">Catatan wajib diisi saat menolak perubahan judul.</p>
    {/if}
  </div>

  {#if submitError}
    <div role="alert" class="mt-3 rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {submitError}
    </div>
  {/if}

  <div class="mt-5 flex flex-wrap items-center justify-between gap-2 border-t border-st-stroke pt-4">
    <button
      type="button"
      onclick={() => closeReview()}
      disabled={submitting}
      class="inline-flex h-8 items-center rounded-md px-3 text-sm text-st-muted hover:bg-st-surface-hi disabled:opacity-50"
    >
      Kembali
    </button>
    <div class="flex gap-2">
      <button
        type="button"
        disabled={submitting}
        onclick={() => runReview("reject")}
        class="inline-flex h-8 items-center gap-1.5 rounded-md bg-danger px-3 text-sm font-medium text-white hover:bg-danger-700 disabled:opacity-50"
      >
        <XCircle size={14} />
        {submitting && pendingDecision === "reject" ? "Memproses…" : "Tolak"}
      </button>
      <button
        type="button"
        disabled={submitting}
        onclick={() => runReview("approve")}
        class="inline-flex h-8 items-center gap-1.5 rounded-md bg-success px-3 text-sm font-medium text-white hover:bg-success-700 disabled:opacity-50"
      >
        <CheckCircle2 size={14} />
        {submitting && pendingDecision === "approve" ? "Memproses…" : "Setujui"}
      </button>
    </div>
  </div>
</Dialog>

{#if toast}
  <div class="fixed bottom-4 right-4 z-50 max-w-sm rounded-2xl border border-st-stroke bg-st-surface p-4 shadow-lg">
    <p class="text-sm font-medium text-st-text">{toast.title}</p>
    <p class="mt-1 text-xs text-st-muted">{toast.description}</p>
  </div>
{/if}
