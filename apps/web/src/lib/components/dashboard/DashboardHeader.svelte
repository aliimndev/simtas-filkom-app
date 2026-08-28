<script lang="ts">
  import Reveal from "$lib/components/landing/Reveal.svelte";

  let { name, subtitle }: { name: string; subtitle?: string } = $props();

  function greeting(): string {
    const h = new Date().getHours();
    if (h < 11) return "selamat pagi";
    if (h < 15) return "selamat siang";
    if (h < 18) return "selamat sore";
    return "selamat malam";
  }

  function formatDate(d: Date): string {
    return d.toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
  }

  const firstName = $derived(name.split(" ")[0] || name);
</script>

<Reveal>
  <div class="flex flex-wrap items-end justify-between gap-x-6 gap-y-3">
    <div>
      <p class="landing-eyebrow">Sistem Manajemen Tugas Akhir</p>
      <h1 class="mt-3 text-balance landing-heading text-3xl sm:text-4xl">
        <span class="accent-text italic">{firstName}</span>, {greeting()}.
      </h1>
      {#if subtitle}
        <p class="mt-2 text-sm text-st-muted">{subtitle}</p>
      {/if}
    </div>
    <p class="font-mono text-[0.7rem] uppercase tracking-[0.25em] tabular-nums text-st-muted" aria-hidden="true">
      {formatDate(new Date())}
    </p>
  </div>
</Reveal>
