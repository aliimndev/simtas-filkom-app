<script lang="ts">
  import { ArrowRight, Archive, Bell, CalendarCheck, FileText, LayoutDashboard, MessagesSquare, Search } from "lucide-svelte";
  import LandingButton from "./LandingButton.svelte";

  const PREVIEW_STATS = [
    { label: "Judul Diproses", value: "128" },
    { label: "Bimbingan Aktif", value: "46" },
    { label: "Seminar Terjadwal", value: "12" },
    { label: "Arsip Final", value: "312" },
  ];

  const PREVIEW_ROWS = [
    { title: "Analisis Sentimen Kebijakan...", stage: "Bimbingan", tint: "text-(--st-accent-from)" },
    { title: "Sistem Rekomendasi Mata...", stage: "Seminar", tint: "text-(--st-accent-to)" },
    { title: "Deteksi Dini Plagiarisme...", stage: "Sidang", tint: "text-st-muted" },
  ];

  const SIDEBAR = [
    { icon: LayoutDashboard, label: "Dashboard", active: true },
    { icon: FileText, label: "Pengajuan" },
    { icon: MessagesSquare, label: "Bimbingan" },
    { icon: CalendarCheck, label: "Seminar" },
    { icon: Archive, label: "Arsip" },
  ];
</script>

<section class="relative overflow-hidden">
  <div class="landing-container pt-32 md:pt-40">
    <div class="mx-auto max-w-3xl text-center">
      <span class="landing-eyebrow">
        SIMTAS://FILKOM <span class="text-st-stroke">·</span> Universitas Djuanda
      </span>
      <h1 class="mt-6 text-balance text-4xl font-semibold leading-[1.05] tracking-tight text-st-text md:text-6xl">
        Satu sistem untuk seluruh perjalanan{' '}
        <span class="accent-text italic md:text-7xl md:leading-[1.1]">Tugas Akhir Skripsi.</span>
      </h1>
      <p class="mx-auto mt-6 max-w-2xl text-sm leading-7 text-st-muted md:text-base">
        Kelola pengajuan judul, bimbingan, seminar, sidang, hingga arsip dalam satu alur yang jelas dan terdokumentasi.
      </p>
      <div class="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
        <LandingButton href="/login">Masuk ke Sistem <ArrowRight size={16} /></LandingButton>
        <a
          href="/#alur"
          class="inline-flex h-12 items-center justify-center rounded-full border border-st-stroke bg-st-surface px-6 text-sm font-medium text-st-text transition hover:bg-st-surface-hi"
        >
          Lihat alur proses
        </a>
      </div>
    </div>

    <!-- Decorative product mockup (hidden from assistive tech) -->
    <div aria-hidden="true" class="hero-surface mt-14 overflow-hidden rounded-3xl border border-st-stroke bg-st-surface shadow-xl shadow-black/5">
      <div class="grid h-full grid-cols-1 text-left lg:grid-cols-[13.5rem_1fr]">
        <aside class="hidden flex-col gap-1 border-r border-st-stroke p-5 lg:flex">
          <div class="mb-4 flex items-center gap-2.5 pl-1">
            <span class="accent-ring flex h-8 w-8 items-center justify-center rounded-full bg-st-surface">
              <span class="font-display text-[13px] italic text-st-text">sf</span>
            </span>
            <span class="text-sm font-medium tracking-tight text-st-text">SIMTAS <span class="text-st-muted">FILKOM</span></span>
          </div>
          {#each SIDEBAR as item}
            <div class="flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-[13px] {item.active ? 'bg-st-surface-hi font-medium text-(--st-accent-from)' : 'text-st-muted'}">
              <item.icon size={16} aria-hidden="true" />
              {item.label}
            </div>
          {/each}
        </aside>

        <div class="flex min-w-0 flex-col">
          <div class="flex items-center justify-between gap-4 border-b border-st-stroke px-5 py-3">
            <div class="flex items-center gap-2 rounded-full border border-st-stroke bg-st-surface px-3 py-1.5 text-[13px] text-st-muted">
              <Search size={14} aria-hidden="true" /> Cari judul skripsi…
            </div>
            <div class="flex items-center gap-3">
              <span class="relative flex h-8 w-8 items-center justify-center rounded-full border border-st-stroke bg-st-surface text-st-muted">
                <Bell size={14} aria-hidden="true" />
                <span class="absolute right-1 top-1 h-1.5 w-1.5 rounded-full bg-(--st-accent-to)"></span>
              </span>
              <span class="flex h-8 w-8 items-center justify-center rounded-full bg-st-surface-hi font-display text-xs text-st-text">AR</span>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3 p-5 md:grid-cols-4">
            {#each PREVIEW_STATS as s}
              <div class="rounded-xl border border-st-stroke bg-st-surface p-4">
                <p class="text-[11px] uppercase tracking-[0.15em] text-st-muted">{s.label}</p>
                <p class="mt-1.5 font-display text-2xl text-st-text tabular-nums">{s.value}</p>
              </div>
            {/each}
          </div>

          <div class="mx-5 mb-5 rounded-xl border border-st-stroke bg-st-surface">
            <div class="grid grid-cols-[1fr_auto] gap-4 border-b border-st-stroke px-4 py-2.5 font-mono text-[11px] uppercase tracking-[0.2em] text-st-muted">
              <span>Judul</span>
              <span>Tahap</span>
            </div>
            {#each PREVIEW_ROWS as row}
              <div class="grid grid-cols-[1fr_auto] items-center gap-4 border-b border-st-stroke px-4 py-3 last:border-0">
                <span class="truncate text-[13px] text-st-text">{row.title}</span>
                <span class="rounded-full border border-st-stroke px-2.5 py-1 text-[11px] {row.tint}">{row.stage}</span>
              </div>
            {/each}
          </div>
        </div>
      </div>
    </div>
  </div>
</section>
