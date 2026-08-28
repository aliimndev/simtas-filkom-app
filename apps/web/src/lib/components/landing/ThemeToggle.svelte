<script lang="ts">
  import { Sun, Moon } from "lucide-svelte";
  import { onMount } from "svelte";

  let dark = $state(false);

  onMount(() => {
    dark = document.documentElement.classList.contains("dark");
  });

  function toggle() {
    dark = !dark;
    document.documentElement.classList.toggle("dark", dark);
    try {
      localStorage.setItem("theme", dark ? "dark" : "light");
    } catch {}
  }
</script>

<button
  type="button"
  onclick={toggle}
  aria-label={dark ? "Aktifkan mode terang" : "Aktifkan mode gelap"}
  class="inline-flex h-9 w-9 items-center justify-center rounded-full text-st-muted transition hover:bg-st-surface-hi hover:text-st-text"
>
  {#if dark}
    <Moon size={18} />
  {:else}
    <Sun size={18} />
  {/if}
</button>
