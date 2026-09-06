<script>
	import { radio } from '$lib/radio.svelte.js';
	import Player from '$lib/Player.svelte';
	import RequestTab from '$lib/RequestTab.svelte';
	import HistoryTab from '$lib/HistoryTab.svelte';

	let tab = $state('request');

	$effect(() => {
		radio.loadAll();
		// connect() hands back its own teardown, which is what $effect wants.
		return radio.connect();
	});

	const tabs = [
		{ id: 'request', label: 'Request' },
		{ id: 'history', label: 'History' }
	];
</script>

<div class="mx-auto min-h-screen max-w-3xl px-4 pb-16 sm:px-6">
	<header class="flex items-center justify-between pt-8">
		<span class="text-sm font-semibold tracking-[0.2em] text-white/70 uppercase">Cadence</span>
		<span class="text-xs text-white/25">{radio.version}</span>
	</header>

	<Player />

	<nav class="mt-10 flex justify-center">
		<div class="inline-flex gap-1 rounded-full border border-white/10 bg-white/5 p-1">
			{#each tabs as entry (entry.id)}
				<button
					onclick={() => (tab = entry.id)}
					aria-current={tab === entry.id ? 'page' : undefined}
					class="rounded-full px-5 py-1.5 text-sm transition {tab === entry.id
						? 'bg-white text-neutral-900'
						: 'text-white/50 hover:text-white'}"
				>
					{entry.label}
				</button>
			{/each}
		</div>
	</nav>

	<div class="mt-6">
		{#if tab === 'request'}
			<RequestTab />
		{:else}
			<HistoryTab />
		{/if}
	</div>

	<footer class="mt-16 text-center text-xs text-white/25">
		<a class="transition hover:text-white/60" target="_blank" rel="noreferrer" href="https://github.com/kenellorando/cadence">Source</a>
		<span class="px-2">•</span>
		<a class="transition hover:text-white/60" target="_blank" rel="noreferrer" href="https://github.com/kenellorando/cadence/wiki/API-Reference">API</a>
	</footer>
</div>
