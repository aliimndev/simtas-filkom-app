<script lang="ts">
  import { ArrowUpRight } from "lucide-svelte";

  type IconComponent = typeof ArrowUpRight;
  import NumberTicker from "$lib/components/landing/NumberTicker.svelte";

  type StatTone = "primary" | "secondary" | "success" | "warning" | "danger";

  let {
    title,
    value,
    icon: Icon,
    href,
    suffix,
    tone = "primary",
  }: {
    title: string;
    value: string | number;
    icon?: IconComponent;
    href?: string;
    suffix?: string;
    tone?: StatTone;
  } = $props();

  const toneChip: Record<StatTone, string> = {
    primary: "bg-primary-50 text-primary",
    secondary: "bg-secondary-50 text-secondary-700",
    success: "bg-success-50 text-success",
    warning: "bg-warning-50 text-warning",
    danger: "bg-danger-50 text-danger-700",
  };
</script>

{#snippet inner()}
  <div class="h-full rounded-2xl border border-st-stroke bg-st-surface p-5 {Icon ? 'accent-ring' : ''}">
    {#if Icon}
      <div class="flex items-center gap-3">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full {toneChip[tone]}">
          <Icon size={18} aria-hidden="true" />
        </div>
        <p class="font-mono text-[0.7rem] uppercase leading-tight tracking-[0.2em] text-st-muted">
          {title}
        </p>
      </div>
      <p class="mt-3 font-display text-3xl leading-none tabular-nums tracking-tight text-st-text">
        {#if typeof value === "number"}
          <NumberTicker {value} />
        {:else}
          {value}
        {/if}
        {#if suffix}<span class="ml-1 font-body text-sm font-normal text-st-muted">{suffix}</span>{/if}
      </p>
    {:else}
      <h3 class="landing-heading text-lg">{title}</h3>
      <p class="mt-2 font-display text-3xl leading-none tabular-nums tracking-tight text-st-text">
        {#if typeof value === "number"}
          <NumberTicker {value} />
        {:else}
          {value}
        {/if}
        {#if suffix}<span class="ml-1 font-body text-sm font-normal text-st-muted">{suffix}</span>{/if}
      </p>
    {/if}
  </div>
{/snippet}

{#if href}
  <a
    {href}
    class="group block h-full rounded-2xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
    aria-label="{title}: {value}"
  >
    {@render inner()}
  </a>
{:else}
  {@render inner()}
{/if}

<style>
  .accent-ring { position: relative; }
</style>
