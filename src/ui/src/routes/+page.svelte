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
	<header class="flex items-baseline justify-between pt-7">
		<span class="font-label text-sm font-semibold tracking-[0.3em] text-bone uppercase">Cadence</span>
		<span class="label">Web radio</span>
	</header>

	<Player />

	<!-- The two panes are one panel with a tabbed header, the way a piece of
	     equipment switches what its display is showing. -->
	<section class="panel mt-5">
		<div class="flex border-b border-edge">
			{#each tabs as entry (entry.id)}
				<button
					onclick={() => (tab = entry.id)}
					aria-current={tab === entry.id ? 'page' : undefined}
					class="label border-b-2 px-4 py-3 transition sm:px-5 {tab === entry.id
						? 'border-signal text-signal'
						: 'border-transparent hover:text-bone-dim'}"
				>
					{entry.label}
				</button>
			{/each}
		</div>

		{#if tab === 'request'}
			<RequestTab />
		{:else}
			<HistoryTab />
		{/if}
	</section>

	<footer class="mt-8 flex items-center justify-between">
		<span class="label">Cadence Radio</span>
		<span class="label">
			<a class="transition hover:text-signal" target="_blank" rel="noreferrer" href="https://github.com/kenellorando/cadence">Source</a>
			<span class="px-2 text-edge-light">/</span>
			<a class="transition hover:text-signal" target="_blank" rel="noreferrer" href="https://github.com/kenellorando/cadence/wiki/API-Reference">API</a>
		</span>
	</footer>
</div>
