<script>
	import { radio } from './radio.svelte.js';

	let query = $state('');
	let requested = $state(new Set());

	function onKey(event) {
		if (event.key === 'Enter') radio.runSearch(query);
	}

	async function request(id) {
		await radio.request(id);
		// Mark the row so it is obvious which request just went through.
		requested = new Set(requested).add(id);
	}
</script>

<div class="relative">
	<span class="pointer-events-none absolute top-1/2 left-4 -translate-y-1/2 text-white/30">⌕</span>
	<input
		type="text"
		bind:value={query}
		onkeyup={onKey}
		placeholder="Search for a song or artist, then press Enter"
		aria-label="Search for a song or artist"
		class="w-full rounded-xl border border-white/10 bg-white/5 py-3 pr-4 pl-10 text-white placeholder:text-white/25 transition outline-none focus:border-accent/60 focus:bg-white/[0.07]"
	/>
</div>

{#if radio.searchResults.length > 0}
	<ul class="mt-4 divide-y divide-white/5 overflow-hidden rounded-xl border border-white/10">
		{#each radio.searchResults as song (song.ID)}
			<li
				class="flex items-center gap-4 bg-white/[0.02] px-4 py-3 transition hover:bg-white/[0.06]"
			>
				<div class="min-w-0 flex-1">
					<p class="truncate font-medium text-white" title={song.Title}>{song.Title}</p>
					<p class="truncate text-sm text-white/45" title={song.Artist}>{song.Artist}</p>
				</div>
				<button
					onclick={() => request(song.ID)}
					disabled={requested.has(song.ID)}
					class="shrink-0 rounded-full border border-white/15 px-4 py-1.5 text-sm transition hover:border-accent/60 hover:bg-accent/10 hover:text-white disabled:cursor-default disabled:border-emerald-500/30 disabled:bg-emerald-500/10 disabled:text-emerald-400"
				>
					{requested.has(song.ID) ? 'Queued' : 'Request'}
				</button>
			</li>
		{/each}
	</ul>
{:else if radio.searchStatus}
	<p class="py-10 text-center text-sm text-white/35">No songs matched that search.</p>
{/if}

{#if radio.searchStatus}
	<p class="mt-4 text-center text-sm text-white/45">{radio.searchStatus}</p>
{/if}
