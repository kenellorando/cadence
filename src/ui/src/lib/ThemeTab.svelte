<script>
	import { theme, THEMES } from './theme.svelte.js';
</script>

<ul>
	{#each THEMES as option (option.id)}
		{@const active = theme.current === option.id}
		<li>
			<button
				onclick={() => theme.set(option.id)}
				aria-pressed={active}
				class="flex w-full items-center gap-4 border-b border-edge/60 px-4 py-3 text-left transition sm:px-5 {active
					? 'bg-surface-inset'
					: 'hover:bg-surface-inset/50'}"
			>
				<!-- Literal values, not tokens: a swatch has to show its own theme's
				     colours while a different theme is on screen. -->
				<span class="flex shrink-0 overflow-hidden rounded-[2px] border border-edge">
					{#each option.swatch as colour (colour)}
						<span class="h-6 w-3" style="background: {colour}"></span>
					{/each}
				</span>
				<span class="min-w-0 flex-1 truncate text-sm {active ? 'text-signal' : 'text-ink'}">
					{option.name}
				</span>
				<!-- The active row is marked by a lamp rather than a word, matching the
				     on-air tally at the top of the panel. -->
				<span
					class="h-1.5 w-1.5 shrink-0 rounded-full {active ? 'bg-signal' : 'bg-transparent'}"
					aria-hidden="true"
				></span>
			</button>
		</li>
	{/each}
</ul>
