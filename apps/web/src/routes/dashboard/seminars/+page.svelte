<script lang="ts">
  import { CalendarDays, GraduationCap, MapPin, Star } from "lucide-svelte";
  import { api } from "$lib/api";
  import { auth } from "$lib/auth.store";
  import StatCard from "$lib/components/ui/StatCard.svelte";
  import StatusBadge from "$lib/components/ui/StatusBadge.svelte";
  import Pagination from "$lib/components/ui/Pagination.svelte";
  import { thesisStatusProps, seminarStatusProps } from "$lib/constants/statuses";
  import { formatDateTime, toDatetimeLocalValue, minDatetimeLocalValue } from "$lib/utils/format";
  import { readError } from "$lib/utils/errors";
  import Reveal from "$lib/components/landing/Reveal.svelte";

  const PAGE_SIZE = 10;

  let status = $state("");
  let page = $state(1);
  let list = $state<any[]>([]);
  let total = $state(0);
  let loading = $state(true);
  let error = $state("");
  let schedulingId = $state("");
  let schedulingAt = $state("");
  let schedulingRoom = $state("");
  let schedulingExaminerIds = $state<string[]>([]);
  let schedulingLoading = $state(false);
  let schedulingError = $state("");
  let schedulingSuccess = $state("");
  let examinerOptions = $state<any[]>([]);
  let canSchedule = $derived(["KAPRODI", "ADMIN_FAKULTAS"].includes(String($auth.user?.role ?? "").toUpperCase()));

  const MIN_LEAD_MS = 3 * 24 * 60 * 60 * 1000;

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
      const query: Record<string, string> = { page: String(page), per_page: String(PAGE_SIZE) };
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

  async function loadExaminers() {
    try {
      const res = await api.api.v1.users.$get({ query: { role: "dosen_penguji", is_active: "true", per_page: "100" } });
      const json: any = res.ok ? await res.json() : null;
      examinerOptions = json?.data ?? [];
    } catch {
      examinerOptions = [];
    }
  }

  function openScheduling(sem: any) {
    schedulingId = sem.id;
    schedulingAt = toDatetimeLocalValue(sem.scheduledAt ?? sem.scheduled_at);
    schedulingRoom = sem.room ?? "";
    schedulingExaminerIds = (sem.examiners ?? []).map((examiner: any) => examiner.id);
    schedulingError = "";
    schedulingSuccess = "";
  }

  function toggleExaminer(id: string) {
    schedulingExaminerIds = schedulingExaminerIds.includes(id)
      ? schedulingExaminerIds.filter((selected) => selected !== id)
      : [...schedulingExaminerIds, id];
  }

  async function scheduleSeminar() {
    schedulingError = "";
    schedulingSuccess = "";
    const date = new Date(schedulingAt);
    if (!schedulingAt || Number.isNaN(date.getTime()) || date.getTime() < Date.now() + MIN_LEAD_MS) {
      schedulingError = "Jadwal Seminar harus minimal 3 hari dari sekarang.";
      return;
    }
    if (!schedulingRoom.trim()) {
      schedulingError = "Ruangan wajib diisi.";
      return;
    }
    if (schedulingExaminerIds.length < 2) {
      schedulingError = "Pilih minimal 2 Penguji.";
      return;
    }

    schedulingLoading = true;
    try {
      const res = await (api.api.v1.seminars as any)[":id"].schedule.$put({
        param: { id: schedulingId },
        json: {
          scheduled_at: date.toISOString(),
          room: schedulingRoom.trim(),
          examiner_ids: schedulingExaminerIds,
        },
      });
      if (!res.ok) {
        schedulingError = await readError(res, "Gagal menyimpan jadwal Seminar.");
        return;
      }
      schedulingSuccess = "Jadwal Seminar berhasil disimpan.";
      schedulingId = "";
      await load();
    } catch {
      schedulingError = "Gagal menyimpan jadwal Seminar.";
    } finally {
      schedulingLoading = false;
    }
  }

  $effect(() => {
    status; page;
    load();
  });

  $effect(() => {
    if (canSchedule) loadExaminers();
  });

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
  {#if schedulingSuccess}
    <div role="status" class="rounded-md border border-success/40 bg-success-50 px-3 py-2 text-sm text-success">
      {schedulingSuccess}
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
            <div class="text-right text-xs text-st-muted md:block">
              <p class="font-medium text-st-text">Penguji</p>
              {#each sem.examiners as e (e.id)}
                <p>{e.fullName ?? e.full_name}</p>
              {/each}
            </div>
          {/if}
          {#if canSchedule && (sem.status === "pending" || sem.status === "scheduled")}
            {#if schedulingId === sem.id}
              <div class="w-full rounded-xl border border-st-stroke bg-st-bg p-4" aria-label="Form jadwal Seminar">
                <div class="grid gap-3 md:grid-cols-3">
                  <label class="text-xs font-medium text-st-muted">
                    Waktu Seminar
                    <input
                      type="datetime-local"
                      min={minDatetimeLocalValue(MIN_LEAD_MS)}
                      bind:value={schedulingAt}
                      class="mt-1 w-full rounded-md border border-st-stroke bg-st-surface px-3 py-2 text-sm text-st-text"
                    />
                  </label>
                  <label class="text-xs font-medium text-st-muted">
                    Ruangan
                    <input
                      type="text"
                      bind:value={schedulingRoom}
                      placeholder="Ruang 101"
                      maxlength="100"
                      class="mt-1 w-full rounded-md border border-st-stroke bg-st-surface px-3 py-2 text-sm text-st-text"
                    />
                  </label>
                  <div class="text-xs font-medium text-st-muted">
                    Penguji (minimal 2)
                    <div class="mt-1 max-h-28 space-y-1 overflow-y-auto rounded-md border border-st-stroke bg-st-surface p-2">
                      {#each examinerOptions as examiner (examiner.id)}
                        <label class="flex items-center gap-2 text-sm font-normal text-st-text">
                          <input
                            type="checkbox"
                            checked={schedulingExaminerIds.includes(examiner.id)}
                            onchange={() => toggleExaminer(examiner.id)}
                          />
                          <span>{examiner.fullName ?? examiner.full_name}</span>
                        </label>
                      {/each}
                      {#if examinerOptions.length === 0}
                        <span class="text-st-muted">Belum ada Penguji aktif.</span>
                      {/if}
                    </div>
                  </div>
                </div>
                {#if schedulingError}
                  <p role="alert" class="mt-3 text-sm text-danger-700"> {schedulingError}</p>
                {/if}
                <div class="mt-3 flex flex-wrap justify-end gap-2">
                  <button
                    type="button"
                    onclick={() => (schedulingId = "")}
                    class="inline-flex h-8 items-center rounded-md border border-st-stroke bg-st-surface px-3 text-sm text-st-text hover:bg-st-surface-hi"
                  >
                    Batal
                  </button>
                  <button
                    type="button"
                    disabled={schedulingLoading}
                    onclick={scheduleSeminar}
                    class="inline-flex h-8 items-center rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground hover:bg-primary-700 disabled:opacity-50"
                  >
                    {schedulingLoading ? "Menyimpan…" : sem.status === "scheduled" ? "Simpan Perubahan" : "Jadwalkan"}
                  </button>
                </div>
              </div>
            {:else}
              <button
                type="button"
                onclick={() => openScheduling(sem)}
                class="w-full rounded-md border border-primary/40 px-3 py-2 text-sm font-medium text-primary hover:bg-primary-50"
              >
                {sem.status === "scheduled" ? "Ubah Jadwal / Penguji" : "Jadwalkan Seminar"}
              </button>
            {/if}
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

  <Pagination {total} {page} pageSize={PAGE_SIZE} onPage={(next) => (page = next)} />
</div>
