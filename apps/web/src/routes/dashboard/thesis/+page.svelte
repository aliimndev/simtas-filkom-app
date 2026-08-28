<script lang="ts">
  import { BookOpen, CheckCircle2, Clock, History, PencilLine, XCircle } from "lucide-svelte";
  import { api } from "$lib/api";
  import { auth } from "$lib/auth.store";
  import StatusBadge from "$lib/components/ui/StatusBadge.svelte";
  import Dialog from "$lib/components/ui/Dialog.svelte";
  import { thesisStatusProps, titleChangeStatusProps } from "$lib/constants/statuses";
  import { formatDate } from "$lib/utils/format";
  import { apiErrorMessage } from "$lib/utils/errors";
  import Reveal from "$lib/components/landing/Reveal.svelte";

  const user = $derived($auth.user);

  let thesis = $state<any | null>(null);
  let requests = $state<any[]>([]);
  let loading = $state(true);
  let error = $state("");
  let requestsLoading = $state(false);
  let submitOpen = $state(false);
  let cancelOpen = $state(false);
  let cancelTarget = $state<any | null>(null);
  let actionError = $state<string | null>(null);
  let submitting = $state(false);
  let cancelling = $state(false);
  let cancelError = $state<string | null>(null);

  let newTitle = $state("");
  let newReason = $state("");
  let titleError = $state<string | null>(null);

  function errorMessage(err: unknown, fallback: string): string {
    return apiErrorMessage(err, fallback);
  }

  async function loadThesis() {
    loading = true;
    error = "";
    try {
      const res = await api.api.v1.theses.$get({ query: { page: "1", per_page: "1" } });
      if (!res.ok) throw await res.json();
      const json: any = await res.json();
      const list = json?.data ?? [];
      const own = user?.id ? list.find((t: any) => (t.student?.id ?? t.studentId) === user.id) ?? list[0] : list[0];
      thesis = own ?? null;
    } catch (e: any) {
      error = errorMessage(e, "Gagal memuat skripsi.");
    } finally {
      loading = false;
    }
  }

  async function loadRequests(id: string) {
    requestsLoading = true;
    try {
      const res = await api.api.v1.titleChangeRequests.$get({});
      if (!res.ok) throw await res.json();
      const json: any = await res.json();
      const all: any[] = json?.data ?? [];
      requests = all.filter((r: any) => (r.thesisId ?? r.thesis_id) === id);
    } catch {
      requests = [];
    } finally {
      requestsLoading = false;
    }
  }

  $effect(() => {
    if ($auth.accessToken) loadThesis();
  });

  $effect(() => {
    if (thesis?.id) loadRequests(thesis.id);
  });

  const sp = $derived(thesis ? thesisStatusProps(thesis.status ?? "") : null);
  const hasPending = $derived(requests.some((r) => r.status === "PENDING"));
  const canRequest = $derived(
    user?.role === "mahasiswa" &&
      thesis &&
      (thesis.status === "approved" || thesis.status === "in_progress") &&
      (thesis.supervisors?.length ?? 0) > 0 &&
      !hasPending,
  );

  function validateTitle(value: string): string | null {
    const v = value.trim();
    if (!v) return "Judul baru wajib diisi";
    if (v.length > 500) return "Judul maksimal 500 karakter";
    const words = v.split(/\s+/).filter(Boolean).length;
    if (words < 10) return "Judul minimal 10 kata";
    return null;
  }

  function openSubmit() {
    newTitle = "";
    newReason = "";
    titleError = null;
    actionError = null;
    submitOpen = true;
  }

  function closeSubmit() {
    submitOpen = false;
  }

  function openCancel(req: any) {
    cancelError = null;
    cancelTarget = req;
  }

  function closeCancel() {
    cancelTarget = null;
  }

  async function onSubmitChange() {
    if (!thesis?.id) return;
    const tErr = validateTitle(newTitle);
    if (tErr) {
      titleError = tErr;
      return;
    }
    titleError = null;
    actionError = null;
    submitting = true;
    try {
      const res = await api.api.v1.titleChangeRequests.$post({
        json: {
          thesisId: thesis.id,
          requestedTitle: newTitle.trim(),
          reason: newReason.trim() || undefined,
        },
      });
      if (!res.ok) throw await res.json().catch(() => ({}));
      closeSubmit();
      await loadRequests(thesis.id);
    } catch (e: any) {
      actionError = errorMessage(e, "Gagal mengajukan perubahan judul.");
    } finally {
      submitting = false;
    }
  }

  async function confirmCancel() {
    if (!cancelTarget?.id) return;
    cancelling = true;
    cancelError = null;
    try {
      const res = await api.api.v1.titleChangeRequests[cancelTarget.id].cancel.$patch({});
      if (!res.ok) throw await res.json().catch(() => ({}));
      closeCancel();
      if (thesis?.id) await loadRequests(thesis.id);
    } catch (e: any) {
      cancelError = errorMessage(e, "Gagal membatalkan permintaan.");
    } finally {
      cancelling = false;
    }
  }
</script>

<div class="space-y-6">
  <div class="flex flex-wrap items-center justify-between gap-3">
    <div>
      <p class="landing-eyebrow">Tugas Akhir Skripsi Saya</p>
      <h1 class="mt-2 text-balance landing-heading text-2xl">
        Skripsi <span class="accent-text italic">Saya</span>
      </h1>
    </div>
    {#if !loading && thesis}
      <div class="flex flex-wrap items-center gap-2">
        {#if canRequest}
          <button
            type="button"
            onclick={openSubmit}
            class="inline-flex h-9 items-center gap-2 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary-700"
          >
            <PencilLine size={16} /> Ajukan Perubahan Judul
          </button>
        {/if}
        <a
          href="/dashboard/thesis/new"
          class="inline-flex h-9 items-center justify-center rounded-md border border-st-stroke bg-st-surface px-3 py-2 text-sm font-medium text-st-text transition-colors hover:bg-st-surface-hi"
        >
          Ajukan Judul Baru
        </a>
      </div>
    {/if}
  </div>

  {#if loading}
    <div class="rounded-2xl border border-st-stroke bg-st-surface p-6 text-sm text-st-muted">
      Memuat skripsi…
    </div>
  {:else if error}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {error}
    </div>
  {:else if !thesis}
    <Reveal>
      <div class="mx-auto max-w-lg space-y-6 rounded-2xl border border-st-stroke bg-st-surface py-10 text-center">
        <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-primary-50 text-primary">
          <BookOpen size={32} />
        </div>
        <div>
          <h2 class="text-xl font-bold text-st-text">Anda belum memiliki skripsi</h2>
          <p class="mt-2 text-sm text-st-muted">
            Ajukan judul skripsi Anda untuk memulai perjalanan Tugas Akhir.
          </p>
        </div>
        <a
          href="/dashboard/thesis/new"
          class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary-700"
        >
          Ajukan Judul Skripsi
        </a>
      </div>
    </Reveal>
  {:else}
    <Reveal>
      <article class="rounded-2xl border border-st-stroke bg-st-surface p-6">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div class="space-y-1.5">
            <h2 class="text-xl leading-snug text-st-text">{thesis.title}</h2>
            {#if sp}
              <StatusBadge variant={sp.variant} label={sp.label} />
            {/if}
          </div>
        </div>

        <div class="mt-4 space-y-4">
          {#if thesis.abstract}
            <div>
              <p class="mb-1 text-sm font-medium text-st-muted">Abstrak</p>
              <p class="text-sm leading-relaxed text-st-text">{thesis.abstract}</p>
            </div>
          {/if}
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <p class="text-sm font-medium text-st-muted">Bidang Keahlian</p>
              <p class="text-sm text-st-text">{thesis.fieldOfStudy ?? thesis.field_of_study ?? "—"}</p>
            </div>
            <div>
              <p class="text-sm font-medium text-st-muted">Pembimbing</p>
              <p class="text-sm text-st-text">
                {thesis.supervisors?.length
                  ? thesis.supervisors.map((s: any) => s.fullName ?? s.full_name).join(", ")
                  : "Belum ditentukan"}
              </p>
            </div>
          </div>
          {#if thesis.kaprodiNotes ?? thesis.kaprodi_notes}
            <div class="rounded-lg border border-secondary/30 bg-secondary-50 p-4">
              <p class="text-sm font-medium text-secondary-foreground">Catatan Kaprodi</p>
              <p class="mt-1 text-sm text-st-text">{thesis.kaprodiNotes ?? thesis.kaprodi_notes}</p>
            </div>
          {/if}
        </div>
      </article>
    </Reveal>

    <Reveal delay={80}>
      <article class="rounded-2xl border border-st-stroke bg-st-surface p-6">
        <div class="flex items-center justify-between gap-2">
          <div class="space-y-1">
            <h3 class="flex items-center gap-2 text-base font-medium text-st-text">
              <History size={16} class="text-primary" /> Riwayat Perubahan Judul
            </h3>
            <p class="text-sm text-st-muted">
              Pengajuan dan hasil review perubahan judul skripsi Anda
            </p>
          </div>
        </div>

        <div class="mt-4">
          {#if requestsLoading}
            <p class="py-6 text-center text-sm text-st-muted">Memuat riwayat…</p>
          {:else if requests.length === 0}
            <p class="py-6 text-center text-sm text-st-muted">
              Belum ada pengajuan perubahan judul.
            </p>
          {:else}
            <div class="space-y-3">
              {#each requests as r (r.id)}
                {@const rb = titleChangeStatusProps(r.status)}
                <div class="rounded-xl border border-st-stroke p-4">
                  <div class="flex flex-wrap items-start justify-between gap-2">
                    <div class="min-w-0 flex-1">
                      <div class="flex flex-wrap items-center gap-2">
                        <p class="font-medium text-st-text">Judul baru: {r.requestedTitle ?? r.requested_title}</p>
                        <span class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-[11px] font-medium {rb.variant === 'pending' ? 'bg-warning-50 text-warning border border-warning/20' : rb.variant === 'approved' ? 'bg-success-50 text-success border border-success/20' : 'bg-danger-50 text-danger-700 border border-danger/20'}">
                          {#if r.status === "PENDING"}
                            <Clock size={12} />
                          {:else if r.status === "APPROVED"}
                            <CheckCircle2 size={12} />
                          {:else if r.status === "REJECTED"}
                            <XCircle size={12} />
                          {/if}
                          {rb.label}
                        </span>
                      </div>
                      <p class="mt-1 text-xs text-st-muted">
                        Diajukan {formatDate(r.createdAt ?? r.created_at)}
                        {#if r.requestedBy?.fullName ?? r.requested_by?.full_name}
                          oleh {r.requestedBy?.fullName ?? r.requested_by?.full_name}
                        {/if}
                      </p>
                    </div>
                    {#if r.status === "PENDING"}
                      <button
                        type="button"
                        onclick={() => openCancel(r)}
                        class="inline-flex h-8 items-center gap-1.5 rounded-md px-2 text-sm font-medium text-danger-700 transition-colors hover:bg-danger-50"
                      >
                        <XCircle size={14} /> Batalkan
                      </button>
                    {/if}
                  </div>

                  <div class="mt-3 grid grid-cols-1 gap-3 text-sm sm:grid-cols-2">
                    <div class="rounded-lg bg-st-bg p-3">
                      <p class="text-xs font-medium text-st-muted">Judul Sebelumnya</p>
                      <p class="mt-0.5 leading-snug text-st-text line-through decoration-st-muted/40">
                        {r.previousTitle ?? r.previous_title}
                      </p>
                    </div>
                    <div class="rounded-lg bg-primary-50 p-3">
                      <p class="text-xs font-medium text-primary-700">Judul yang Diajukan</p>
                      <p class="mt-0.5 leading-snug text-st-text">
                        {r.requestedTitle ?? r.requested_title}
                      </p>
                    </div>
                  </div>

                  {#if r.reason}
                    <p class="mt-3 text-sm text-st-text">
                      <span class="font-medium text-st-muted">Alasan: </span>{r.reason}
                    </p>
                  {/if}

                  {#if ((r.reviewedBy?.fullName ?? r.reviewed_by?.full_name) || r.reviewNotes) || r.review_notes}
                    <div class="mt-3 border-t border-st-stroke pt-3 text-sm">
                      {#if r.reviewedBy?.fullName ?? r.reviewed_by?.full_name}
                        <p class="text-st-muted">
                          <span class="font-medium text-st-text">Direview oleh:</span>
                          {r.reviewedBy?.fullName ?? r.reviewed_by?.full_name}
                        </p>
                      {/if}
                      {#if r.reviewNotes ?? r.review_notes}
                        <p class="mt-1 text-st-muted">
                          <span class="font-medium text-st-text">Catatan Pembimbing:</span>
                          {r.reviewNotes ?? r.review_notes}
                        </p>
                      {/if}
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </article>
    </Reveal>
  {/if}
</div>

<Dialog bind:open={submitOpen} labelledBy="submit-title" initialFocus="#requested-title">
  <h2 id="submit-title" class="text-lg font-semibold text-st-text">Ajukan Perubahan Judul</h2>
  <p class="mt-1 text-sm text-st-muted">
    Perubahan judul akan diproses oleh Dosen Pembimbing Anda.
  </p>

  <form
    class="mt-4 space-y-4"
    onsubmit={(e) => {
      e.preventDefault();
      onSubmitChange();
    }}
  >
    {#if actionError}
      <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
        {actionError}
      </div>
    {/if}
    <div>
      <label for="current-title" class="text-sm font-medium text-st-text">Judul Saat Ini</label>
      <input
        id="current-title"
        type="text"
        readonly
        value={thesis?.title ?? ""}
        class="mt-1 w-full rounded-md border border-st-stroke bg-st-bg px-3 py-2 text-sm text-st-muted"
      />
    </div>
    <div>
      <label for="requested-title" class="text-sm font-medium text-st-text">Judul Baru <span class="text-danger-700">*</span></label>
      <textarea
        id="requested-title"
        rows="3"
        placeholder="Tulis judul baru skripsi Anda (minimal 10 kata)"
        bind:value={newTitle}
        class="mt-1 w-full rounded-md border {titleError ? 'border-danger-700' : 'border-st-stroke'} bg-st-surface px-3 py-2 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
      ></textarea>
      {#if titleError}
        <p class="mt-1 text-xs text-danger-700">{titleError}</p>
      {/if}
    </div>
    <div>
      <label for="reason" class="text-sm font-medium text-st-text">Alasan Perubahan</label>
      <textarea
        id="reason"
        rows="2"
        placeholder="Alasan mengajukan perubahan judul (opsional)"
        bind:value={newReason}
        class="mt-1 w-full rounded-md border border-st-stroke bg-st-surface px-3 py-2 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
      ></textarea>
    </div>
    <div class="flex flex-wrap justify-end gap-2 pt-1">
      <button
        type="button"
        onclick={closeSubmit}
        class="inline-flex h-9 items-center rounded-md px-3 text-sm font-medium text-st-text transition-colors hover:bg-st-surface-hi"
      >
        Batal
      </button>
      <button
        type="submit"
        disabled={submitting}
        class="inline-flex h-9 items-center gap-2 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary-700 disabled:opacity-50"
      >
        <PencilLine size={16} /> {submitting ? "Mengajukan…" : "Ajukan"}
      </button>
    </div>
  </form>
</Dialog>

<Dialog bind:open={cancelOpenDerived} labelledBy="cancel-title" initialFocus="button:not([disabled])">
  {#snippet _unused()}
  {/snippet}
</Dialog>
