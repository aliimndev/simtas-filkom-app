<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { X } from "lucide-svelte";
  import { auth } from "$lib/auth.store";
  import Sidebar from "$lib/components/dashboard/Sidebar.svelte";
  import TopBar from "$lib/components/dashboard/TopBar.svelte";

  let { children } = $props();

  let mobileOpen = $state(false);
  let ready = $state(false);
  let mobileDrawer = $state<HTMLElement>();

  function handleDrawerKeydown(event: KeyboardEvent) {
    if (!mobileOpen) return;
    if (event.key === "Escape") {
      mobileOpen = false;
      return;
    }
    if (event.key !== "Tab" || !mobileDrawer) return;

    const focusable = Array.from(
      mobileDrawer.querySelectorAll<HTMLElement>("a[href], button:not([disabled])"),
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

  onMount(() => {
    if (!$auth.accessToken) {
      goto("/login");
      return;
    }
    ready = true;
  });

  $effect(() => {
    if (!mobileOpen || !mobileDrawer) return;
    const previous = document.activeElement as HTMLElement | null;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    const frame = requestAnimationFrame(() => {
      mobileDrawer?.querySelector<HTMLElement>("a[href], button:not([disabled])")?.focus();
    });
    return () => {
      cancelAnimationFrame(frame);
      document.body.style.overflow = previousOverflow;
      if (!mobileOpen) previous?.focus();
    };
  });
</script>

{#if ready}
  <div class="flex min-h-screen bg-st-bg">
    <div class="hidden lg:block">
      <Sidebar />
    </div>

    {#if mobileOpen}
      <div
        class="fixed inset-0 z-50 lg:hidden"
        role="dialog"
        tabindex="-1"
        aria-modal="true"
        aria-label="Navigasi dashboard"
        onkeydown={handleDrawerKeydown}
      >
        <div class="absolute inset-0 bg-black/40" onclick={() => (mobileOpen = false)} aria-hidden="true"></div>
        <div bind:this={mobileDrawer} class="absolute left-0 top-0 h-full w-60">
          <Sidebar onNavigate={() => (mobileOpen = false)} />
        </div>
        <button
          type="button"
          onclick={() => (mobileOpen = false)}
          class="absolute left-61 top-3 rounded-full bg-st-surface p-2 text-st-text"
          aria-label="Tutup menu"
        >
          <X size={20} aria-hidden="true" />
        </button>
      </div>
    {/if}

    <div class="flex min-w-0 flex-1 flex-col">
      <TopBar onOpenMenu={() => (mobileOpen = true)} />
      <main id="main-content" class="mx-auto w-full max-w-350 flex-1 p-4 sm:p-6 lg:p-8">
        {@render children()}
      </main>
    </div>
  </div>
{/if}
