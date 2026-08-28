<script lang="ts">
  import { goto } from "$app/navigation";
  import { ArrowLeft, ArrowRight } from "lucide-svelte";
  import FlickeringGrid from "$lib/components/FlickeringGrid.svelte";
  import { login } from "$lib/auth.store";

  let email = $state("");
  let password = $state("");
  let error = $state<string | null>(null);
  let loading = $state(false);
  let errors = $state<{ email?: string; password?: string }>({});

  const year = new Date().getFullYear();

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    error = null;
    errors = {};
    if (!email) errors.email = "Email wajib diisi";
    else if (!/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(email)) errors.email = "Format email tidak valid";
    if (!password) errors.password = "Password wajib diisi";
    if (Object.keys(errors).length > 0) return;

    loading = true;
    try {
      await login(email, password);
      goto("/");
    } catch (err) {
      error = err instanceof Error ? err.message : "Login gagal. Periksa email dan password Anda.";
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Autentikasi — SIMTAS FILKOM</title>
</svelte:head>

<div class="grid min-h-screen lg:grid-cols-2">
  <!-- Brand panel (left, desktop only) -->
  <div
    class="relative hidden overflow-hidden border-r border-border bg-background lg:flex lg:flex-col lg:justify-between lg:p-12"
  >
    <FlickeringGrid
      color="#07a2b6"
      maxOpacity={0.35}
      className="pointer-events-none absolute inset-0"
    />
    <div
      aria-hidden="true"
      class="pointer-events-none absolute left-1/2 top-[-10%] h-160 w-160 -translate-x-1/2 rounded-full opacity-20 blur-[120px]"
      style="background: radial-gradient(circle, var(--st-accent-to), transparent 60%)"
    ></div>

    <a href="/" class="accent-ring relative flex w-fit items-center gap-2.5">
      <span
        class="flex h-9 w-9 items-center justify-center rounded-full bg-background"
      >
        <span class="font-display text-[15px] italic text-foreground">sf</span>
      </span>
      <span class="text-sm font-medium tracking-tight text-foreground">
        SIMTAS <span class="text-muted-foreground">FILKOM</span>
      </span>
    </a>

    <div class="relative max-w-md">
      <p class="landing-eyebrow">SIMTAS://FILKOM</p>
      <h1 class="landing-heading mt-5 text-4xl md:text-5xl">
        Satu sistem untuk seluruh perjalanan
        <span class="accent-text italic">Tugas Akhir Skripsi</span>.
      </h1>
      <p class="mt-5 text-sm leading-relaxed text-muted-foreground md:text-base">
        Dari pengajuan judul, bimbingan, seminar, hingga sidang dan arsip—seluruh
        proses Tugas Akhir Fakultas Ilmu Komputer dalam satu ekosistem digital.
      </p>
    </div>

    <p
      class="relative font-mono text-[0.7rem] uppercase tracking-[0.2em] text-muted-foreground"
    >
      Fakultas Ilmu Komputer · Universitas Djuanda
    </p>
  </div>

  <!-- Login form (right) -->
  <div class="flex w-full flex-col justify-center px-6 py-10 sm:px-10">
    <div class="mx-auto w-full max-w-sm">
      <a
        href="/"
        class="mb-8 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition hover:text-foreground lg:hidden"
      >
        <ArrowLeft class="h-4 w-4" /> Kembali ke beranda
      </a>

      <div class="mb-8">
        <span class="landing-eyebrow">Masuk ke Sistem</span>
        <h2 class="landing-heading mt-3 text-3xl md:text-4xl">
          Selamat datang kembali.
        </h2>
        <p class="mt-2 text-sm text-muted-foreground">
          Gunakan kredensial yang diberikan administrator fakultas.
        </p>
      </div>

      {#if error}
        <div
          role="alert"
          class="mb-4 rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700"
        >
          {error}
        </div>
      {/if}

      <form onsubmit={handleSubmit} novalidate class="space-y-4">
        <div>
          <label for="email" class="mb-1.5 block text-sm font-medium text-foreground"
            >Email</label
          >
          <input
            id="email"
            type="email"
            bind:value={email}
            placeholder="nama@unida.ac.id"
            autocomplete="email"
            class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground outline-none transition focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
          />
          {#if errors.email}
            <p class="mt-1 text-xs text-danger-700">{errors.email}</p>
          {/if}
        </div>

        <div>
          <div class="flex items-center justify-between">
            <label
              for="password"
              class="mb-1.5 block text-sm font-medium text-foreground">Password</label
            >
            <a
              href="/forgot-password"
              class="mb-1.5 text-xs font-medium text-primary hover:underline"
              >Lupa password?</a
            >
          </div>
          <input
            id="password"
            type="password"
            bind:value={password}
            placeholder="••••••••"
            autocomplete="current-password"
            class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm text-foreground outline-none transition focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
          />
          {#if errors.password}
            <p class="mt-1 text-xs text-danger-700">{errors.password}</p>
          {/if}
        </div>

        <button
          type="submit"
          disabled={loading}
          class="flex w-full items-center justify-center gap-1.5 rounded-md bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground transition hover:bg-primary-700 active:scale-[0.98] disabled:opacity-60"
        >
          {loading ? "Masuk..." : "Masuk"}
          <ArrowRight class="h-4 w-4" />
        </button>
      </form>

      <p
        class="mt-8 text-center font-mono text-[0.7rem] uppercase tracking-[0.2em] text-muted-foreground"
      >
        FILKOM Universitas Djuanda — © {year}
      </p>
    </div>
  </div>
</div>
