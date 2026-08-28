<script lang="ts">
  import { MessageSquarePlus, Check, Calendar, User } from "lucide-svelte";
  import { api } from "$lib/api";
  import { auth } from "$lib/auth.store";
  import StatCard from "$lib/components/dashboard/StatCard.svelte";
  import StatusBadge from "$lib/components/dashboard/StatusBadge.svelte";
  import Reveal from "$lib/components/landing/Reveal.svelte";
  import { thesisStatusProps, type StatusVariant } from "$lib/components/dashboard/thesis-status";

  let thesisId = $state("");
  let status = $state("");
  let page = $state(1);
  let list = $state<any[]>([]);
  let total = $state(0);
  let summary = $state<{ total?: number; approved?: number; pending?: number } | null>(null);
  let loading = $state(true);
  let error = $state("");
  let actionError = $state("");

  const isStudent = $derived(($auth.user as any)?.role === "mahasiswa");

  function variantFor(s: string): StatusVariant {
    const v = thesisStatusProps(s).variant;
    return v;
  }
  function labelFor(s: string): string {
    return thesisStatusProps(s).label;
  }

  async function load() {
    loading = true;
    error = "";
    try {
      const query: Record<string, string> = { page: String(page), per_page: "20" };
      if (thesisId) query.thesisId = thesisId;
      if (status) query.status = status;
      const res = await api.api.v1.consultationLogs.$get({ query });
      const json: any = res.ok ? await res.json() : null;
      list = json?.data ?? [];
      total = json?.meta?.total ?? list.length;
      summary = json?.summary ?? null;

      if (!thesisId && isStudent) {
        const me: any = ($auth.user as any) ?? {};
        if (me.thesisId) thesisId = me.thesisId;
      }
      if (thesisId) {
        const sumRes = await api.api.v1.consultationLogs.summary.$get({ query: { thesisId } });
        const sumJson: any = sumRes.ok ? await sumRes.json() : null;
        if (sumJson?.data) summary = sumJson.data;
      }
    } catch (e) {
      error = "Gagal memuat catatan bimbingan.";
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    thesisId; status; page;
    load();
  });

  async function approve(id: string) {
    actionError = "";
    try {
      if (!thesisId) {
        actionError = "thesisId belum dipilih.";
        return;
      }
      const res = await api.api.v1.consultationLogs[":id"].approve.$patch({
        param: { id },
        query: { thesisId },
      });
      if (!res.ok) {
        const body = await res.json().catch(() => ({}));
        actionError = (body as any)?.error?.message ?? "Gagal menyetujui bimbingan.";
        return;
      }
      await load();
    } catch (e) {
      actionError = "Gagal menyetujui bimbingan.";
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
</script>

<div class="space-y-6">
  <Reveal>
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <p class="landing-eyebrow">Catatan Bimbingan</p>
        <h1 class="mt-2 text-balance landing-heading text-2xl">
          {#if isStudent}
            Bimbingan <span class="accent-text italic">Anda</span>
          {:else}
            Bimbingan <span class="accent-text italic">Mahasiswa</span>
          {/if}
        </h1>
        <p class="mt-1.5 text-sm text-st-muted">
          {isStudent
            ? "Catatan bimbingan dengan dosen pembimbing Anda"
            : "Catatan bimbingan mahasiswa yang Anda bimbing"}
        </p>
      </div>
    </div>
  </Reveal>

  {#if summary}
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
      <StatCard title="Total" value={summary.total ?? list.length} icon={MessageSquarePlus} tone="primary" />
      <StatCard title="Disetujui" value={summary.approved ?? 0} icon={Check} tone="success" />
      <StatCard title="Menunggu" value={summary.pending ?? 0} icon={Calendar} tone="warning" />
    </div>
  {/if}

  <div class="flex flex-wrap items-center gap-3">
    {#if !isStudent}
      <input
        type="search"
        placeholder="Thesis ID (UUID)…"
        bind:value={thesisId}
        oninput={() => (page = 1)}
        class="w-full max-w-xs rounded-md border border-st-stroke bg-st-surface px-3 py-2 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
      />
    {/if}
    <select
      bind:value={status}
      onchange={() => (page = 1)}
      class="w-44 rounded-md border border-st-stroke bg-st-surface px-3 py-2 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
    >
      <option value="">Semua status</option>
      <option value="pending">Menunggu</option>
      <option value="approved">Disetujui</option>
      <option value="rejected">Ditolak</option>
    </select>
  </div>

  {#if actionError}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {actionError}
    </div>
  {/if}
  {#if error}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {error}
    </div>
  {/if}

  {#if loading}
    <div class="py-10 text-center text-sm text-st-muted">Memuat catatan bimbingan…</div>
  {:else}
    <div class="space-y-3">
      {#each list as c (c.id)}
        {@const variant = variantFor(c.status ?? "")}
        {@const label = labelFor(c.status ?? "")}
        {@const date = c.consultationDate ?? c.consultation_date ?? c.date}
        {@const studentName = c.studentName ?? c.student_name ?? c.student?.fullName ?? c.student?.full_name}
        <article class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-st-stroke bg-st-surface p-5">
          <div class="flex min-w-0 items-center gap-3">
            <div class="accent-ring flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-(--st-accent-from)/10 text-(--st-accent-to)">
              <MessageSquarePlus size={20} />
            </div>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <p class="font-medium text-st-text">
                  {c.thesisTitle ?? c.thesis_title ?? "Bimbingan"}
                </p>
                <StatusBadge {variant} {label} />
              </div>
              <p class="mt-0.5 line-clamp-1 text-sm text-st-muted">
                {c.topicsDiscussed ?? c.topics_discussed ?? c.topic ?? "—"}
              </p>
              <p class="mt-1 flex flex-wrap items-center gap-3 text-xs text-st-muted">
                <span class="inline-flex items-center gap-1">
                  <Calendar size={12} aria-hidden="true" /> {formatDate(date)}
                </span>
                {#if studentName}
                  <span class="inline-flex items-center gap-1">
                    <User size={12} aria-hidden="true" /> {studentName}
                  </span>
                {/if}
              </p>
            </div>
          </div>
          {#if !isStudent && (c.status === "pending" || !c.status)}
            <button
              type="button"
              onclick={() => approve(c.id)}
              class="inline-flex h-8 items-center gap-1.5 rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground hover:bg-primary-700"
            >
              <Check size={14} aria-hidden="true" /> Setujui
            </button>
          {/if}
        </article>
      {/each}
      {#if list.length === 0}
        <div class="rounded-2xl border border-st-stroke bg-st-surface py-12 text-center">
          <MessageSquarePlus size={36} class="mx-auto text-st-muted" />
          <p class="mt-3 landing-heading text-lg">
            Belum ada <span class="accent-text italic">catatan bimbingan</span>
          </p>
          <p class="mt-1 text-sm text-st-muted">Mulai dengan mencatat sesi bimbingan pertama Anda.</p>
        </div>
      {/if}
    </div>
  {/if}

  {#if total > 20}
    <div class="flex items-center justify-between pt-2">
      <p class="text-sm text-st-muted">Total {total} catatan · Halaman {page}</p>
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
          disabled={page >= Math.ceil(total / 20)}
          onclick={() => (page += 1)}
          class="inline-flex h-8 items-center rounded-md border border-st-stroke bg-st-surface px-3 text-sm text-st-text hover:bg-st-surface-hi disabled:opacity-50"
        >
          Berikutnya
        </button>
      </div>
    </div>
  {/if}
</div>
