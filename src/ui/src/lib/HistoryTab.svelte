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
				<!-- Same frame as the request list, so the two read as one thing. A
				     song with no embedded art answers 404 and the image removes
				     itself, leaving the frame. -->
				<span class="well h-10 w-10 shrink-0 overflow-hidden">
					{#if song.ID}
						<img
							src="/api/song/{song.ID}/art"
							alt=""
							loading="lazy"
							class="h-full w-full object-cover"
							onerror={(event) => event.currentTarget.remove()}
						/>
					{/if}
				</span>
				<div class="min-w-0 flex-1">
					<p class="truncate text-sm text-ink" title={song.Title}>{song.Title}</p>
					<p class="truncate text-xs text-ink-faint" title={song.Artist}>{song.Artist}</p>
				</div>
				<span class="shrink-0 font-mono text-xs whitespace-nowrap text-ink-faint tabular-nums">
					{timeAgo(song.Ended)}
				</span>
			</li>
		{/each}
	</ul>
{/if}
