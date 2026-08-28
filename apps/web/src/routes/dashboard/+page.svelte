<script lang="ts">
  import { FileText, MessagesSquare, BookOpen, GraduationCap, Users, Archive } from "lucide-svelte";
  import { auth } from "$lib/auth.store";
  import { api } from "$lib/api";
  import DashboardHeader from "$lib/components/dashboard/DashboardHeader.svelte";
  import StatCard from "$lib/components/ui/StatCard.svelte";
  import StatusBadge from "$lib/components/ui/StatusBadge.svelte";
  import { thesisStatusProps } from "$lib/components/dashboard/thesis-status";

  const user = $derived($auth.user);
  const role = $derived(user?.role ?? "");
  const name = $derived(user?.fullName ?? user?.full_name ?? "Pengguna");

  let summary = $state<any>(null);
  let recent = $state<any[]>([]);
  let loading = $state(true);
  let error = $state("");

  async function load() {
    loading = true;
    error = "";
    try {
      if (role === "ADMIN_FAKULTAS" || role === "KAPRODI") {
        const r = await api.api.v1.dashboard.summary.$get();
        if (!r.ok) throw new Error(`summary:${r.status}`);
        summary = await r.json();
      }
      const t = await api.api.v1.theses.$get({ query: { page: "1", per_page: "5" } });
      if (!t.ok) throw new Error(`theses:${t.status}`);
      const j = await t.json();
      recent = j.data ?? [];
    } catch {
      error = "Data dashboard belum dapat dimuat. Periksa koneksi Anda lalu coba lagi.";
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    role;
    load();
  });
</script>

<div class="space-y-8">
  <DashboardHeader
    {name}
    subtitle={role === "MAHASISWA" || role === "mahasiswa"
      ? "Pantau progres Tugas Akhir Skripsi Anda."
      : "Kelola proses Tugas Akhir Skripsi mahasiswa."}
  />

  {#if error}
    <div role="alert" class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-danger-700/40 bg-danger-50 px-4 py-3 text-sm text-danger-700">
      <span>{error}</span>
      <button
        type="button"
        onclick={load}
        class="inline-flex h-8 items-center rounded-md border border-danger-700/40 px-3 font-medium transition hover:bg-danger-100 disabled:opacity-50"
        disabled={loading}
      >
        {loading ? "Memuat…" : "Coba lagi"}
      </button>
    </div>
  {/if}

  {#if role === "ADMIN_FAKULTAS" || role === "KAPRODI" || role === "admin_fakultas" || role === "kaprodi"}
    <section class="grid grid-cols-2 gap-3 sm:grid-cols-4">
      <StatCard title="Judul Diproses" value={summary?.thesisPending ?? 0} icon={BookOpen} href="/dashboard/theses" tone="primary" />
      <StatCard title="Bimbingan Aktif" value={summary?.supervisionsActive ?? 0} icon={MessagesSquare} href="/dashboard/supervision" tone="primary" />
      <StatCard title="Seminar Terjadwal" value={summary?.seminarsUpcoming ?? 0} icon={GraduationCap} href="/dashboard/seminars" tone="primary" />
      <StatCard title="Sidang Bulan Ini" value={summary?.defensesThisMonth ?? 0} icon={GraduationCap} href="/dashboard/defenses" tone="primary" />
    </section>
  {/if}

  <section>
    <h2 class="landing-heading text-xl">Aksi Cepat</h2>
    <div class="mt-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
      {#if role === "MAHASISWA" || role === "mahasiswa"}
        <StatCard title="Tugas Akhir" value="Buka" icon={FileText} href="/dashboard/thesis" />
        <StatCard title="Bimbingan" value="Buka" icon={MessagesSquare} href="/dashboard/supervision" />
      {:else if role === "DOSEN_PEMBIMBING" || role === "dosen_pembimbing"}
        <StatCard title="Bimbingan" value="Review" icon={MessagesSquare} href="/dashboard/supervision" />
        <StatCard title="Dokumen" value="Review" icon={BookOpen} href="/dashboard/documents" />
      {:else}
        <StatCard title="Daftar Skripsi" value="Buka" icon={BookOpen} href="/dashboard/theses" />
        <StatCard title="Pengguna" value="Kelola" icon={Users} href="/dashboard/admin/users" />
        <StatCard title="Arsip" value="Buka" icon={Archive} href="/dashboard/archives" />
      {/if}
    </div>
  </section>

  <section>
    <div class="flex items-center justify-between">
      <h2 class="landing-heading text-xl">Aktivitas Terbaru</h2>
      <a href="/dashboard/theses" class="inline-flex items-center gap-1 font-mono text-[0.7rem] uppercase tracking-[0.2em] text-(--st-accent-to) hover:text-(--st-accent-from)">
        Lihat semua
      </a>
    </div>
    <div class="mt-4 space-y-3">
      {#if loading}
        <div class="py-10 text-center text-sm text-st-muted">Memuat…</div>
      {:else if recent.length === 0}
        <div class="rounded-2xl border border-st-stroke bg-st-surface p-8 text-center">
          <BookOpen size={32} class="mx-auto text-st-muted" />
          <p class="mt-3 landing-heading text-lg">Belum ada <span class="accent-text italic">aktivitas</span></p>
          <p class="mt-1 text-sm text-st-muted">Data terbaru akan muncul di sini.</p>
        </div>
      {:else}
        {#each recent as t (t.id)}
          {@const sp = thesisStatusProps(t.status ?? "")}
          <article class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-st-stroke bg-st-surface p-5">
            <div class="min-w-0">
              <div class="flex flex-wrap items-center gap-2">
                <p class="font-medium text-st-text">{t.student?.fullName ?? t.student?.full_name ?? "—"}</p>
                <StatusBadge variant={sp.variant} label={sp.label} />
              </div>
              <p class="mt-1 line-clamp-1 text-sm text-st-muted">{t.title}</p>
            </div>
          </article>
        {/each}
      {/if}
    </div>
  </section>
</div>
