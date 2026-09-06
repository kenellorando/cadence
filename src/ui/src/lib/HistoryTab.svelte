<script>
	import { radio } from './radio.svelte.js';
	import { timeAgo } from './timeAgo.js';

	// Newest first, without mutating the shared array.
	const entries = $derived([...radio.history].reverse());
</script>

{#if entries.length === 0}
	<p class="py-10 text-center text-sm text-white/35">Nothing has played yet.</p>
{:else}
	<ul class="divide-y divide-white/5 overflow-hidden rounded-xl border border-white/10">
		{#each entries as song (song.Ended)}
			<li class="flex items-center gap-4 bg-white/[0.02] px-4 py-3">
				<div class="min-w-0 flex-1">
					<p class="truncate font-medium text-white" title={song.Title}>{song.Title}</p>
					<p class="truncate text-sm text-white/45" title={song.Artist}>{song.Artist}</p>
				</div>
				<span class="shrink-0 text-xs whitespace-nowrap text-white/35">{timeAgo(song.Ended)}</span>
			</li>
		{/each}
	</ul>
{/if}
