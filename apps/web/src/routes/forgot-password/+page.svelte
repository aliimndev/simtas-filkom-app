<script lang="ts">
  import { ArrowLeft, ArrowRight, CheckCircle2 } from "lucide-svelte";
  import { api } from "$lib/api";

  let email = $state("");
  let loading = $state(false);
  let error = $state<string | null>(null);
  let sent = $state(false);

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    error = null;
    loading = true;

    try {
      const response = await api.api.v1.auth.password.forgot.$post({
        json: { email: email.trim() },
      });
      if (!response.ok) {
        const body = await response.json().catch(() => null);
        throw new Error((body as any)?.error?.message ?? "Gagal mengirim tautan reset password.");
      }
      sent = true;
    } catch (err) {
      error = err instanceof Error ? err.message : "Gagal mengirim tautan reset password.";
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Lupa Password — SIMTAS FILKOM</title>
  <meta name="description" content="Minta tautan untuk mengatur ulang password akun SIMTAS FILKOM." />
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-background px-6 py-10 sm:px-10">
  <main id="main-content" class="w-full max-w-sm">
    <a href="/login" class="mb-8 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition hover:text-foreground">
      <ArrowLeft class="h-4 w-4" aria-hidden="true" /> Kembali ke login
    </a>

    <div class="mb-8">
      <span class="landing-eyebrow">Pemulihan Akun</span>
      <h1 class="landing-heading mt-3 text-3xl md:text-4xl">Atur ulang password.</h1>
      <p class="mt-2 text-sm text-muted-foreground">
        Masukkan email akun Anda. Jika terdaftar, tautan reset akan dikirim ke email tersebut.
      </p>
    </div>

    {#if sent}
      <div role="status" aria-live="polite" class="rounded-2xl border border-primary/30 bg-primary-50 p-5 text-sm text-foreground">
        <div class="flex items-start gap-3">
          <CheckCircle2 class="mt-0.5 h-5 w-5 shrink-0 text-primary" aria-hidden="true" />
          <div>
            <p class="font-medium">Permintaan berhasil diproses.</p>
            <p class="mt-1 text-muted-foreground">
              Periksa inbox email Anda untuk melanjutkan. Jika email terdaftar, tautan reset akan tersedia selama satu jam.
            </p>
          </div>
        </div>
        <a href="/login" class="mt-5 inline-flex items-center gap-1.5 text-sm font-medium text-primary hover:underline">
          Kembali ke login <ArrowRight class="h-4 w-4" aria-hidden="true" />
        </a>
      </div>
    {:else}
      {#if error}
        <div id="forgot-password-error" role="alert" aria-live="assertive" class="mb-4 rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
          {error}
        </div>
      {/if}

      <form onsubmit={handleSubmit} class="space-y-4">
        <div>
          <label for="email" class="mb-1.5 block text-sm font-medium text-foreground">Email</label>
          <input
            id="email"
            type="email"
            bind:value={email}
            placeholder="nama@unida.ac.id"
            autocomplete="email"
            required
            aria-invalid={error ? "true" : undefined}
            aria-describedby={error ? "forgot-password-error" : undefined}
            class="w-full rounded-md border border-border bg-background px-3 py-2.5 text-base text-foreground outline-none transition focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring sm:text-sm"
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          aria-busy={loading}
          class="flex w-full items-center justify-center gap-1.5 rounded-md bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground transition hover:bg-primary-700 disabled:cursor-wait disabled:opacity-60"
        >
          {loading ? "Mengirim tautan…" : "Kirim tautan reset"}
          <ArrowRight class="h-4 w-4" aria-hidden="true" />
        </button>
      </form>
    {/if}
  </main>
</div>
