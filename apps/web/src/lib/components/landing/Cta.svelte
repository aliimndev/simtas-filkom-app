<script lang="ts">
  import { ArrowRight } from "lucide-svelte";
  import Reveal from "./Reveal.svelte";
  import LandingButton from "./LandingButton.svelte";
</script>

<section class="relative overflow-hidden">
  <div aria-hidden="true" class="pointer-events-none absolute inset-0 flex items-center justify-center">
    <div class="h-160 w-160 rounded-full bg-(--st-accent-to) opacity-[0.09] blur-[120px]"></div>
  </div>
  <div class="landing-container relative flex flex-col items-center py-24 text-center md:py-32">
    <Reveal>
      <h2 class="landing-heading mx-auto max-w-2xl text-3xl md:text-5xl">Lanjutkan proses <span class="accent-text italic">Tugas Akhir Anda</span>.</h2>
    </Reveal>
    <Reveal delay={80}>
      <p class="mx-auto mt-5 max-w-xl text-sm leading-relaxed text-st-muted md:text-base">
        Akses SIMTAS FILKOM untuk mengelola proses Tugas Akhir secara terstruktur dan terdokumentasi.
      </p>
    </Reveal>
    <Reveal delay={160}>
      <div class="mt-9 flex flex-col items-center gap-3 sm:flex-row">
        <LandingButton href="/login">Masuk ke Sistem <ArrowRight size={16} /></LandingButton>
      </div>
    </Reveal>
  </div>
</section>
