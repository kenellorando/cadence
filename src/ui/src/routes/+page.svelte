<script>
	import { radio } from '$lib/radio.svelte.js';
	import { theme } from '$lib/theme.svelte.js';
	import Player from '$lib/Player.svelte';
	import RequestTab from '$lib/RequestTab.svelte';
	import HistoryTab from '$lib/HistoryTab.svelte';
	import ThemeTab from '$lib/ThemeTab.svelte';

	let tab = $state('request');

	$effect(() => {
		theme.load();
		radio.loadAll();
		// connect() hands back its own teardown, which is what $effect wants.
		return radio.connect();
	});

	const tabs = [
		{ id: 'request', label: 'Request' },
		{ id: 'history', label: 'History' },
		{ id: 'theme', label: 'Theme' }
	];
</script>

<div class="mx-auto min-h-screen max-w-3xl px-4 pt-7 pb-16 sm:px-6">
	<Player />

	<!-- The panes are one panel with a tabbed header, the way a piece of
	     equipment switches what its display is showing. -->
	<section class="panel mt-5">
		<div class="flex justify-center border-b border-edge">
			{#each tabs as entry (entry.id)}
				<button
					onclick={() => (tab = entry.id)}
					aria-current={tab === entry.id ? 'page' : undefined}
					class="label border-b-2 px-4 py-3 transition sm:px-5 {tab === entry.id
						? 'border-signal text-signal'
						: 'border-transparent hover:text-ink-dim'}"
				>
					{entry.label}
				</button>
			{/each}
		</div>

		{#if tab === 'request'}
			<RequestTab />
		{:else if tab === 'history'}
			<HistoryTab />
		{:else}
			<ThemeTab />
		{/if}
	</section>

	<footer class="mt-12 flex flex-col items-center gap-2 text-center">
		<span class="font-logo text-base text-logo uppercase sm:text-lg">{theme.logo}</span>
		<div class="label flex items-center gap-2.5">
			<span class="font-mono tracking-normal tabular-nums">{radio.version}</span>
			<span class="text-edge-light">/</span>
			<a
				class="transition hover:text-signal"
				target="_blank"
				rel="noreferrer"
				href="https://github.com/kenellorando/cadence">Source</a
			>
			<span class="text-edge-light">/</span>
			<a
				class="transition hover:text-signal"
				target="_blank"
				rel="noreferrer"
				href="https://github.com/kenellorando/cadence/wiki/API-Reference">API</a
			>
		</div>
	</footer>
</div>
