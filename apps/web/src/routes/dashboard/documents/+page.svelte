<script lang="ts">
  import { FileText, Download, CheckCircle2, RefreshCcw, Eye, Loader2 } from "lucide-svelte";
  import { api } from "$lib/api";
  import { auth } from "$lib/auth.store";
  import StatusBadge from "$lib/components/dashboard/StatusBadge.svelte";
  import Reveal from "$lib/components/landing/Reveal.svelte";
  import type { StatusVariant } from "$lib/components/dashboard/thesis-status";

  const DOCUMENT_TYPES: { value: string; label: string }[] = [
    { value: "proposal", label: "Proposal" },
    { value: "draft_chapter", label: "Draft Bab" },
    { value: "seminar_doc", label: "Dokumen Seminar" },
    { value: "defense_doc", label: "Dokumen Sidang" },
    { value: "final_thesis", label: "Skripsi Final" },
    { value: "revision_sheet", label: "Lembar Revisi" },
    { value: "endorsement_letter", label: "Surat Pengesahan" },
  ];

  const TYPE_LABEL: Record<string, string> = Object.fromEntries(
    DOCUMENT_TYPES.map((t) => [t.value, t.label]),
  );

  const STATUS_LABEL: Record<string, string> = {
    pending_review: "Menunggu Review",
    approved: "Disetujui",
    revision_required: "Perlu Revisi",
  };

  const STATUS_VARIANT: Record<string, StatusVariant> = {
    pending_review: "pending",
    approved: "approved",
    revision_required: "rejected",
  };

  const isStudent = $derived(($auth.user as any)?.role === "mahasiswa");

  let theses = $state<any[]>([]);
  let thesesLoading = $state(true);
  let selectedThesisId = $state<string | null>(null);

  let list = $state<any[]>([]);
  let loading = $state(true);
  let error = $state("");
  let actionError = $state("");
  let downloadingId = $state<string | null>(null);

  let selectedType = $state("");

  const activeThesis = $derived(
    isStudent ? theses[0] : theses.find((t) => t.id === selectedThesisId) ?? theses[0],
  );
  const activeThesisId = $derived(activeThesis?.id as string | undefined);

  async function loadTheses() {
    thesesLoading = true;
    try {
      const res = await api.api.v1.theses.$get({ query: { per_page: "100" } });
      const json: any = res.ok ? await res.json() : null;
      theses = json?.data ?? json ?? [];
    } catch {
      theses = [];
    } finally {
      thesesLoading = false;
    }
  }

  async function loadDocuments() {
    if (!activeThesisId) {
      list = [];
      loading = false;
      return;
    }
    loading = true;
    error = "";
    try {
      const query: Record<string, string> = { thesisId: activeThesisId, per_page: "50" };
      if (selectedType) query.document_type = selectedType;
      const res = await api.api.v1.documents.$get({ query });
      const json: any = res.ok ? await res.json() : null;
      list = json?.data ?? [];
    } catch {
      error = "Gagal memuat daftar dokumen.";
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    loadTheses();
  });

  $effect(() => {
    activeThesisId; selectedType;
    loadDocuments();
  });

  async function downloadDoc(id: string) {
    if (!activeThesisId) return;
    downloadingId = id;
    actionError = "";
    try {
      const res = await api.api.v1.documents[":id"].download.$get({
        param: { id },
        query: { thesisId: activeThesisId },
      });
      const json: any = res.ok ? await res.json() : null;
      const url = json?.data?.url ?? json?.url;
      if (url) {
        window.open(url, "_blank", "noopener");
      } else {
        actionError = "URL unduhan tidak tersedia.";
      }
    } catch {
      actionError = "Gagal mengunduh dokumen.";
    } finally {
      downloadingId = null;
    }
  }

  async function reviewDoc(id: string, decision: "approved" | "revision_required") {
    actionError = "";
    try {
      const res = await api.api.v1.documents[":id"].review.$patch({
        param: { id },
        json: { decision, notes: decision === "revision_required" ? "Silakan perbaiki sesuai catatan." : undefined },
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        actionError = (body as any)?.error?.message ?? "Gagal memperbarui status dokumen.";
        return;
      }
      await loadDocuments();
    } catch {
      actionError = "Gagal memperbarui status dokumen.";
    }
  }

  function statusBadge(status?: string) {
    const key = status ?? "pending_review";
    return {
      variant: STATUS_VARIANT[key] ?? "draft",
      label: STATUS_LABEL[key] ?? key.replace(/_/g, " "),
    };
  }

  function formatDate(s?: string) {
    if (!s) return "—";
    try {
      return new Date(s).toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
    } catch {
      return s;
    }
  }

  function formatSize(bytes?: number | null) {
    if (bytes == null) return null;
    return `${(bytes / 1024 / 1024).toFixed(2)} MB`;
  }
</script>

<div class="space-y-6">
  <Reveal>
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <p class="landing-eyebrow">Dokumen Skripsi</p>
        <h1 class="mt-2 text-balance landing-heading text-2xl">
          {#if isStudent}
            Unggah dan Pantau <span class="accent-text italic">Dokumen</span>
          {:else}
            Review Dokumen <span class="accent-text italic">Mahasiswa</span>
          {/if}
        </h1>
        <p class="mt-1.5 text-sm text-st-muted">
          {isStudent
            ? "Unggah dan pantau dokumen skripsi Anda"
            : "Review dokumen yang diunggah mahasiswa bimbingan"}
        </p>
        {#if !isStudent}
          {#if thesesLoading}
            <p class="mt-2 text-xs text-st-muted">Memuat daftar skripsi…</p>
          {:else if theses.length > 1}
            <div class="mt-3">
              <label for="thesis-picker" class="sr-only">Pilih skripsi mahasiswa</label>
              <select
                id="thesis-picker"
                value={activeThesisId ?? ""}
                onchange={(e) => (selectedThesisId = (e.currentTarget as HTMLSelectElement).value || null)}
                class="max-w-xs rounded-md border border-st-stroke bg-st-surface px-3 py-2 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
              >
                {#each theses as t (t.id)}
                  <option value={t.id}>
                    {(t.student?.fullName ?? t.student?.full_name ?? "—") + " — " + (t.title ?? "").slice(0, 40)}{(t.title?.length ?? 0) > 40 ? "…" : ""}
                  </option>
                {/each}
              </select>
            </div>
          {/if}
        {/if}
      </div>
    </div>
  </Reveal>

  {#if error}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {error}
    </div>
  {/if}
  {#if actionError}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {actionError}
    </div>
  {/if}

  {#if thesesLoading}
    <div class="py-10 text-center text-sm text-st-muted">Memuat dokumen…</div>
  {:else if !activeThesisId}
    <div class="rounded-2xl border border-st-stroke bg-st-surface py-14 text-center">
      <FileText size={40} class="mx-auto text-st-muted" aria-hidden="true" />
      <p class="mt-4 font-semibold text-st-text">Belum ada skripsi aktif</p>
      <p class="mt-1 text-sm text-st-muted">
        {isStudent
          ? "Ajukan skripsi terlebih dahulu untuk mengelola dokumen."
          : "Pilih skripsi mahasiswa untuk melihat dokumennya."}
      </p>
    </div>
  {:else}
    <div class="flex flex-wrap gap-2">
      <button
        type="button"
        onclick={() => (selectedType = "")}
        class="inline-flex h-8 items-center rounded-md border px-3 text-sm font-medium transition-colors {selectedType === '' ? 'border-primary bg-primary text-primary-foreground' : 'border-st-stroke bg-st-surface text-st-text hover:bg-st-surface-hi'}"
      >
        Semua
      </button>
      {#each DOCUMENT_TYPES as t (t.value)}
        <button
          type="button"
          onclick={() => (selectedType = t.value)}
          class="inline-flex h-8 items-center rounded-md border px-3 text-sm font-medium transition-colors {selectedType === t.value ? 'border-primary bg-primary text-primary-foreground' : 'border-st-stroke bg-st-surface text-st-text hover:bg-st-surface-hi'}"
        >
          {t.label}
        </button>
      {/each}
    </div>

    {#if loading}
      <div class="py-10 text-center text-sm text-st-muted">Memuat dokumen…</div>
    {:else}
      <div class="space-y-3">
        {#each list as doc (doc.id)}
          {@const sb = statusBadge(doc.status)}
          {@const typeLabel = TYPE_LABEL[doc.document_type] ?? doc.document_type}
          {@const size = formatSize(doc.file_size ?? doc.fileSize)}
          <article class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-st-stroke bg-st-surface p-5">
            <div class="flex min-w-0 items-center gap-3">
              <div class="accent-ring flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-(--st-accent-from)/10 text-(--st-accent-to)">
                <FileText size={20} aria-hidden="true" />
              </div>
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <p class="truncate font-medium text-st-text">{doc.file_name ?? doc.fileName ?? "—"}</p>
                  <StatusBadge variant={sb.variant} label={sb.label} />
                </div>
                <p class="mt-0.5 text-xs text-st-muted">
                  {typeLabel}{doc.chapter_number ? ` · Bab ${doc.chapter_number}` : ""}{doc.version != null ? ` · v${doc.version}` : ""}{size ? ` · ${size}` : ""} · {formatDate(doc.created_at ?? doc.createdAt)}
                </p>
                {#if doc.reviewer_notes && doc.status === "revision_required"}
                  <p class="mt-1 text-xs text-danger-700">Catatan revisi: {doc.reviewer_notes}</p>
                {/if}
              </div>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <button
                type="button"
                disabled={downloadingId === doc.id}
                onclick={() => downloadDoc(doc.id)}
                class="inline-flex h-8 items-center gap-1.5 rounded-md border border-st-stroke bg-st-surface px-3 text-sm font-medium text-st-text hover:bg-st-surface-hi disabled:opacity-50"
              >
                {#if downloadingId === doc.id}
                  <Loader2 size={14} aria-hidden="true" class="animate-spin" />
                {:else}
                  <Download size={14} aria-hidden="true" />
                {/if}
                Unduh
              </button>
              {#if !isStudent && doc.status === "pending_review"}
                <button
                  type="button"
                  onclick={() => reviewDoc(doc.id, "approved")}
                  class="inline-flex h-8 items-center gap-1.5 rounded-md bg-success px-3 text-sm font-medium text-white hover:bg-success/90"
                >
                  <CheckCircle2 size={14} aria-hidden="true" /> Setujui
                </button>
                <button
                  type="button"
                  onclick={() => reviewDoc(doc.id, "revision_required")}
                  class="inline-flex h-8 items-center gap-1.5 rounded-md border border-st-stroke bg-st-surface px-3 text-sm font-medium text-st-text hover:bg-st-surface-hi"
                >
                  <RefreshCcw size={14} aria-hidden="true" /> Minta Revisi
                </button>
              {:else if !isStudent}
                <button
                  type="button"
                  disabled={downloadingId === doc.id}
                  onclick={() => downloadDoc(doc.id)}
                  class="inline-flex h-8 items-center gap-1.5 rounded-md border border-st-stroke bg-st-surface px-3 text-sm font-medium text-st-text hover:bg-st-surface-hi disabled:opacity-50"
                >
                  <Eye size={14} aria-hidden="true" /> Lihat
                </button>
              {/if}
            </div>
          </article>
        {/each}
        {#if list.length === 0}
          <div class="rounded-2xl border border-st-stroke bg-st-surface py-12 text-center">
            <FileText size={36} class="mx-auto text-st-muted" aria-hidden="true" />
            <p class="mt-3 landing-heading text-lg">Belum ada <span class="accent-text italic">dokumen</span></p>
            <p class="mt-1 text-sm text-st-muted">
              {isStudent
                ? "Unggah dokumen skripsi pertama Anda untuk memulai."
                : "Mahasiswa belum mengunggah dokumen untuk skripsi ini."}
            </p>
          </div>
        {/if}
      </div>
    {/if}
  {/if}
</div>
