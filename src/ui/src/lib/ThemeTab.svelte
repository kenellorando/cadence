<script>
	import { theme, THEMES } from './theme.svelte.js';
</script>

<ul>
	{#each THEMES as option (option.id)}
		<li>
			<button
				onclick={() => theme.set(option.id)}
				aria-pressed={theme.current === option.id}
				class="flex w-full items-center gap-4 border-b border-edge/60 px-4 py-3 text-left transition hover:bg-surface-inset/60 sm:px-5"
			>
				<!-- Literal values, not tokens: a swatch has to show its own theme's
				     colours while a different theme is on screen. -->
				<span class="flex shrink-0 overflow-hidden rounded-[2px] border border-edge">
					{#each option.swatch as colour (colour)}
						<span class="h-6 w-3" style="background: {colour}"></span>
					{/each}
				</span>
				<span class="min-w-0 flex-1">
					<span class="block truncate text-sm text-ink">{option.name}</span>
					<span class="block truncate text-xs text-ink-faint">{option.note}</span>
				</span>
				<span class="label shrink-0 {theme.current === option.id ? 'text-signal' : ''}">
					{theme.current === option.id ? 'Active' : 'Select'}
				</span>
			</button>
		</li>
	{/each}
</ul>
