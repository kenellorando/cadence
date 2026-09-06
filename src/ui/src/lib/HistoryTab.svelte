<script>
	import { radio } from './radio.svelte.js';
	import { timeAgo } from './timeAgo.js';

	// Newest first, without mutating the shared array.
	const entries = $derived([...radio.history].reverse());
</script>

{#if entries.length === 0}
	<p class="my-2 text-center text-sm">No history available (yet).</p>
{:else}
	<div class="overflow-x-auto">
		<table class="w-full text-left">
			<thead class="border-b border-neutral-300 dark:border-neutral-700">
				<tr>
					<th class="py-2 pr-4 font-medium">Ended</th>
					<th class="py-2 pr-4 font-medium">Artist</th>
					<th class="py-2 font-medium">Title</th>
				</tr>
			</thead>
			<tbody>
				{#each entries as song (song.Ended)}
					<tr class="border-b border-neutral-200 dark:border-neutral-800">
						<td class="py-2 pr-4 whitespace-nowrap">{timeAgo(song.Ended)}</td>
						<td class="py-2 pr-4">{song.Artist}</td>
						<td class="py-2">{song.Title}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
