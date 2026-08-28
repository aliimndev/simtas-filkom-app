<script lang="ts">
  import { ArrowLeft, ArrowRight, CheckCircle2 } from "lucide-svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { api } from "$lib/api";

  const token = $derived($page.url.searchParams.get("token") ?? "");

  let newPassword = $state("");
  let confirmPassword = $state("");
  let loading = $state(false);
  let error = $state<string | null>(null);
  let success = $state(false);

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    error = null;

    if (!token) {
      error = "Token reset password tidak ditemukan atau tidak valid.";
      return;
    }
    if (newPassword.length < 8) {
      error = "Password minimal 8 karakter.";
      return;
    }
    if (!/[A-Z]/.test(newPassword) || !/[0-9]/.test(newPassword)) {
      error = "Password harus memiliki huruf kapital dan angka.";
      return;
    }
    if (newPassword !== confirmPassword) {
      error = "Konfirmasi password tidak cocok.";
      return;
    }

    loading = true;
    try {
      const response = await api.api.v1.auth.password.reset.$post({
        json: { token, newPassword, confirmPassword },
      });
      if (!response.ok) {
        const body = await response.json().catch(() => null);
        throw new Error((body as any)?.error?.message ?? "Gagal mengatur ulang password.");
      }
      success = true;
    } catch (err) {
      error = err instanceof Error ? err.message : "Gagal mengatur ulang password.";
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Reset Password — SIMTAS FILKOM</title>
  <meta name="description" content="Atur ulang password akun SIMTAS FILKOM." />
</svelte:head>

<div class="flex min-h-screen items-center justify-center bg-background px-6 py-10 sm:px-10">
  <main id="main-content" class="w-full max-w-sm">
    <a href="/login" class="mb-8 inline-flex items-center gap-1.5 text-sm text-muted-foreground transition hover:text-foreground">
      <ArrowLeft class="h-4 w-4" aria-hidden="true" /> Kembali ke login
    </a>

    <div class="mb-8">
      <span class="landing-eyebrow">Pemulihan Akun</span>
      <h1 class="landing-heading mt-3 text-3xl md:text-4xl">Buat password baru.</h1>
      <p class="mt-2 text-sm text-muted-foreground">
        Gunakan minimal 8 karakter dengan satu huruf kapital dan satu angka.
      </p>
    </div>

    {#if success}
      <div role="status" aria-live="polite" class="rounded-2xl border border-primary/30 bg-primary-50 p-5 text-sm text-foreground">
        <div class="flex items-start gap-3">
          <CheckCircle2 class="mt-0.5 h-5 w-5 shrink-0 text-primary" aria-hidden="true" />
          <div>
            <p class="font-medium">Password berhasil diubah.</p>
            <p class="mt-1 text-muted-foreground">Silakan masuk menggunakan password baru Anda.</p>
          </div>
        </div>
        <button
          type="button"
          onclick={() => goto("/login")}
          class="mt-5 inline-flex items-center gap-1.5 text-sm font-medium text-primary hover:underline"
        >
          Ke halaman login <ArrowRight class="h-4 w-4" aria-hidden="true" />
        </button>
      </div>
    {:else}
      {#if error}
        <div id="reset-password-error" role="alert" aria-live="assertive" class="mb-4 rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
          {error}
        </div>
      {/if}

      <form onsubmit={handleSubmit} class="space-y-4">
        <div>
          <label for="new-password" class="mb-1.5 block text-sm font-medium text-foreground">Password baru</label>
          <input
            id="new-password"
            type="password"
            bind:value={newPassword}
            autocomplete="new-password"
            required
            minlength="8"
            aria-invalid={error ? "true" : undefined}
            aria-describedby={error ? "reset-password-error" : undefined}
            class="w-full rounded-md border border-border bg-background px-3 py-2.5 text-base text-foreground outline-none transition focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring sm:text-sm"
          />
        </div>
        <div>
          <label for="confirm-password" class="mb-1.5 block text-sm font-medium text-foreground">Konfirmasi password</label>
          <input
            id="confirm-password"
            type="password"
            bind:value={confirmPassword}
            autocomplete="new-password"
            required
            minlength="8"
            aria-invalid={error ? "true" : undefined}
            aria-describedby={error ? "reset-password-error" : undefined}
            class="w-full rounded-md border border-border bg-background px-3 py-2.5 text-base text-foreground outline-none transition focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring sm:text-sm"
          />
        </div>
        <button
          type="submit"
          disabled={loading || !token}
          aria-busy={loading}
          class="flex w-full items-center justify-center gap-1.5 rounded-md bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground transition hover:bg-primary-700 disabled:cursor-wait disabled:opacity-60"
        >
          {loading ? "Menyimpan password…" : "Simpan password baru"}
          <ArrowRight class="h-4 w-4" aria-hidden="true" />
        </button>
      </form>
    {/if}
  </main>
</div>
