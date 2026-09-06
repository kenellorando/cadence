<script>
	import { radio } from './radio.svelte.js';

	let query = $state('');

	function onKey(event) {
		if (event.key === 'Enter') radio.runSearch(query);
	}
</script>

<input
	type="text"
	bind:value={query}
	onkeyup={onKey}
	placeholder="Filter for a song or artist!"
	aria-label="Search for a song or artist"
	class="w-full rounded-full border border-neutral-300 px-4 py-2 outline-none focus:border-neutral-500 dark:border-neutral-700 dark:bg-neutral-900"
/>

{#if radio.searchResults.length > 0}
	<div class="mt-4 overflow-x-auto">
		<table class="w-full text-left">
			<thead class="border-b border-neutral-300 dark:border-neutral-700">
				<tr>
					<th class="py-2 pr-4 font-medium">Artist</th>
					<th class="py-2 pr-4 font-medium">Title</th>
					<th class="py-2 font-medium">Availability</th>
				</tr>
			</thead>
			<tbody>
				{#each radio.searchResults as song (song.ID)}
					<tr class="border-b border-neutral-200 dark:border-neutral-800">
						<td class="py-2 pr-4">{song.Artist}</td>
						<td class="py-2 pr-4">{song.Title}</td>
						<td class="py-2">
							<button
								onclick={() => radio.request(song.ID)}
								class="rounded border border-neutral-300 px-3 py-1 text-sm transition hover:bg-neutral-100 dark:border-neutral-700 dark:hover:bg-neutral-800"
							>
								Request
							</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}

<p class="my-2 text-center text-sm">{radio.searchStatus}</p>
