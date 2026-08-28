<script lang="ts">
  import { Archive, Bell, CalendarCheck, FileText, GraduationCap, LayoutDashboard, MessagesSquare, ShieldCheck, UserCheck, Users, type LucideIcon } from "lucide-svelte";
  import Reveal from "./Reveal.svelte";
  import NumberTicker from "./NumberTicker.svelte";

  const FEATURES: { icon: LucideIcon; title: string; desc: string }[] = [
    { icon: FileText, title: "Pengajuan Judul", desc: "Ajukan judul dan dapatkan review dari kaprodi secara digital." },
    { icon: MessagesSquare, title: "Bimbingan Online", desc: "Catat log bimbingan dan menunggu validasi dosen pembimbing." },
    { icon: CalendarCheck, title: "Seminar & Sidang", desc: "Penjadwalan, penguji, dan penilaian dalam satu alur." },
    { icon: Archive, title: "Arsip Digital", desc: "Dokumen final terdokumentasi dan mudah diakses kembali." },
    { icon: LayoutDashboard, title: "Dashboard Real-time", desc: "Statistik dan status per peran: mahasiswa, dosen, kaprodi, admin." },
    { icon: Bell, title: "Notifikasi Email", desc: "Pemberitahuan penting terkirim ke pemangku kepentingan." },
  ];

  const ROLES: { icon: LucideIcon; title: string }[] = [
    { icon: GraduationCap, title: "Mahasiswa" },
    { icon: UserCheck, title: "Dosen Pembimbing" },
    { icon: Users, title: "Dosen Penguji" },
    { icon: ShieldCheck, title: "Kaprodi" },
    { icon: LayoutDashboard, title: "Administrator Fakultas" },
  ];

  const STATS: { value: number; suffix?: string; label: string }[] = [
    { value: 150, suffix: "+", label: "Mahasiswa Terlayani" },
    { value: 200, suffix: "+", label: "Dokumen Tersimpan" },
    { value: 100, suffix: "+", label: "Sidang Terproses" },
    { value: 30, suffix: "+", label: "Dosen Terbantu" },
  ];
</script>

<section id="fitur" aria-labelledby="kemampuan-title" class="relative scroll-mt-24 overflow-hidden">
  <div aria-hidden="true" class="pointer-events-none absolute inset-x-0 top-0 h-80 bg-linear-to-b from-transparent to-(--st-accent-from) opacity-[0.05]"></div>

  <div class="landing-container relative py-20 md:py-28">
    <div class="max-w-3xl">
      <Reveal>
        <span class="landing-eyebrow"><span class="h-1.5 w-1.5 rounded-full bg-(--st-accent-from)"></span> KEMAMPUAN</span>
      </Reveal>
      <Reveal delay={80}>
        <h2 id="kemampuan-title" class="landing-heading mt-5 text-balance text-3xl md:text-5xl">
          Semua proses dalam <span class="accent-text italic">satu sistem</span>.
        </h2>
      </Reveal>
      <Reveal delay={120}>
        <p class="mt-5 max-w-2xl text-sm leading-7 text-st-muted md:text-base">
          SIMTAS membantu setiap peran mengelola proses Tugas Akhir sesuai kebutuhan dan kewenangannya.
        </p>
      </Reveal>
    </div>

    <div class="mt-14 flex flex-col gap-4">
      {#each FEATURES as feature, i}
        {@const Icon = feature.icon}
        {@const number = String(i + 1).padStart(2, "0")}
        <div class="sticky top-20" style="z-index:{i + 1}">
          <article class="relative flex min-h-44 flex-col overflow-hidden rounded-2xl border border-st-stroke bg-st-surface p-6 md:min-h-52 md:p-8">
            <div aria-hidden="true" class="pointer-events-none absolute -right-16 -top-16 h-40 w-40 rounded-full bg-(--st-accent-from)/10 blur-2xl"></div>
            <div class="relative flex items-start justify-between gap-4">
              <span class="accent-ring flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-st-surface-hi">
                <Icon size={24} class="text-(--st-accent-from)" aria-hidden="true" />
              </span>
              <span aria-hidden="true" class="select-none font-display text-5xl font-semibold tracking-tight text-st-text opacity-[0.07] md:text-6xl">{number}</span>
            </div>
            <div class="relative mt-6 md:mt-8">
              <h3 class="text-xl font-semibold tracking-tight text-st-text md:text-2xl">{feature.title}</h3>
              <p class="mt-3 max-w-2xl text-sm leading-7 text-st-muted md:text-base">{feature.desc}</p>
            </div>
          </article>
        </div>
      {/each}
    </div>

    <div class="mt-16 grid grid-cols-2 gap-px overflow-hidden rounded-2xl border border-st-stroke bg-st-stroke sm:grid-cols-2 lg:grid-cols-4">
      {#each STATS as stat, i}
        <Reveal delay={i * 70} class="h-full">
          <div class="flex h-full flex-col gap-1 bg-st-surface px-6 py-8 text-center transition-colors hover:bg-st-surface-hi md:py-10">
            <span class="font-display text-3xl text-(--st-accent-from) tabular-nums md:text-4xl">
              <NumberTicker value={stat.value} />{#if stat.suffix}<span class="text-(--st-accent-to)">{stat.suffix}</span>{/if}
            </span>
            <span class="text-sm text-st-muted">{stat.label}</span>
          </div>
        </Reveal>
      {/each}
    </div>

    <div class="mt-16 border-t border-st-stroke pt-12">
      <Reveal><span class="landing-eyebrow">PENGGUNA</span></Reveal>
      <Reveal delay={80}>
        <h3 class="landing-heading mt-5 max-w-2xl text-2xl md:text-4xl">Dibuat untuk <span class="accent-text italic">setiap peran</span>.</h3>
      </Reveal>
      <div class="mt-8 flex flex-wrap gap-3">
        {#each ROLES as role, i}
          {@const Icon = role.icon}
          <Reveal delay={i * 45}>
            <div class="inline-flex items-center gap-2 rounded-full border border-st-stroke bg-st-surface px-4 py-2 text-sm text-st-text transition hover:bg-st-surface-hi">
              <Icon size={16} class="text-(--st-accent-from)" aria-hidden="true" />
              <span>{role.title}</span>
            </div>
          </Reveal>
        {/each}
      </div>
    </div>
  </div>
</section>
