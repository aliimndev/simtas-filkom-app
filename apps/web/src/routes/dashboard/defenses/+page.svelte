<script lang="ts">
  import { ClipboardCheck, CalendarDays, MapPin, Star } from "lucide-svelte";
  import { api } from "$lib/api";
  import StatusBadge from "$lib/components/dashboard/StatusBadge.svelte";
  import type { StatusVariant } from "$lib/components/dashboard/thesis-status";

  let status = $state("");
  let list = $state<any[]>([]);
  let loading = $state(true);
  let error = $state("");

  const filters: { value: string; label: string }[] = [
    { value: "", label: "Semua" },
    { value: "submitted", label: "submitted" },
    { value: "scheduled", label: "scheduled" },
    { value: "in_revision", label: "in revision" },
    { value: "passed", label: "passed" },
    { value: "failed", label: "failed" },
  ];

  function defenseStatusProps(s: string): { variant: StatusVariant; label: string } {
    switch (s) {
      case "passed": return { variant: "completed", label: "Lulus" };
      case "failed": return { variant: "rejected", label: "Tidak Lulus" };
      case "in_revision": return { variant: "pending", label: "Revisi" };
      case "scheduled": return { variant: "in_progress", label: "Terjadwal" };
      case "submitted": return { variant: "pending", label: "Diajukan" };
      default: return { variant: "draft", label: s ? s.replace(/_/g, " ") : "Diajukan" };
    }
  }

  async function load() {
    loading = true;
    error = "";
    try {
      const query: Record<string, string> = { per_page: "50" };
      if (status) query.status = status;
      const res = await api.api.v1.theses.$get({ query });
      const json: any = res.ok ? await res.json() : null;
      list = (json?.data ?? []).filter((t: any) => t.status === "defense_ready" || status);
    } catch (e) {
      error = "Gagal memuat daftar sidang.";
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    status;
    load();
  });

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
</script>

<div class="space-y-6">
  <div>
    <p class="landing-eyebrow">Sidang Tugas Akhir</p>
    <h1 class="mt-2 text-balance landing-heading text-2xl">
      Jadwal dan hasil <span class="accent-text italic">sidang</span>
    </h1>
    <p class="mt-1.5 text-sm text-st-muted">Jadwal dan hasil sidang tugas akhir</p>
  </div>

  <div class="flex flex-wrap gap-2">
    {#each filters as f (f.value)}
      <button
        type="button"
        onclick={() => (status = f.value)}
        class="inline-flex h-8 items-center rounded-md px-3 text-sm font-medium {status === f.value
          ? 'bg-primary text-primary-foreground hover:bg-primary-700'
          : 'border border-st-stroke bg-st-surface text-st-text hover:bg-st-surface-hi'}"
      >
        {f.label}
      </button>
    {/each}
  </div>

  {#if error}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {error}
    </div>
  {/if}

  {#if loading}
    <div class="py-10 text-center text-sm text-st-muted">Memuat sidang…</div>
  {:else}
    <div class="space-y-3">
      {#each list as d (d.id)}
        {@const sp = defenseStatusProps(d.defenseStatus ?? d.status ?? "")}
        <article class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-st-stroke bg-st-surface p-5">
          <div class="flex min-w-0 items-center gap-3">
            <div class="accent-ring flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-success-50 text-success">
              <ClipboardCheck size={20} />
            </div>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <p class="font-medium text-st-text">{d.student?.fullName ?? d.student?.full_name ?? "Mahasiswa"}</p>
                <StatusBadge variant={sp.variant} label={sp.label} />
              </div>
              <p class="mt-0.5 truncate text-sm text-st-muted">{d.title ?? d.thesisTitle ?? "—"}</p>
              <div class="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-st-muted">
                {#if d.scheduledAt ?? d.scheduled_at}
                  <span class="inline-flex items-center gap-1">
                    <CalendarDays size={14} /> {formatDateTime(d.scheduledAt ?? d.scheduled_at)}
                  </span>
                {/if}
                {#if d.room}
                  <span class="inline-flex items-center gap-1">
                    <MapPin size={14} /> {d.room}
                  </span>
                {/if}
                {#if (d.averageScore ?? d.average_score) != null}
                  <span class="inline-flex items-center gap-1 text-warning">
                    <Star size={14} /> {Number(d.averageScore ?? d.average_score).toFixed(2)}
                  </span>
                {/if}
              </div>
            </div>
          </div>
          {#if d.examiners && d.examiners.length > 0}
            <div class="hidden text-right text-xs text-st-muted md:block">
              <p class="font-medium text-st-text">Penguji</p>
              {#each d.examiners as e (e.id)}
                <p>{e.fullName ?? e.full_name}</p>
              {/each}
            </div>
          {/if}
        </article>
      {/each}
      {#if list.length === 0}
        <div class="rounded-2xl border border-st-stroke bg-st-surface py-12 text-center">
          <ClipboardCheck size={36} class="mx-auto text-st-muted" />
          <p class="mt-3 landing-heading text-lg">Tidak ada <span class="accent-text italic">sidang</span></p>
          <p class="mt-1 text-sm text-st-muted">Jadwal sidang akan muncul di sini.</p>
        </div>
      {/if}
    </div>
  {/if}
</div>
