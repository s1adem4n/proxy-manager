<script lang="ts">
  import { onMount } from "svelte";

  interface Proxy {
    id: string;
    match: string;
    upstream: string;
  }

  const API_URL = import.meta.env.DEV
    ? "http://localhost:8080"
    : window.location.origin;

  let proxies: Proxy[] = $state([]);
  const getProxies = async () => {
    const response = await fetch(`${API_URL}/proxies`);
    if (!response.ok) {
      console.error("Failed to fetch proxies");
      return;
    }
    const data = await response.json();
    proxies = data;
  };
  let containerProxies: Proxy[] = $state([]);
  const getContainerProxies = async () => {
    const response = await fetch(`${API_URL}/container-proxies`);
    if (!response.ok) {
      console.error("Failed to fetch container proxies");
      return;
    }
    const data = await response.json();
    containerProxies = data;
  };

  let match = $state("");
  let upstream = $state("");
  let key = $state("");

  let mounted = $state(false);

  $effect(() => {
    if (!mounted) return;

    localStorage.setItem("key", key);
  });

  onMount(() => {
    key = localStorage.getItem("key") || "";

    getProxies();
    getContainerProxies();

    const interval = setInterval(() => {
      getProxies();
      getContainerProxies();
    }, 5000);

    mounted = true;

    return () => {
      clearInterval(interval);
    };
  });

  const inputClass =
    "py-2 px-4 border border-gray-200 rounded-lg focus:outline-none focus:border-blue-500 w-full";
</script>

<div class="flex flex-col w-full max-w-xl mx-auto gap-4 p-4">
  <h1 class="text-4xl font-bold">Caddy Proxies</h1>
  <input
    type="password"
    class={inputClass}
    bind:value={key}
    placeholder="API Key"
  />
  <div class="flex gap-2">
    <input
      class={inputClass}
      type="text"
      bind:value={match}
      placeholder="Match"
    />
    <input
      class={inputClass}
      type="text"
      bind:value={upstream}
      placeholder="Upstream"
    />
    <button
      class="text-2xl"
      onclick={() => {
        fetch(`${API_URL}/proxies`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            "X-Key": key,
          },
          body: JSON.stringify({ match, upstream }),
        }).then(getProxies);
      }}
    >
      ✔️
    </button>
  </div>
  <div
    class="grid grid-cols-[auto_min-content_auto_min-content] justify-between gap-y-2"
  >
    {#each proxies.filter((proxy) => !containerProxies.some((cp) => cp.match === proxy.match)) as proxy}
      <a href="https://{proxy.match}" target="_blank" class="hover:underline">
        {proxy.match}
      </a>
      <span> &rarr; </span>
      <span>
        {proxy.upstream}
      </span>
      <div class="flex gap-2">
        <button
          onclick={() => {
            fetch(`${API_URL}/proxies/${proxy.id}`, {
              method: "DELETE",
              headers: {
                "X-Key": key,
              },
            }).then(getProxies);
          }}
        >
          🗑️
        </button>
      </div>
    {/each}
    {#if containerProxies.length > 0}
      <span class="col-span-full h-px bg-gray-100"></span>
      {#each containerProxies as proxy}
        <a href="https://{proxy.match}" target="_blank" class="hover:underline">
          {proxy.match}
        </a>
        <span> &rarr; </span>
        <span>
          {proxy.upstream}
        </span>
        <div class="flex gap-2">
          <button
            class="disabled:opacity-50 disabled:cursor-not-allowed"
            disabled
            onclick={() => {
              fetch(`${API_URL}/proxies/${proxy.id}`, {
                method: "DELETE",
                headers: {
                  "X-Key": key,
                },
              }).then(getProxies);
            }}
          >
            🗑️
          </button>
        </div>
      {/each}
    {/if}
  </div>
</div>
