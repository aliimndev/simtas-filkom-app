<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { ChevronDown, LogOut, Menu, Search, UserRound, Bell } from "lucide-svelte";
  import { auth } from "$lib/auth.store";
  import { roleLabel, APP_NAME } from "$lib/constants/navigation";
  import ThemeToggle from "$lib/components/landing/ThemeToggle.svelte";

  let { onOpenMenu }: { onOpenMenu?: () => void } = $props();

  const user = $derived($auth.user);
  const displayName = $derived(user?.fullName ?? user?.full_name ?? "User");
  const initial = $derived(displayName[0]?.toUpperCase() ?? "?");

  let menuOpen = $state(false);
  let pathname = $state("/dashboard");
  onMount(() => {
    pathname = window.location.pathname;
    const onClick = (e: MouseEvent) => {
      const t = e.target as HTMLElement;
      if (!t.closest?.("[data-usermenu]")) menuOpen = false;
    };
    window.addEventListener("click", onClick);
    return () => window.removeEventListener("click", onClick);
  });

  function logout() {
    auth.set({ accessToken: null, user: null });
    goto("/login");
  }

  const ROUTE_LABELS: Record<string, string> = {
    dashboard: "Dashboard",
    thesis: "Tugas Akhir",
    theses: "Daftar Tugas Akhir",
    supervision: "Bimbingan",
    documents: "Dokumen",
    seminars: "Seminar",
    defenses: "Sidang",
    archives: "Arsip",
    profile: "Profil",
    schedules: "Jadwal",
    admin: "Administrasi",
    users: "Pengguna",
    "academic-years": "Tahun Akademik",
    "audit-logs": "Audit Log",
    "title-change-reviews": "Review Perubahan Judul",
    new: "Baru",
  };

  function buildBreadcrumb(p: string) {
    const segs = p.split("/").filter(Boolean);
    if (segs.length === 0) return [{ label: "Dashboard", href: "/dashboard" }];
    const out: { label: string; href: string }[] = [];
    let path = "";
    for (const s of segs) {
      path += "/" + s;
      const label = ROUTE_LABELS[s] ?? s.replace(/-/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
      out.push({ label, href: path });
    }
    return out;
  }

  const crumbs = $derived(buildBreadcrumb(pathname));
</script>

<header class="flex h-12 items-center justify-between gap-3 border-b border-st-stroke bg-st-surface px-4 lg:px-5">
  <div class="flex min-w-0 items-center gap-2">
    {#if onOpenMenu}
      <button
        type="button"
        onclick={onOpenMenu}
        class="rounded-full p-1.5 text-st-muted transition-colors hover:bg-st-surface-hi hover:text-st-text lg:hidden"
        aria-label="Buka menu"
      >
        <Menu size={20} />
      </button>
    {/if}
    <span class="truncate font-display text-[15px] text-st-text lg:hidden">{APP_NAME}</span>

    <nav aria-label="Breadcrumb" class="hidden lg:block">
      <ol class="flex items-center gap-1 font-mono text-[0.7rem] uppercase tracking-[0.15em]">
        {#each crumbs as item, i (i)}
          <li class="flex items-center gap-1">
            {#if i > 0}<span class="text-st-muted/40" aria-hidden="true">/</span>{/if}
            {#if i < crumbs.length - 1}
              <a href={item.href} class="text-st-muted transition-colors hover:text-st-text">{item.label}</a>
            {:else}
              <span class="font-medium text-st-text">{item.label}</span>
            {/if}
          </li>
        {/each}
      </ol>
    </nav>
  </div>

  <div class="flex items-center gap-1">
    <button
      type="button"
      class="hidden rounded-full p-1.5 text-st-muted transition-colors hover:bg-st-surface-hi hover:text-st-text sm:block"
      aria-label="Cari"
    >
      <Search size={18} />
    </button>

    <ThemeToggle />

    <button
      type="button"
      class="relative rounded-full p-1.5 text-st-muted transition-colors hover:bg-st-surface-hi hover:text-st-text"
      aria-label="Notifikasi"
    >
      <Bell size={18} />
    </button>

    <div class="relative" data-usermenu>
      <button
        type="button"
        onclick={() => (menuOpen = !menuOpen)}
        class="flex items-center gap-1.5 rounded-full px-1.5 py-1 text-sm font-medium text-st-text transition-colors hover:bg-st-surface-hi"
        aria-haspopup="menu"
        aria-expanded={menuOpen}
      >
        <span class="flex h-6 w-6 items-center justify-center rounded-full bg-(--st-accent-from)/10 font-display text-[10px] italic text-(--st-accent-to)">
          {initial}
        </span>
        <ChevronDown size={12} class="text-st-muted" />
      </button>
      {#if menuOpen}
        <div
          class="absolute right-0 top-full z-50 mt-1 w-48 overflow-hidden rounded-2xl border border-st-stroke bg-st-surface py-1 shadow-lg shadow-black/10"
          role="menu"
        >
          <div class="border-b border-st-stroke px-3 py-2">
            <p class="text-[13px] font-medium text-st-text">{displayName}</p>
            <p class="text-[11px] text-st-muted">{roleLabel(user?.role)}</p>
          </div>
          <a
            href="/profile"
            class="flex items-center gap-2 px-3 py-2 text-[13px] text-st-text transition-colors hover:bg-st-surface-hi"
            role="menuitem"
          >
            <UserRound size={14} /> Profil Saya
          </a>
          <button
            type="button"
            onclick={logout}
            class="flex w-full items-center gap-2 px-3 py-2 text-left text-[13px] text-danger-700 transition-colors hover:bg-danger-50"
            role="menuitem"
          >
            <LogOut size={14} /> Keluar
          </button>
        </div>
      {/if}
    </div>
  </div>
</header>
