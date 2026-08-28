<script lang="ts">
  import { auth } from "$lib/auth.store";
</script>

{#if $auth.user}
  <p>Welcome {$auth.user.fullName} ({$auth.user.role})</p>
{:else}
  <a href="/login">Login</a>
{/if}
