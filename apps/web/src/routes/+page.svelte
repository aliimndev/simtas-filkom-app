<script lang="ts">
  import Navbar from "$lib/components/landing/Navbar.svelte";
  import Hero from "$lib/components/landing/Hero.svelte";
  import Features from "$lib/components/landing/Features.svelte";
  import Pipeline from "$lib/components/landing/Pipeline.svelte";
  import Cta from "$lib/components/landing/Cta.svelte";
  import Footer from "$lib/components/landing/Footer.svelte";
</script>

<svelte:head>
  <title>SIMTAS FILKOM</title>
  <meta
    name="description"
    content="Satu sistem untuk seluruh perjalanan Tugas Akhir Skripsi — pengajuan judul, bimbingan, seminar, sidang, dan arsip. SIMTAS FILKOM Universitas Djuanda."
  />
</svelte:head>

<div class="simtas-dark flex min-h-screen flex-col">
  <Navbar />

  <main id="main-content" class="flex-1">
    <Hero />
    <Features />
    <Pipeline />
    <Cta />
  </main>

  <Footer />
</div>
