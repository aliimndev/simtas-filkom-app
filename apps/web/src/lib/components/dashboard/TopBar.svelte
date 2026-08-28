<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { ChevronDown, LogOut, Menu, Search, UserRound, Bell } from "lucide-svelte";
  import { auth } from "$lib/auth.store";
  import { api } from "$lib/api";
  import { roleLabel, APP_NAME } from "$lib/constants/navigation";
  import ThemeToggle from "$lib/components/landing/ThemeToggle.svelte";

  let { onOpenMenu }: { onOpenMenu?: () => void } = $props();

  const user = $derived($auth.user);
  const displayName = $derived(user?.fullName ?? user?.full_name ?? "User");
  const initial = $derived(displayName[0]?.toUpperCase() ?? "?");

  let menuOpen = $state(false);
  let notificationOpen = $state(false);
  let notifications = $state<any[]>([]);
  let unreadCount = $state(0);
  let notificationLoading = $state(false);
  let notificationPanel = $state<HTMLElement>();
  let notificationReturnFocus: HTMLElement | null = null;
  let pathname = $state("/dashboard");

  function closeNotifications() {
    notificationOpen = false;
    const target = notificationReturnFocus;
    notificationReturnFocus = null;
    requestAnimationFrame(() => target?.focus());
  }

  function toggleNotifications() {
    if (notificationOpen) {
      closeNotifications();
      return;
    }
    notificationReturnFocus = document.activeElement as HTMLElement | null;
    notificationOpen = true;
    void loadNotifications();
  }

  function handleNotificationKeydown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      closeNotifications();
      return;
    }
    if (event.key !== "Tab" || !notificationPanel) return;

    const focusable = Array.from(
      notificationPanel.querySelectorAll<HTMLElement>(
        "button:not([disabled]), a[href], [tabindex]:not([tabindex='-1'])",
      ),
    );
    if (focusable.length === 0) return;

    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }

  async function loadNotifications() {
    if (!$auth.accessToken) return;
    notificationLoading = true;
    try {
      const [listResponse, countResponse] = await Promise.all([
        api.api.v1.notifications.$get({ query: { limit: "5" } }),
        api.api.v1.notifications["unread-count"].$get(),
      ]);
      if (listResponse.ok) notifications = (await listResponse.json()) as any[];
      if (countResponse.ok) unreadCount = Number((await countResponse.json()).unread_count ?? 0);
    } catch {
      notifications = [];
    } finally {
      notificationLoading = false;
    }
  }

  onMount(() => {
    pathname = window.location.pathname;
    void loadNotifications();
    const onClick = (e: MouseEvent) => {
      const t = e.target as HTMLElement;
      if (!t.closest?.("[data-usermenu]")) menuOpen = false;
      if (!t.closest?.("[data-notifications]") && notificationOpen) closeNotifications();
    };
    window.addEventListener("click", onClick);
    return () => window.removeEventListener("click", onClick);
  });

  $effect(() => {
    if (!notificationOpen || !notificationPanel) return;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const frame = requestAnimationFrame(() => {
      notificationPanel?.querySelector<HTMLElement>("button:not([disabled]), a[href]")?.focus();
    });
    return () => {
      cancelAnimationFrame(frame);
      document.body.style.overflow = previousOverflow;
    };
  });

  const searchHref = $derived(
    user?.role?.toLowerCase() === "mahasiswa"
      ? "/dashboard/thesis"
      : user?.role?.toLowerCase() === "dosen_pembimbing"
        ? "/dashboard/supervision"
        : "/dashboard/theses",
  );

  function notificationHref(link?: string | null) {
    if (!link) return "/dashboard";
    return link.startsWith("/dashboard/") ? link : link.startsWith("/") ? `/dashboard${link}` : "/dashboard";
  }

  async function openNotification(notification: any) {
    if (!notification.is_read) {
      try {
        await api.api.v1.notifications[":id"].read.$patch({ param: { id: notification.id } });
      } catch {
        // Keep navigation useful even if marking the notification fails.
      }
    }
    closeNotifications();
    goto(notificationHref(notification.link));
  }

  async function markAllNotificationsRead() {
    try {
      await api.api.v1.notifications["read-all"].$post();
      notifications = notifications.map((notification) => ({ ...notification, is_read: true }));
      unreadCount = 0;
    } catch {
      // Leave the current state intact when the API is unavailable.
    }
  }

  function formatNotificationDate(value?: string) {
    if (!value) return "";
    return new Date(value).toLocaleDateString("id-ID", { day: "2-digit", month: "short" });
  }

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
    <a
      href={searchHref}
      class="hidden rounded-full p-1.5 text-st-muted transition-colors hover:bg-st-surface-hi hover:text-st-text sm:block"
      aria-label="Cari dan buka daftar kerja"
    >
      <Search size={18} aria-hidden="true" />
    </a>

    <ThemeToggle />

    <div class="relative" data-notifications>
      <button
        type="button"
        onclick={toggleNotifications}
        class="relative rounded-full p-1.5 text-st-muted transition-colors hover:bg-st-surface-hi hover:text-st-text"
        aria-label={`Notifikasi${unreadCount > 0 ? `, ${unreadCount} belum dibaca` : ""}`}
        aria-expanded={notificationOpen}
        aria-haspopup="dialog"
      >
        <Bell size={18} aria-hidden="true" />
        {#if unreadCount > 0}
          <span class="absolute right-0.5 top-0.5 flex min-h-3.5 min-w-3.5 items-center justify-center rounded-full bg-danger px-1 text-[9px] font-bold leading-none text-white">
            {unreadCount > 9 ? "9+" : unreadCount}
          </span>
        {/if}
      </button>

      {#if notificationOpen}
        <div
          bind:this={notificationPanel}
          class="absolute right-0 top-full z-50 mt-2 w-[min(22rem,calc(100vw-2rem))] overflow-hidden rounded-2xl border border-st-stroke bg-st-surface shadow-lg shadow-black/10"
          role="dialog"
          tabindex="-1"
          aria-modal="true"
          aria-label="Notifikasi"
          onkeydown={handleNotificationKeydown}
        >
          <div class="flex items-center justify-between border-b border-st-stroke px-4 py-3">
            <div>
              <h2 class="text-sm font-semibold text-st-text">Notifikasi</h2>
              <p class="mt-0.5 text-xs text-st-muted">Pembaruan terbaru untuk Anda</p>
            </div>
            {#if unreadCount > 0}
              <button type="button" onclick={markAllNotificationsRead} class="text-xs font-medium text-(--st-accent-to) hover:underline">
                Tandai semua dibaca
              </button>
            {/if}
          </div>

          {#if notificationLoading}
            <p class="px-4 py-8 text-center text-sm text-st-muted">Memuat notifikasi…</p>
          {:else if notifications.length === 0}
            <p class="px-4 py-8 text-center text-sm text-st-muted">Belum ada notifikasi.</p>
          {:else}
            <div class="max-h-80 overflow-y-auto">
              {#each notifications as notification (notification.id)}
                <button
                  type="button"
                  onclick={() => openNotification(notification)}
                  class="flex w-full items-start gap-3 border-b border-st-stroke px-4 py-3 text-left transition-colors last:border-0 hover:bg-st-surface-hi {notification.is_read ? '' : 'bg-primary-50/50'}"
                >
                  <span class="mt-1.5 h-2 w-2 shrink-0 rounded-full {notification.is_read ? 'bg-st-stroke' : 'bg-(--st-accent-to)'}" aria-hidden="true"></span>
                  <span class="min-w-0 flex-1">
                    <span class="block text-sm font-medium text-st-text">{notification.title}</span>
                    <span class="mt-0.5 block line-clamp-2 text-xs leading-relaxed text-st-muted">{notification.message}</span>
                    <span class="mt-1 block text-[10px] uppercase tracking-[0.12em] text-st-muted">{formatNotificationDate(notification.created_at)}</span>
                  </span>
                </button>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
    </div>

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
            href="/dashboard/profile"
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
