<script lang="ts">
  import { ShieldCheck } from "lucide-svelte";
  import { api } from "$lib/api";

  let action = $state("");
  let page = $state(1);
  const perPage = 25;

  let list = $state<any[]>([]);
  let total = $state(0);
  let loading = $state(true);
  let error = $state("");

  const actionOptions: { value: string; label: string }[] = [
    { value: "", label: "Semua aksi" },
    { value: "USER_LOGIN", label: "Login" },
    { value: "USER_LOGOUT", label: "Logout" },
    { value: "USER_CREATED", label: "User dibuat" },
    { value: "THESIS_SUBMITTED", label: "Skripsi diajukan" },
    { value: "THESIS_REVIEWED", label: "Skripsi direview" },
    { value: "DOCUMENT_UPLOADED", label: "Dokumen diunggah" },
    { value: "SEMINAR_SCHEDULED", label: "Seminar dijadwalkan" },
    { value: "DEFENSE_SCHEDULED", label: "Sidang dijadwalkan" },
    { value: "THESIS_GRADUATED", label: "Wisuda" },
  ];

  async function load() {
    loading = true;
    error = "";
    try {
      const query: Record<string, string> = { page: String(page), per_page: String(perPage) };
      if (action) query.action = action;
      const res = await api.api.v1.auditLogs.$get({ query });
      const json: any = res.ok ? await res.json() : null;
      list = json?.data ?? [];
      total = json?.meta?.total ?? 0;
    } catch (e) {
      error = "Gagal memuat audit log.";
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    action; page;
    load();
  });

  function formatDateTime(s?: string) {
    if (!s) return "—";
    try {
      return new Date(s).toLocaleString("id-ID", {
        day: "2-digit",
        month: "short",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
    } catch {
      return s;
    }
  }
</script>

<div class="space-y-6">
  <div>
    <p class="landing-eyebrow">Audit Log</p>
    <h1 class="mt-2 text-balance landing-heading text-2xl">
      Jejak aktivitas seluruh <span class="accent-text italic">pengguna</span> sistem
    </h1>
  </div>

  <div class="flex flex-wrap items-center gap-3">
    <select
      bind:value={action}
      onchange={() => (page = 1)}
      class="w-56 rounded-md border border-st-stroke bg-st-surface px-3 py-2 text-sm text-st-text outline-none focus-visible:border-primary focus-visible:ring-2 focus-visible:ring-ring"
    >
      {#each actionOptions as opt}
        <option value={opt.value}>{opt.label}</option>
      {/each}
    </select>
  </div>

  {#if error}
    <div role="alert" class="rounded-2xl border border-danger-700/40 bg-danger-50 px-4 py-3 text-sm text-danger-700">
      {error}
    </div>
  {/if}

  {#if loading}
    <div class="rounded-2xl border border-st-stroke bg-st-surface py-10 text-center text-sm text-st-muted">
      Memuat audit log…
    </div>
  {:else}
    <div class="overflow-x-auto rounded-2xl border border-st-stroke bg-st-surface">
      <table class="w-full text-left text-sm">
        <thead class="border-b border-st-stroke text-xs uppercase tracking-[0.15em] text-st-muted">
          <tr>
            <th class="px-4 py-3 font-medium">Waktu</th>
            <th class="px-4 py-3 font-medium">Pengguna</th>
            <th class="px-4 py-3 font-medium">Aksi</th>
            <th class="px-4 py-3 font-medium">Entitas</th>
            <th class="px-4 py-3 font-medium">IP</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-st-stroke">
          {#each list as l (l.id)}
            <tr class="text-st-text">
              <td class="whitespace-nowrap px-4 py-3 font-mono text-xs text-st-muted">
                {formatDateTime(l.created_at ?? l.createdAt)}
              </td>
              <td class="px-4 py-3">
                {l.user?.full_name ?? l.user?.fullName ?? "—"}
              </td>
              <td class="px-4 py-3">
                <span class="inline-flex items-center rounded-md bg-st-surface-hi px-2 py-0.5 font-mono text-xs text-st-text">
                  {l.action}
                </span>
              </td>
              <td class="px-4 py-3 text-xs text-st-muted">
                {l.entity_type ?? l.entityType ?? "—"}
                {#if l.entity_id ?? l.entityId}
                  · {(l.entity_id ?? l.entityId).slice(0, 8)}…
                {/if}
              </td>
              <td class="px-4 py-3 font-mono text-xs text-st-muted">
                {l.ip_address ?? l.ipAddress ?? "—"}
              </td>
            </tr>
          {/each}
          {#if list.length === 0}
            <tr>
              <td colspan="5" class="px-4 py-12 text-center">
                <ShieldCheck size={36} class="mx-auto text-st-muted" />
                <p class="mt-3 landing-heading text-lg">
                  Belum ada <span class="accent-text italic">log</span> aktivitas
                </p>
                <p class="mt-1 text-sm text-st-muted">
                  Aktivitas sistem akan tercatat di sini secara otomatis.
                </p>
              </td>
            </tr>
          {/if}
        </tbody>
      </table>
    </div>

    {#if total > perPage}
      <div class="flex items-center justify-between pt-2">
        <p class="text-sm text-st-muted">Total {total} log · Halaman {page}</p>
        <div class="flex gap-2">
          <button
            type="button"
            disabled={page <= 1}
            onclick={() => (page -= 1)}
            class="inline-flex h-8 items-center rounded-md border border-st-stroke bg-st-surface px-3 text-sm text-st-text hover:bg-st-surface-hi disabled:opacity-50"
          >
            Sebelumnya
          </button>
          <button
            type="button"
            disabled={page >= Math.ceil(total / perPage)}
            onclick={() => (page += 1)}
            class="inline-flex h-8 items-center rounded-md border border-st-stroke bg-st-surface px-3 text-sm text-st-text hover:bg-st-surface-hi disabled:opacity-50"
          >
            Berikutnya
          </button>
        </div>
      </div>
    {/if}
  {/if}
</div>
