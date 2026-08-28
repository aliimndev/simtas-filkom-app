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

  onMount(() => {
    if (!$auth.accessToken) {
      goto("/login");
      return;
    }
    ready = true;
  });
</script>

{#if ready}
  <div class="flex min-h-screen bg-st-bg">
    <div class="hidden lg:block">
      <Sidebar />
    </div>

    {#if mobileOpen}
      <div class="fixed inset-0 z-50 lg:hidden">
        <div class="absolute inset-0 bg-black/40" onclick={() => (mobileOpen = false)} aria-hidden="true"></div>
        <div class="absolute left-0 top-0 h-full w-60">
          <Sidebar onNavigate={() => (mobileOpen = false)} />
        </div>
        <button
          type="button"
          onclick={() => (mobileOpen = false)}
          class="absolute left-61 top-3 rounded-full bg-st-surface p-2 text-st-text"
          aria-label="Tutup menu"
        >
          <X size={20} />
        </button>
      </div>
    {/if}

    <div class="flex min-w-0 flex-1 flex-col">
      <TopBar onOpenMenu={() => (mobileOpen = true)} />
      <main class="mx-auto w-full max-w-350 flex-1 p-4 sm:p-6 lg:p-8">
        {@render children()}
      </main>
    </div>
  </div>
{/if}
