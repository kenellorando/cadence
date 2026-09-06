<script>
	import { radio } from '$lib/radio.svelte.js';
	import Player from '$lib/Player.svelte';
	import RequestTab from '$lib/RequestTab.svelte';
	import HistoryTab from '$lib/HistoryTab.svelte';

	let tab = $state('request');

	$effect(() => {
		radio.loadAll();
		// connect() returns its own teardown, which is what $effect wants back.
		return radio.connect();
	});

	const tabs = [
		{ id: 'request', label: 'Request' },
		{ id: 'history', label: 'History' }
	];
</script>

<main class="mx-auto max-w-2xl px-4">
	<Player />

	<div class="flex justify-center gap-2 border-b border-neutral-300 dark:border-neutral-700">
		{#each tabs as entry (entry.id)}
			<button
				onclick={() => (tab = entry.id)}
				aria-current={tab === entry.id ? 'page' : undefined}
				class="-mb-px border-b-2 px-4 py-2 transition {tab === entry.id
					? 'border-neutral-900 dark:border-neutral-100'
					: 'border-transparent text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-200'}"
			>
				{entry.label}
			</button>
		{/each}
	</div>

	<div class="py-4">
		{#if tab === 'request'}
			<RequestTab />
		{:else}
			<HistoryTab />
		{/if}
	</div>

	<footer class="mx-auto my-12 max-w-xs text-center text-sm text-neutral-500">
		<div>Cadence Radio {radio.version}</div>
		<div>
			<a class="underline" target="_blank" rel="noreferrer" href="https://github.com/kenellorando/cadence">Source</a>
			•
			<a class="underline" target="_blank" rel="noreferrer" href="https://github.com/kenellorando/cadence/wiki/API-Reference">API</a>
		</div>
	</footer>
</main>
