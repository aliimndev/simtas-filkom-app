<script lang="ts">
  import { CalendarDays, GraduationCap, MapPin, Star } from "lucide-svelte";
  import { api } from "$lib/api";
  import { auth } from "$lib/auth.store";
  import StatCard from "$lib/components/dashboard/StatCard.svelte";
  import StatusBadge from "$lib/components/dashboard/StatusBadge.svelte";
  import { thesisStatusProps } from "$lib/components/dashboard/thesis-status";
  import Reveal from "$lib/components/landing/Reveal.svelte";

  let status = $state("");
  let page = $state(1);
  let list = $state<any[]>([]);
  let total = $state(0);
  let loading = $state(true);
  let error = $state("");

  const STATUSES = [
    { value: "", label: "Semua" },
    { value: "pending", label: "Diajukan" },
    { value: "scheduled", label: "Terjadwal" },
    { value: "passed", label: "Lulus" },
    { value: "failed", label: "Tidak Lulus" },
  ];

  async function load() {
    loading = true;
    error = "";
    try {
      const query: Record<string, string> = { page: String(page), per_page: "10" };
      if (status) query.status = status;
      const res = await api.api.v1.seminars.$get({ query });
      const json: any = res.ok ? await res.json() : null;
      list = json?.data ?? [];
      total = json?.meta?.total ?? 0;
    } catch (e) {
      error = "Gagal memuat daftar seminar.";
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    status; page;
    load();
  });

  function seminarStatusProps(s: string): { variant: "pending" | "approved" | "rejected" | "draft" | "in_progress" | "completed"; label: string } {
    switch (s) {
      case "passed": return { variant: "completed", label: "Lulus" };
      case "failed": return { variant: "rejected", label: "Tidak Lulus" };
      case "scheduled": return { variant: "in_progress", label: "Terjadwal" };
      case "pending": return { variant: "pending", label: "Diajukan" };
      default: return { variant: "draft", label: s ? s.replace(/_/g, " ") : "—" };
    }
  }

  function formatDateTime(s?: string) {
    if (!s) return "—";
    try {
      return new Date(s).toLocaleString("id-ID", {
        day: "2-digit",
        month: "short",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
    } catch {
      return s;
    }
  }

  let stats = $derived(() => {
    const total_ = list.length;
    const pending = list.filter((s) => s.status === "pending").length;
    const scheduled = list.filter((s) => s.status === "scheduled").length;
    const passed = list.filter((s) => s.status === "passed").length;
    return { total: total_, pending, scheduled, passed };
  });
</script>

<div class="mx-auto w-full max-w-350 space-y-6">
  <Reveal>
    <div>
      <p class="landing-eyebrow">Seminar Proposal</p>
      <h1 class="mt-2 text-balance landing-heading text-2xl">
        Jadwal dan hasil <span class="accent-text italic">seminar</span>
      </h1>
      <p class="mt-1.5 text-sm text-st-muted">Jadwal dan hasil seminar proposal / kemajuan</p>
    </div>
  </Reveal>

  <Reveal delay={80}>
    <div class="grid grid-cols-2 gap-3 md:grid-cols-4">
      <StatCard title="Total" value={stats().total} icon={GraduationCap} tone="primary" />
      <StatCard title="Terjadwal" value={stats().scheduled} icon={CalendarDays} tone="secondary" />
      <StatCard title="Lulus" value={stats().passed} icon={Star} tone="success" />
      <StatCard title="Diajukan" value={stats().pending} icon={MapPin} tone="warning" />
    </div>
  </Reveal>

  <Reveal delay={120}>
    <div class="flex flex-wrap gap-2">
      {#each STATUSES as s (s.value)}
        <button
          type="button"
          onclick={() => { status = s.value; page = 1; }}
          class="inline-flex h-8 items-center rounded-md px-3 text-sm font-medium transition {status === s.value ? 'bg-primary text-primary-foreground hover:bg-primary-700' : 'border border-st-stroke bg-st-surface text-st-text hover:bg-st-surface-hi'}"
        >
          {s.label}
        </button>
      {/each}
    </div>
  </Reveal>

  {#if error}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {error}
    </div>
  {/if}

  {#if loading}
    <div class="py-10 text-center text-sm text-st-muted">Memuat seminar…</div>
  {:else}
    <div class="space-y-3">
      {#each list as sem (sem.id)}
        {@const sp = seminarStatusProps(sem.status ?? "")}
        {@const tp = thesisStatusProps(sem.thesisStatus ?? sem.thesis_status ?? "")}
        <article class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-st-stroke bg-st-surface p-5">
          <div class="flex min-w-0 items-center gap-3">
            <div class="accent-ring flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-(--st-accent-from)/10 text-(--st-accent-to)">
              <GraduationCap size={20} />
            </div>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <p class="font-medium text-st-text">{sem.student?.fullName ?? sem.student?.full_name ?? "Mahasiswa"}</p>
                <StatusBadge variant={sp.variant} label={sp.label} />
                {#if sem.stage}
                  <StatusBadge variant="draft" label={"Seminar " + sem.stage} />
                {/if}
                {#if sem.thesisStatus || sem.thesis_status}
                  <StatusBadge variant={tp.variant} label={tp.label} />
                {/if}
              </div>
              <p class="mt-0.5 truncate text-sm text-st-muted">{sem.thesisTitle ?? sem.thesis_title ?? sem.title ?? "—"}</p>
              <div class="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-st-muted">
                {#if sem.scheduledAt ?? sem.scheduled_at}
                  <span class="inline-flex items-center gap-1">
                    <CalendarDays size={14} /> {formatDateTime(sem.scheduledAt ?? sem.scheduled_at)}
                  </span>
                {/if}
                {#if sem.room}
                  <span class="inline-flex items-center gap-1">
                    <MapPin size={14} /> {sem.room}
                  </span>
                {/if}
                {#if sem.averageScore ?? sem.average_score}
                  {@const score = Number(sem.averageScore ?? sem.average_score)}
                  <span class="inline-flex items-center gap-1 text-warning">
                    <Star size={14} /> {score.toFixed(2)}
                  </span>
                {/if}
              </div>
            </div>
          </div>
          {#if sem.examiners && sem.examiners.length > 0}
            <div class="hidden text-right text-xs text-st-muted md:block">
              <p class="font-medium text-st-text">Penguji</p>
              {#each sem.examiners as e (e.id)}
                <p>{e.fullName ?? e.full_name}</p>
              {/each}
            </div>
          {/if}
        </article>
      {/each}
      {#if list.length === 0}
        <div class="rounded-2xl border border-st-stroke bg-st-surface py-12 text-center">
          <GraduationCap size={36} class="mx-auto text-st-muted" />
          <p class="mt-3 landing-heading text-lg">Tidak ada <span class="accent-text italic">seminar</span></p>
          <p class="mt-1 text-sm text-st-muted">Jadwal seminar akan muncul di sini.</p>
        </div>
      {/if}
    </div>
  {/if}

  {#if total > 10}
    <div class="flex items-center justify-between pt-2">
      <p class="text-sm text-st-muted">Total {total} seminar · Halaman {page}</p>
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
