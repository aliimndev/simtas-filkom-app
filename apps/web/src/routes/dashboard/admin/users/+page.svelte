<script lang="ts">
  import { Power, PowerOff, KeyRound, UserPlus, Search, Users } from "lucide-svelte";
  import { api } from "$lib/api";
  import { auth } from "$lib/auth.store";
  import { roleLabel } from "$lib/constants/navigation";
  import StatCard from "$lib/components/dashboard/StatCard.svelte";
  import StatusBadge from "$lib/components/dashboard/StatusBadge.svelte";
  import Reveal from "$lib/components/landing/Reveal.svelte";
  import type { StatusVariant } from "$lib/components/dashboard/thesis-status";

  const PAGE_SIZE = 10;

  const ROLE_OPTIONS = [
    { value: "ADMIN_FAKULTAS", label: roleLabel("ADMIN_FAKULTAS") },
    { value: "KAPRODI", label: roleLabel("KAPRODI") },
    { value: "DOSEN_PEMBIMBING", label: roleLabel("DOSEN_PEMBIMBING") },
    { value: "DOSEN_PENGUJI", label: roleLabel("DOSEN_PENGUJI") },
    { value: "MAHASISWA", label: roleLabel("MAHASISWA") },
  ];

  function roleVariant(role?: string | null): StatusVariant {
    switch ((role ?? "").toUpperCase()) {
      case "ADMIN_FAKULTAS":
        return "completed";
      case "KAPRODI":
        return "in_progress";
      case "DOSEN_PEMBIMBING":
        return "approved";
      case "DOSEN_PENGUJI":
        return "pending";
      case "MAHASISWA":
        return "draft";
      default:
        return "draft";
    }
  }

  let q = $state("");
  let role = $state("");
  let page = $state(1);
  let list = $state<any[]>([]);
  let total = $state(0);
  let loading = $state(true);
  let error = $state("");

  let showCreate = $state(false);
  let form = $state({ email: "", full_name: "", role: "MAHASISWA", nim_nidn: "" });
  let creating = $state(false);
  let createError = $state("");

  let actionError = $state("");
  let togglingId = $state<string | null>(null);
  let resettingId = $state<string | null>(null);

  async function load() {
    loading = true;
    error = "";
    try {
      const query: Record<string, string> = { page: String(page), per_page: String(PAGE_SIZE) };
      if (q) query.search = q;
      if (role) query.role = role;
      const res = await api.api.v1.users.$get({ query });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const json: any = await res.json();
      list = json?.data ?? [];
      total = json?.meta?.total ?? 0;
    } catch (e) {
      error = "Gagal memuat daftar pengguna.";
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    q; role; page;
    load();
  });

  const totalAll = $derived(total);
  const activeCount = $derived(list.filter((u) => u.is_active !== false).length);
  const inactiveCount = $derived(list.filter((u) => u.is_active === false).length);

  async function createUser(e: Event) {
    e.preventDefault();
    creating = true;
    createError = "";
    try {
      const body: Record<string, string> = {
        email: form.email,
        fullName: form.full_name,
        role: form.role,
      };
      if (form.nim_nidn) body.nimNidn = form.nim_nidn;
      const res = await api.api.v1.users.$post({ json: body });
      if (!res.ok) {
        const j: any = await res.json().catch(() => null);
        throw new Error(j?.error?.message ?? `HTTP ${res.status}`);
      }
      form = { email: "", full_name: "", role: "MAHASISWA", nim_nidn: "" };
      showCreate = false;
      page = 1;
      await load();
    } catch (e: any) {
      createError = e?.message ?? "Gagal membuat pengguna.";
    } finally {
      creating = false;
    }
  }

  async function toggleActive(u: any) {
    togglingId = u.id;
    actionError = "";
    try {
      const res = await api.api.v1.users[":id"].deactivate.$patch({ param: { id: u.id } });
      if (!res.ok) {
        const j: any = await res.json().catch(() => null);
        throw new Error(j?.error?.message ?? `HTTP ${res.status}`);
      }
      await load();
    } catch (e: any) {
      actionError = e?.message ?? "Gagal memperbarui status pengguna.";
    } finally {
      togglingId = null;
    }
  }

  async function resetPassword(u: any) {
    if (typeof window !== "undefined" && !window.confirm(`Reset password untuk ${u.full_name ?? u.fullName ?? "pengguna"}?`)) {
      return;
    }
    resettingId = u.id;
    actionError = "";
    try {
      actionError = "Reset password: endpoint belum tersedia di API.";
    } catch (e: any) {
      actionError = e?.message ?? "Gagal mereset password.";
    } finally {
      resettingId = null;
    }
  }
</script>

<div class="space-y-6">
  <Reveal>
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <p class="landing-eyebrow">Manajemen Pengguna</p>
        <h1 class="mt-2 text-balance landing-heading text-2xl">
          Kelola akun <span class="accent-text italic">mahasiswa</span>, dosen, dan staf
        </h1>
      </div>
      <button
        type="button"
        onclick={() => { showCreate = !showCreate; createError = ""; }}
        class="inline-flex h-10 items-center justify-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
      >
        <UserPlus size={16} aria-hidden="true" />
        Tambah Pengguna
      </button>
    </div>
  </Reveal>

  {#if showCreate}
    <Reveal>
      <form onsubmit={createUser} class="rounded-2xl border border-st-stroke bg-st-surface p-6">
        <h2 class="text-base font-semibold text-st-text">Tambah Pengguna Baru</h2>
        <p class="mt-1 text-xs text-st-muted">Password sementara akan dibuat otomatis dan dikirim ke email pengguna.</p>

        {#if createError}
          <div role="alert" class="mt-4 rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
            {createError}
          </div>
        {/if}

        <div class="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <input
            type="text"
            placeholder="Nama lengkap"
            bind:value={form.full_name}
            required
            class="h-10 rounded-md border border-st-stroke bg-st-bg px-3 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
          />
          <input
            type="email"
            placeholder="Email"
            bind:value={form.email}
            required
            class="h-10 rounded-md border border-st-stroke bg-st-bg px-3 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
          />
          <input
            type="text"
            placeholder="NIM / NIDN"
            bind:value={form.nim_nidn}
            class="h-10 rounded-md border border-st-stroke bg-st-bg px-3 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
          />
          <select
            bind:value={form.role}
            class="h-10 rounded-md border border-st-stroke bg-st-bg px-3 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
          >
            {#each ROLE_OPTIONS as r (r.value)}
              <option value={r.value}>{r.label}</option>
            {/each}
          </select>
          <div class="flex gap-2 sm:col-span-2 lg:col-span-3">
            <button
              type="submit"
              disabled={creating}
              class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary-700 disabled:opacity-50"
            >
              {creating ? "Menyimpan…" : "Simpan"}
            </button>
            <button
              type="button"
              onclick={() => { showCreate = false; createError = ""; }}
              class="inline-flex h-10 items-center justify-center rounded-md border border-st-stroke bg-st-surface px-4 text-sm font-medium text-st-text transition-colors hover:bg-st-surface-hi"
            >
              Batal
            </button>
          </div>
        </div>
      </form>
    </Reveal>
  {/if}

  <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
    <StatCard title="Total Pengguna" value={totalAll} icon={Users} tone="primary" />
    <StatCard title="Aktif" value={activeCount} icon={Power} tone="success" />
    <StatCard title="Nonaktif" value={inactiveCount} icon={PowerOff} tone="danger" />
  </div>

  <div class="flex flex-wrap items-center gap-3">
    <div class="relative w-full max-w-xs">
      <Search size={16} class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-st-muted" aria-hidden="true" />
      <input
        type="search"
        placeholder="Cari nama / email / NIM…"
        bind:value={q}
        oninput={() => (page = 1)}
        class="h-10 w-full rounded-md border border-st-stroke bg-st-surface pl-9 pr-3 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
      />
    </div>
    <select
      bind:value={role}
      onchange={() => (page = 1)}
      class="h-10 w-44 rounded-md border border-st-stroke bg-st-surface px-3 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
    >
      <option value="">Semua peran</option>
      {#each ROLE_OPTIONS as r (r.value)}
        <option value={r.value}>{r.label}</option>
      {/each}
    </select>
  </div>

  {#if actionError}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {actionError}
    </div>
  {/if}

  {#if error}
    <div role="alert" class="rounded-md border border-danger-700/40 bg-danger-50 px-3 py-2 text-sm text-danger-700">
      {error}
    </div>
  {/if}

  {#if loading}
    <div class="rounded-2xl border border-st-stroke bg-st-surface py-10 text-center text-sm text-st-muted">
      Memuat pengguna…
    </div>
  {:else}
    <div class="space-y-3">
      {#each list as u (u.id)}
        {@const inactive = u.is_active === false}
        {@const rVariant = roleVariant(u.role)}
        <article class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-st-stroke bg-st-surface p-5">
          <div class="flex min-w-0 items-center gap-3">
            <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-(--st-accent-from)/10 font-display text-base italic text-(--st-accent-to)">
              {(u.full_name ?? u.fullName ?? u.email ?? "?")[0]?.toUpperCase()}
            </div>
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <p class="font-medium text-st-text">{u.full_name ?? u.fullName ?? "—"}</p>
                <StatusBadge variant={rVariant} label={roleLabel(u.role)} />
                <StatusBadge variant={inactive ? "rejected" : "approved"} label={inactive ? "Nonaktif" : "Aktif"} />
              </div>
              <p class="mt-1 line-clamp-1 text-sm text-st-muted">{u.email ?? "—"}</p>
              <p class="mt-0.5 text-xs text-st-muted">NIM/NIDN: {u.nim_nidn ?? u.nimNidn ?? "—"}</p>
            </div>
          </div>
          <div class="flex gap-1">
            <button
              type="button"
              title={inactive ? "Aktifkan" : "Nonaktifkan"}
              aria-label={inactive ? "Aktifkan pengguna" : "Nonaktifkan pengguna"}
              disabled={togglingId === u.id}
              onclick={() => toggleActive(u)}
              class="inline-flex h-9 w-9 items-center justify-center rounded-md border border-st-stroke bg-st-surface text-st-text transition-colors hover:bg-st-surface-hi disabled:opacity-50"
            >
              {#if inactive}
                <Power size={16} class="text-success" aria-hidden="true" />
              {:else}
                <PowerOff size={16} class="text-danger-700" aria-hidden="true" />
              {/if}
            </button>
            <button
              type="button"
              title="Reset password"
              aria-label="Reset password"
              disabled={resettingId === u.id}
              onclick={() => resetPassword(u)}
              class="inline-flex h-9 w-9 items-center justify-center rounded-md border border-st-stroke bg-st-surface text-st-text transition-colors hover:bg-st-surface-hi disabled:opacity-50"
            >
              <KeyRound size={16} aria-hidden="true" />
            </button>
          </div>
        </article>
      {/each}

      {#if list.length === 0}
        <div class="rounded-2xl border border-st-stroke bg-st-surface py-12 text-center">
          <Users size={36} class="mx-auto text-st-muted" aria-hidden="true" />
          <p class="mt-3 landing-heading text-lg">Tidak ada <span class="accent-text italic">pengguna</span> ditemukan</p>
          <p class="mt-1 text-sm text-st-muted">Coba ubah filter atau kata kunci pencarian Anda.</p>
        </div>
      {/if}
    </div>
  {/if}

  {#if total > PAGE_SIZE}
    <div class="flex flex-wrap items-center justify-between gap-3 pt-2">
      <p class="text-sm text-st-muted">Total {total} pengguna · Halaman {page}</p>
      <div class="flex gap-2">
        <button
          type="button"
          disabled={page <= 1}
          onclick={() => (page -= 1)}
          class="inline-flex h-9 items-center justify-center rounded-md border border-st-stroke bg-st-surface px-3 text-sm text-st-text transition-colors hover:bg-st-surface-hi disabled:opacity-50"
        >
          Sebelumnya
        </button>
        <button
          type="button"
          disabled={page >= Math.ceil(total / PAGE_SIZE)}
          onclick={() => (page += 1)}
          class="inline-flex h-9 items-center justify-center rounded-md border border-st-stroke bg-st-surface px-3 text-sm text-st-text transition-colors hover:bg-st-surface-hi disabled:opacity-50"
        >
          Berikutnya
        </button>
      </div>
    </div>
  {/if}
</div>
