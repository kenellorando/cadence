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

<div class="flex items-center gap-3 border-b border-edge px-4 py-3 sm:px-5">
	<label for="search" class="label shrink-0">Search</label>
	<div class="field flex min-w-0 flex-1 items-center px-3 py-1.5">
		<input
			id="search"
			type="text"
			bind:value={query}
			onkeyup={onKey}
			placeholder="Search by title, artist, or album"
			class="w-full bg-transparent font-display text-sm text-ink placeholder:text-ink-faint focus:outline-none"
		/>
	</div>
	{#if radio.searchStatus}
		<span class="label shrink-0 whitespace-nowrap">{radio.searchStatus}</span>
	{/if}
</div>

{#if radio.searchResults.length > 0}
	<ul>
		{#each radio.searchResults as song (song.ID)}
			<li
				class="flex items-center gap-4 border-b border-edge/60 px-4 py-2.5 transition hover:bg-surface-inset/60 sm:px-5"
			>
				<!-- The frame stays whether or not there is artwork, so rows keep a
				     consistent left edge. A song with no embedded art answers 404 and
				     the image removes itself, leaving the empty frame. -->
				<span class="well h-10 w-10 shrink-0 overflow-hidden">
					<img
						src="/api/song/{song.ID}/art"
						alt=""
						loading="lazy"
						class="h-full w-full object-cover"
						onerror={(event) => event.currentTarget.remove()}
					/>
				</span>
				<div class="min-w-0 flex-1">
					<p class="truncate text-sm text-ink" title={song.Title}>{song.Title}</p>
					<p class="truncate text-xs text-ink-faint" title={song.Artist}>{song.Artist}</p>
				</div>
				<button
					onclick={() => request(song.ID)}
					disabled={requested.has(song.ID)}
					class="label shrink-0 rounded-[2px] border border-edge-light px-3 py-1 transition hover:border-signal hover:text-signal disabled:cursor-default disabled:border-level/40 disabled:text-level"
				>
					{requested.has(song.ID) ? 'Queued' : 'Request'}
				</button>
			</li>
		{/each}
	</ul>
{:else}
	<p class="label px-4 py-10 text-center sm:px-5">
		{radio.searchStatus ? 'No matching tracks' : 'Loading library'}
	</p>
{/if}
