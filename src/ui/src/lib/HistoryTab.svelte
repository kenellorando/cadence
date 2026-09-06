<script>
	import { radio } from './radio.svelte.js';
	import { timeAgo } from './timeAgo.js';

	// Newest first, without mutating the shared array.
	const entries = $derived([...radio.history].reverse());
</script>

{#if entries.length === 0}
	<p class="label px-4 py-10 text-center sm:px-5">Nothing logged yet</p>
{:else}
	<ul>
		{#each entries as song (song.Ended)}
			<li class="flex items-center gap-4 border-b border-edge/60 px-4 py-2.5 sm:px-5">
				<div class="min-w-0 flex-1">
					<p class="truncate text-sm text-bone" title={song.Title}>{song.Title}</p>
					<p class="truncate text-xs text-bone-faint" title={song.Artist}>{song.Artist}</p>
				</div>
				<span class="shrink-0 font-mono text-xs whitespace-nowrap text-bone-faint tabular-nums">
					{timeAgo(song.Ended)}
				</span>
			</li>
		{/each}
	</ul>
{/if}
