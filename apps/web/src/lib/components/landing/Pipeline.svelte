<script lang="ts">
  import { Archive, CalendarCheck, FileText, MessagesSquare, SearchCheck, UserCheck } from "lucide-svelte";

  type IconComponent = typeof FileText;
  import Reveal from "./Reveal.svelte";

  const STAGES: { step: string; title: string; desc: string; icon: IconComponent; span: string }[] = [
    { step: "01", title: "Pengajuan Judul", desc: "Mahasiswa mengajukan judul Tugas Akhir Skripsi secara digital.", icon: FileText, span: "lg:col-span-4" },
    { step: "02", title: "Review Judul", desc: "Kaprodi menyetujui atau menolak beserta catatan.", icon: SearchCheck, span: "lg:col-span-2" },
    { step: "03", title: "Penugasan Pembimbing", desc: "Dosen pembimbing ditetapkan untuk mendampingi.", icon: UserCheck, span: "lg:col-span-2" },
    { step: "04", title: "Bimbingan & Dokumen", desc: "Log bimbingan tercatat dan dokumen diunggah bertahap.", icon: MessagesSquare, span: "lg:col-span-4" },
    { step: "05", title: "Seminar & Sidang", desc: "Seminar proposal hingga sidang akhir terjadwal.", icon: CalendarCheck, span: "lg:col-span-3" },
    { step: "06", title: "Arsip & Kelulusan", desc: "Dokumen final diarsipkan dan status kelulusan dicatat.", icon: Archive, span: "lg:col-span-3" },
  ];
</script>

<section id="alur" class="relative overflow-hidden">
  <div aria-hidden="true" class="pointer-events-none absolute inset-x-0 top-0 h-80 bg-linear-to-b from-transparent to-(--st-accent-from) opacity-[0.04]"></div>
  <div class="landing-container py-20 md:py-28">
    <Reveal>
      <span class="landing-eyebrow"><span class="h-px w-8 bg-st-stroke"></span> Alur Tugas Akhir</span>
    </Reveal>
    <Reveal delay={80}>
      <h2 class="landing-heading mt-5 max-w-2xl text-3xl md:text-5xl">Enam tahap, satu <span class="accent-text italic">jejak</span> utuh.</h2>
    </Reveal>
    <Reveal delay={160}>
      <p class="mt-4 max-w-xl text-sm leading-relaxed text-st-muted md:text-base">
        Setiap mahasiswa melalui urutan yang sama—urutan ini bukan dekorasi, melainkan proses nyata yang menentukan maju atau tertunda.
      </p>
    </Reveal>

    <Reveal delay={220}>
      <div class="mt-16 grid grid-cols-1 gap-px overflow-hidden rounded-2xl border border-st-stroke bg-st-stroke sm:grid-cols-2 lg:grid-cols-6">
        {#each STAGES as stage}
          {@const Icon = stage.icon}
          <article class="relative flex min-h-44 flex-col overflow-hidden bg-st-surface p-6 transition-colors hover:bg-st-surface-hi md:min-h-52 md:p-8 {stage.span}">
            <div aria-hidden="true" class="pointer-events-none absolute -right-16 -top-16 h-40 w-40 rounded-full bg-(--st-accent-from)/10 blur-2xl"></div>
            <div class="relative flex items-start justify-between gap-4">
              <span class="accent-ring flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-st-surface-hi">
                <Icon size={24} class="text-(--st-accent-from)" aria-hidden="true" />
              </span>
              <span aria-hidden="true" class="select-none font-display text-5xl font-semibold tracking-tight text-st-text opacity-[0.07] md:text-6xl">{stage.step}</span>
            </div>
            <div class="relative mt-6 md:mt-8">
              <h3 class="text-xl font-semibold tracking-tight text-st-text md:text-2xl">{stage.title}</h3>
              <p class="mt-3 max-w-2xl text-sm leading-7 text-st-muted md:text-base">{stage.desc}</p>
            </div>
          </article>
        {/each}
      </div>
    </Reveal>
  </div>
</section>
