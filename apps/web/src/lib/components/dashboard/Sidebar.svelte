<script lang="ts">
  import { ChevronsUpDown, Settings, HelpCircle } from "lucide-svelte";
  import { auth } from "$lib/auth.store";
  import { navItemsForRoles, roleLabel } from "$lib/constants/navigation";

  let { onNavigate }: { onNavigate?: () => void } = $props();

  const user = $derived($auth.user);
  const sections = $derived(navItemsForRoles(user ? [user.role] : []));
  const initial = $derived((user?.fullName ?? user?.full_name ?? "?")[0].toUpperCase());

  function active(path: string) {
    if (typeof window === "undefined") return false;
    const p = window.location.pathname;
    return p === path || (path !== "/dashboard" && p.startsWith(path));
  }
</script>

<aside class="flex h-full w-60 flex-col border-r border-st-stroke bg-st-surface">
  <div class="flex h-14 items-center gap-2 border-b border-st-stroke px-4">
    <span class="accent-ring flex h-9 w-9 shrink-0 items-center justify-center rounded-full">
      <span class="flex h-[calc(100%-3px)] w-[calc(100%-3px)] items-center justify-center rounded-full bg-st-surface">
        <span class="font-display text-[15px] italic leading-none text-st-text">sf</span>
      </span>
    </span>
    <div class="min-w-0 flex-1">
      <p class="truncate font-display text-[15px] leading-none text-st-text">
        SIMTAS <span class="text-st-muted">FILKOM</span>
      </p>
    </div>
    <ChevronsUpDown size={14} class="shrink-0 text-st-muted" />
  </div>

  <nav class="flex-1 overflow-y-auto px-3 py-4">
    <div class="space-y-6">
      {#each sections as section, i (i)}
        <div>
          {#if section.title}
            <p class="mb-2 px-3 font-mono text-[0.7rem] uppercase tracking-[0.25em] text-st-muted">
              {section.title}
            </p>
          {/if}
          <div class="space-y-0.5">
            {#each section.items as item}
              {@const on = active(item.href)}
              <a
                href={item.href}
                onclick={onNavigate}
                class="accent-ring flex items-center gap-2.5 rounded-full px-3 py-1.5 text-[13px] font-medium transition-colors duration-100 {on
                  ? 'bg-st-surface-hi text-st-text'
                  : 'text-st-muted hover:bg-st-surface-hi hover:text-st-text'}"
              >
                <item.icon size={16} class="shrink-0 {on ? 'text-(--st-accent-to)' : ''}" />
                {item.label}
              </a>
            {/each}
          </div>
        </div>
      {/each}
    </div>

    <div class="mt-6 border-t border-st-stroke pt-4">
      <div class="space-y-0.5">
        <a
          href="/profile"
          onclick={onNavigate}
          class="flex items-center gap-2.5 rounded-full px-3 py-1.5 text-[13px] font-medium text-st-muted transition-colors hover:bg-st-surface-hi hover:text-st-text"
        >
          <Settings size={16} /> Pengaturan
        </a>
        <button
          type="button"
          class="flex w-full items-center gap-2.5 rounded-full px-3 py-1.5 text-[13px] font-medium text-st-muted transition-colors hover:bg-st-surface-hi hover:text-st-text"
          aria-label="Bantuan"
        >
          <HelpCircle size={16} /> Bantuan
        </button>
      </div>
    </div>
  </nav>

  <div class="border-t border-st-stroke px-3 py-2">
    <a
      href="/profile"
      onclick={onNavigate}
      class="accent-ring flex items-center gap-2.5 rounded-full px-2.5 py-2 text-sm transition-colors hover:bg-st-surface-hi"
    >
      <div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-(--st-accent-from)/10 font-display text-[11px] italic text-(--st-accent-to)">
        {initial}
      </div>
      <div class="min-w-0 flex-1 leading-tight">
        <p class="truncate text-[13px] font-medium text-st-text">{user?.fullName ?? user?.full_name ?? "User"}</p>
        <p class="truncate text-[11px] text-st-muted">{roleLabel(user?.role)}</p>
      </div>
    </a>
  </div>
</aside>
