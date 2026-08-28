<script lang="ts">
  import Navbar from "$lib/components/landing/Navbar.svelte";
  import Footer from "$lib/components/landing/Footer.svelte";

  const FAQ = [
    { q: "Siapa yang dapat menggunakan SIMTAS FILKOM?", a: "Mahasiswa, dosen pembimbing, dosen penguji, kaprodi, dan administrator fakultas — masing-masing dengan peran dan kewenangan berbeda." },
    { q: "Bagaimana cara mengajukan judul skripsi?", a: "Masuk ke sistem, buka menu Pengajuan, lalu ajukan judul. Kaprodi akan menyetujui atau menolak beserta catatan." },
    { q: "Apakah bimbingan dicatat di sistem?", a: "Ya. Log bimbingan dicatat dan menunggu validasi dosen pembimbing sehingga progres terdokumentasi." },
    { q: "Lupa password, bagaimana mengatasinya?", a: "Gunakan menu Lupa password pada layar login untuk meminta pengaturan ulang melalui email." },
  ];
</script>

<svelte:head>
  <title>Pusat Bantuan — SIMTAS FILKOM</title>
  <meta
    name="description"
    content="Temukan jawaban tentang penggunaan SIMTAS FILKOM, pengajuan judul, bimbingan, dan pemulihan password."
  />
</svelte:head>

<div class="simtas-dark flex min-h-screen flex-col">
  <Navbar />

  <main id="main-content" class="flex-1">
    <section class="landing-container py-32 md:py-40">
      <span class="landing-eyebrow">Pusat Bantuan</span>
      <h1 class="landing-heading mt-5 text-3xl md:text-5xl">Jawaban untuk langkah penting di SIMTAS.</h1>
      <div class="mt-10 max-w-3xl space-y-3">
        {#each FAQ as item}
          <details class="rounded-2xl border border-st-stroke bg-st-surface p-5">
            <summary class="cursor-pointer text-st-text font-medium">{item.q}</summary>
            <p class="mt-3 text-sm leading-7 text-st-muted">{item.a}</p>
          </details>
        {/each}
      </div>
    </section>
  </main>

  <Footer />
</div>
