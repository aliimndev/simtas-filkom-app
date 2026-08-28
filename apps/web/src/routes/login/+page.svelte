<script lang="ts">
  import { login } from "$lib/auth.store";
  import { goto } from "$app/navigation";

  let email = "";
  let password = "";
  let error = "";

  async function submit() {
    error = "";
    try {
      await login(email, password);
      goto("/");
    } catch (e: any) {
      error = e.message;
    }
  }
</script>

<form on:submit|preventDefault={submit}>
  <input bind:value={email} type="email" placeholder="email" />
  <input bind:value={password} type="password" placeholder="password" />
  <button type="submit">Login</button>
  {#if error}<p>{error}</p>{/if}
</form>
