<script lang="ts">
  import type { StatusVariant } from "./thesis-status";

  let {
    variant,
    label,
    class: className = "",
  }: {
    variant: StatusVariant;
    label: string;
    class?: string;
  } = $props();

  const styles: Record<StatusVariant, string> = {
    pending: "bg-warning-50 text-warning border border-warning/20",
    approved: "bg-success-50 text-success border border-success/20",
    rejected: "bg-danger-50 text-danger-700 border border-danger/20",
    draft: "bg-st-surface-hi text-st-muted border border-st-stroke",
    in_progress: "bg-primary-50 text-primary border border-primary/20",
    completed: "bg-success-50 text-success border border-success/20",
  };
  const dots: Record<StatusVariant, string> = {
    pending: "bg-warning",
    approved: "bg-success",
    rejected: "bg-danger",
    draft: "bg-muted-foreground",
    in_progress: "bg-primary",
    completed: "bg-success",
  };
</script>

<span
  class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-[11px] font-medium {styles[variant]} {className}"
>
  <span class="h-1.5 w-1.5 rounded-full {dots[variant]}"></span>
  {label}
</span>
