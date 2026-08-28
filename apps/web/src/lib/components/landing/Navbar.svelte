<script lang="ts">
  import { onMount } from "svelte";
  import { ArrowRight, Menu, X } from "lucide-svelte";
  import ThemeToggle from "./ThemeToggle.svelte";
  import { auth } from "$lib/auth.store";

  const navLinks = [
    { href: "/#fitur", label: "Fitur" },
    { href: "/#alur", label: "Alur Proses" },
    { href: "/faq", label: "FAQ" },
  ];

  let open = $state(false);
  let scrolled = $state(false);
  let activeSection = $state<string | null>(null);
  let progress = $state(0);
  let menuButton: HTMLButtonElement;
  let mobileNav: HTMLElement;

  const loggedIn = $derived(!!$auth.accessToken);

  function onScroll() {
    const y = window.scrollY;
    scrolled = y > 40;
    const max = document.documentElement.scrollHeight - window.innerHeight;
    progress = max > 0 ? Math.min(1, y / max) : 0;
  }

  onMount(() => {
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });

    const sections = navLinks
      .map((l) => document.getElementById(l.href.split("#")[1] ?? ""))
      .filter((el): el is HTMLElement => el !== null);
    let io: IntersectionObserver | undefined;
    if (sections.length > 0) {
      io = new IntersectionObserver(
        (entries) => {
          let best: { id: string; ratio: number } | null = null;
          for (const e of entries) {
            if (e.isIntersecting && (!best || e.intersectionRatio > best.ratio)) {
              best = { id: e.target.id, ratio: e.intersectionRatio };
            }
          }
          activeSection = best ? `#${best.id}` : null;
        },
        { rootMargin: "-35% 0px -55% 0px" },
      );
      sections.forEach((s) => io!.observe(s));
    }

    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") open = false;
    };
    window.addEventListener("keydown", onKey);

    return () => {
      window.removeEventListener("scroll", onScroll);
      window.removeEventListener("keydown", onKey);
      io?.disconnect();
    };
  });

  function isActive(href: string) {
    return href.startsWith("/#") ? activeSection === href.slice(1) : false;
  }
</script>

<header class="fixed inset-x-0 top-0 z-50 flex justify-center px-4 pt-4 md:pt-6">
  <div
    aria-hidden="true"
    class="accent-gradient fixed inset-x-0 top-0 h-0.75 origin-left transition-transform duration-150 ease-out"
    style="transform: scaleX({progress})"
  ></div>

  <div class="relative w-full transition-all duration-300 ease-out {scrolled ? 'max-w-3xl' : 'max-w-5xl'}">
    <div
      class="flex w-full items-center justify-between rounded-full border py-2 transition-all duration-300 ease-out {scrolled
        ? 'border-st-stroke bg-(--st-surface)/70 px-2 shadow-sm shadow-black/5 backdrop-blur-md'
        : 'border-transparent bg-transparent px-3'}"
    >
      <a href="/" class="group flex min-w-0 items-center gap-2.5 pl-1">
        <span class="accent-ring relative flex h-9 w-9 shrink-0 items-center justify-center rounded-full">
          <span class="flex h-[calc(100%-3px)] w-[calc(100%-3px)] items-center justify-center rounded-full bg-st-bg">
            <span class="font-display text-[15px] italic text-st-text">sf</span>
          </span>
        </span>
        <span class="hidden truncate text-sm font-medium tracking-tight text-st-text sm:block">
          SIMTAS <span class="text-st-muted">FILKOM</span>
        </span>
      </a>

      <nav aria-label="Navigasi utama" class="hidden items-center gap-0.5 md:flex">
        {#each navLinks as l}
          <a
            href={l.href}
            aria-current={isActive(l.href) ? "page" : undefined}
            class="rounded-full px-3.5 py-1.5 text-sm transition {isActive(l.href)
              ? 'bg-st-surface-hi font-medium text-st-text'
              : 'text-st-muted hover:bg-st-surface-hi hover:text-st-text'}"
          >
            {l.label}
          </a>
        {/each}
      </nav>

      <div class="flex items-center gap-1.5">
        <ThemeToggle />
        {#if loggedIn}
          <a
            href="/"
            class="accent-ring hidden items-center gap-1.5 rounded-full border border-st-stroke bg-st-surface-hi px-4 py-1.5 text-sm text-st-text transition hover:text-st-text md:inline-flex"
          >
            Dashboard <ArrowRight size={14} />
          </a>
        {:else}
          <a
            href="/login"
            class="accent-ring hidden rounded-full bg-st-text px-4 py-1.5 text-sm font-medium text-st-bg transition hover:opacity-90 md:inline-block"
          >
            Masuk ke Sistem
          </a>
        {/if}
        <button
          bind:this={menuButton}
          type="button"
          aria-label={open ? "Tutup menu" : "Buka menu"}
          aria-expanded={open}
          aria-controls="mobile-nav"
          onclick={() => (open = !open)}
          class="inline-flex h-9 w-9 items-center justify-center rounded-full text-st-muted transition hover:bg-st-surface-hi hover:text-st-text md:hidden"
        >
          {#if open}<X size={20} />{:else}<Menu size={20} />{/if}
        </button>
      </div>
    </div>

    {#if open}
      <div
        aria-hidden="true"
        onclick={() => (open = false)}
        class="fixed inset-0 z-40 bg-transparent md:hidden"
      ></div>
    {/if}

    {#if open}
      <nav
        bind:this={mobileNav}
        id="mobile-nav"
        aria-label="Navigasi mobile"
        class="absolute inset-x-0 top-full z-50 mt-2 max-h-[75dvh] overflow-y-auto rounded-2xl border border-st-stroke bg-(--st-surface)/90 p-2 pb-[max(0.5rem,env(safe-area-inset-bottom))] shadow-lg shadow-black/10 backdrop-blur-md md:hidden"
      >
        {#each navLinks as l}
          <a
            href={l.href}
            aria-current={isActive(l.href) ? "page" : undefined}
            class="block rounded-xl px-4 py-3 text-sm transition {isActive(l.href)
              ? 'bg-st-surface-hi font-medium text-st-text'
              : 'text-st-muted hover:bg-st-surface-hi hover:text-st-text'}"
          >
            {l.label}
          </a>
        {/each}
        <a
          href={loggedIn ? "/" : "/login"}
          class="mt-1 flex items-center justify-between rounded-xl bg-st-text px-4 py-3 text-sm font-medium text-st-bg"
        >
          {loggedIn ? "Buka Dashboard" : "Masuk ke Sistem"}
          <ArrowRight size={16} />
        </a>
      </nav>
    {/if}
  </div>
</header>
