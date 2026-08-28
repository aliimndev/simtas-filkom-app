<script lang="ts">
  const navLinks = [
    { href: "/#fitur", label: "Fitur" },
    { href: "/#alur", label: "Alur Proses" },
    { href: "/faq", label: "FAQ" },
  ];
</script>

<footer>
  <div class="landing-container grid gap-12 py-16 md:grid-cols-4">
    <div class="md:col-span-2">
      <div class="flex items-center gap-2.5">
        <span class="accent-ring flex h-8 w-8 items-center justify-center rounded-full">
          <span class="flex h-[calc(100%-3px)] w-[calc(100%-3px)] items-center justify-center rounded-full bg-st-bg">
            <span class="font-display text-[13px] italic text-st-text">sf</span>
          </span>
        </span>
        <span class="text-sm font-medium tracking-tight text-st-text">SIMTAS <span class="text-st-muted">FILKOM</span></span>
      </div>
      <p class="mt-5 max-w-sm text-sm leading-relaxed text-st-muted">
        Sistem Manajemen Tugas Akhir Skripsi
        <br />
        Fakultas Ilmu Komputer · Universitas Djuanda
      </p>
      <ul class="mt-6 flex flex-wrap gap-x-5 gap-y-2 text-sm">
        {#each navLinks as l}
          <li><a href={l.href} class="text-st-muted transition hover:text-st-text">{l.label}</a></li>
        {/each}
        <li><a href="/login" class="text-st-muted transition hover:text-st-text">Masuk</a></li>
      </ul>
    </div>
  </div>

  <div class="border-t border-st-stroke">
    <div class="landing-container flex flex-col items-center justify-between gap-2 py-6 text-xs text-st-muted sm:flex-row">
      <p>© 2026 Fakultas Ilmu Komputer · Universitas Djuanda</p>
      <p>SIMTAS://FILKOM</p>
    </div>
  </div>
</footer>
