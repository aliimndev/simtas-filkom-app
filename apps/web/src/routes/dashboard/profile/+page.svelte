<script lang="ts">
  import { auth } from "$lib/auth.store";
  import { roleLabel } from "$lib/constants/navigation";
  import { goto } from "$app/navigation";

  const user = $derived($auth.user);

  function logout() {
    auth.set({ accessToken: null, user: null });
    goto("/login");
  }
</script>

<div class="space-y-6">
  <div>
    <p class="landing-eyebrow">Profil</p>
    <h1 class="mt-2 text-balance landing-heading text-2xl">Profil <span class="accent-text italic">Saya</span></h1>
  </div>

  <article class="rounded-2xl border border-st-stroke bg-st-surface p-6">
    <div class="flex items-center gap-4">
      <div class="flex h-14 w-14 items-center justify-center rounded-full bg-(--st-accent-from)/10 font-display text-xl italic text-(--st-accent-to)">
        {(user?.fullName ?? user?.full_name ?? "?")[0]?.toUpperCase()}
      </div>
      <div>
        <p class="text-lg font-medium text-st-text">{user?.fullName ?? user?.full_name ?? "—"}</p>
        <p class="text-sm text-st-muted">{roleLabel(user?.role)}</p>
      </div>
    </div>

    <dl class="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2">
      <div>
        <dt class="text-xs uppercase tracking-[0.2em] text-st-muted">Email</dt>
        <dd class="mt-1 text-sm text-st-text">{user?.email ?? "—"}</dd>
      </div>
      <div>
        <dt class="text-xs uppercase tracking-[0.2em] text-st-muted">Peran</dt>
        <dd class="mt-1 text-sm text-st-text">{roleLabel(user?.role)}</dd>
      </div>
      <div>
        <dt class="text-xs uppercase tracking-[0.2em] text-st-muted">ID</dt>
        <dd class="mt-1 font-mono text-xs text-st-muted">{user?.id ?? "—"}</dd>
      </div>
    </dl>
  </article>

  <div class="flex justify-end">
    <button
      type="button"
      onclick={logout}
      class="inline-flex h-10 items-center justify-center rounded-md border border-danger-700/40 bg-st-surface px-4 text-sm font-medium text-danger-700 transition-colors hover:bg-danger-50"
    >
      Keluar
    </button>
  </div>
</div>
