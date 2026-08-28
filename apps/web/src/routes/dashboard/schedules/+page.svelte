<script lang="ts">
  import { onMount } from "svelte";
  import {
    CalendarDays,
    ClipboardCheck,
    GraduationCap,
    MapPin,
  } from "lucide-svelte";
  import { api } from "$lib/api";
  import { auth } from "$lib/auth.store";
  import StatCard from "$lib/components/dashboard/StatCard.svelte";
  import Reveal from "$lib/components/landing/Reveal.svelte";

  type Item = {
    id: string;
    type: "seminar" | "defense";
    student_name?: string;
    thesis_title?: string;
    scheduled_at?: string;
    room?: string;
  };

  type Grouped = Record<string, Item[]>;

  let items = $state<Item[]>([]);
  let loading = $state(true);
  let error = $state("");
  let academicYear = $state<string | null>(null);

  async function load() {
    loading = true;
    error = "";
    try {
      const [schedulesRes, ayRes] = await Promise.allSettled([
        (api.api.v1 as any).schedules?.upcoming?.$get?.() ??
          Promise.resolve({ ok: false }),
        api.api.v1.academicYears.$get({ query: { active: "true" } }),
      ]);

      if (schedulesRes.status === "fulfilled" && schedulesRes.value?.ok) {
        const json: any = await schedulesRes.value.json();
        items = Array.isArray(json?.data) ? json.data : [];
      } else {
        items = [];
      }

      if (ayRes.status === "fulfilled" && ayRes.value?.ok) {
        const json: any = await ayRes.value.json();
        const list = Array.isArray(json?.data) ? json.data : [];
        const active = list.find((y: any) => y.isActive ?? y.is_active) ?? list[0];
        academicYear = active?.label ?? active?.name ?? null;
      }
    } catch (e) {
      error = "Gagal memuat jadwal.";
      items = [];
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    if ($auth.accessToken) load();
  });

  const grouped = $derived(
    items.reduce<Grouped>((acc, item) => {
      const key = item.scheduled_at
        ? new Date(item.scheduled_at).toDateString()
        : "Lainnya";
      (acc[key] ??= []).push(item);
      return acc;
    }, {})
  );

  const seminarCount = $derived(
    items.filter((i) => i.type === "seminar").length
  );
  const defenseCount = $derived(
    items.filter((i) => i.type === "defense").length
  );

  function dateKey(s?: string) {
    if (!s) return "—";
    try {
      return new Date(s).toLocaleDateString("id-ID", {
        weekday: "long",
        day: "2-digit",
        month: "long",
        year: "numeric",
      });
    } catch {
      return s;
    }
  }

  function dateTime(s?: string) {
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
  <Reveal>
    <div>
      <p class="landing-eyebrow">Jadwal Ujian</p>
      <h1 class="mt-2 text-balance landing-heading text-2xl">
        Seminar & <span class="accent-text italic">sidang</span> 14 hari ke depan
      </h1>
      {#if academicYear}
        <p class="mt-1 text-sm text-st-muted">Tahun akademik {academicYear}</p>
      {/if}
    </div>
  </Reveal>

  <Reveal delay={80}>
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
      <StatCard title="Seminar Terjadwal" value={seminarCount} icon={GraduationCap} tone="primary" />
      <StatCard title="Sidang Terjadwal" value={defenseCount} icon={ClipboardCheck} tone="success" />
    </div>
  </Reveal>

  {#if error}
    <div
      role="alert"
      class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700"
    >
      {error}
    </div>
  {/if}

  {#if loading}
    <div class="py-10 text-center text-sm text-st-muted">Memuat jadwal…</div>
  {:else if items.length === 0}
    <div class="rounded-2xl border border-st-stroke bg-st-surface py-12 text-center">
      <CalendarDays size={36} class="mx-auto text-st-muted" aria-hidden="true" />
      <p class="mt-3 landing-heading text-lg">
        Tidak ada <span class="accent-text italic">jadwal mendatang</span>
      </p>
      <p class="mt-1 text-sm text-st-muted">Jadwal ujian akan muncul di sini.</p>
    </div>
  {:else}
    <div class="space-y-6">
      {#each Object.entries(grouped) as [key, dayItems] (key)}
        <Reveal>
          <div>
            <p class="mb-2 text-sm font-semibold uppercase tracking-wide text-st-muted">
              {dateKey(dayItems[0]?.scheduled_at)}
            </p>
            <div class="space-y-2">
              {#each dayItems as item (`${item.type}-${item.id}`)}
                <article class="flex flex-wrap items-center gap-3 rounded-2xl border border-st-stroke bg-st-surface p-4">
                  <div
                    class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full {item.type === 'seminar'
                      ? 'bg-(--st-accent-from)/10 text-(--st-accent-to)'
                      : 'bg-success-50 text-success'}"
                  >
                    {#if item.type === "seminar"}
                      <GraduationCap size={20} aria-hidden="true" />
                    {:else}
                      <ClipboardCheck size={20} aria-hidden="true" />
                    {/if}
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex flex-wrap items-center gap-2">
                      <p class="font-medium text-st-text">{item.student_name ?? "Mahasiswa"}</p>
                      <span
                        class="inline-flex h-5 items-center rounded-full px-2 text-[0.7rem] font-medium {item.type === 'seminar'
                          ? 'bg-primary-50 text-primary'
                          : 'bg-success-50 text-success'}"
                      >
                        {item.type === "seminar" ? "Seminar" : "Sidang"}
                      </span>
                    </div>
                    <p class="mt-0.5 line-clamp-1 text-sm text-st-muted">
                      {item.thesis_title ?? ""}
                    </p>
                  </div>
                  <div class="flex flex-wrap items-center gap-3 text-sm text-st-muted">
                    <span class="inline-flex items-center gap-1">
                      <CalendarDays size={14} aria-hidden="true" />
                      {dateTime(item.scheduled_at)}
                    </span>
                    {#if item.room}
                      <span class="inline-flex items-center gap-1">
                        <MapPin size={14} aria-hidden="true" />
                        {item.room}
                      </span>
                    {/if}
                  </div>
                </article>
              {/each}
            </div>
          </div>
        </Reveal>
      {/each}
    </div>
  {/if}
</div>
