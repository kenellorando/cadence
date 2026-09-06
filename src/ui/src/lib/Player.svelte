<script>
	import { radio } from './radio.svelte.js';

	let audio = $state(null);
	let playing = $state(false);
	let loading = $state(false);
	let error = $state('');
	let volume = $state(0.3);

	$effect(() => {
		const stored = Number(localStorage.getItem('volumeKey'));
		if (Number.isFinite(stored) && stored > 0) volume = stored;
	});

	$effect(() => {
		if (audio) audio.volume = volume;
	});

	// Space is the transport control people expect, but only when they are not
	// typing into the search field.
	$effect(() => {
		const onKey = (event) => {
			if (event.code !== 'Space') return;
			const tag = event.target?.tagName;
			if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'BUTTON') return;
			event.preventDefault();
			toggle();
		};
		window.addEventListener('keydown', onKey);
		return () => window.removeEventListener('keydown', onKey);
	});

	function setVolume(event) {
		volume = Number(event.currentTarget.value);
		localStorage.setItem('volumeKey', String(volume));
	}

	async function toggle() {
		if (!audio || !radio.connected) return;
		error = '';

		if (playing) {
			// Clearing the source disconnects rather than buffering on in the
			// background while paused.
			audio.pause();
			audio.removeAttribute('src');
			audio.load();
			playing = false;
			return;
		}

		loading = true;
		audio.src = radio.streamURL;
		audio.load();
		try {
			await audio.play();
			playing = true;
		} catch {
			playing = false;
			// Silent failure is what made the old player feel broken: an
			// unreachable stream looks identical to a dead button unless it is
			// said out loud.
			error = 'Playback failed. The stream may be unreachable from here.';
		} finally {
			loading = false;
		}
	}
</script>

<!-- Room light from the artwork: warm, low and desaturated, so it reads as
     illumination on the panel rather than a coloured glow behind a card. -->
{#if radio.art}
	<div
		aria-hidden="true"
		class="pointer-events-none fixed inset-0 -z-10 bg-cover bg-center opacity-[0.14] blur-2xl saturate-50"
		style="background-image: url({radio.art})"
	></div>
{/if}
<div aria-hidden="true" class="pointer-events-none fixed inset-0 -z-10 bg-panel/70"></div>

<section class="panel mt-5">
	<!-- Header strip: tally lamp, station name, firmware-style version. -->
	<div class="flex items-center gap-3 border-b border-edge px-4 py-2.5 sm:px-5">
		{#if radio.connected}
			<span class="tally" aria-hidden="true"></span>
			<span class="label text-signal">On air</span>
		{:else}
			<span class="tally-off" aria-hidden="true"></span>
			<span class="label">Offline</span>
		{/if}
		<span class="ml-auto font-mono text-[0.68rem] tracking-wide text-bone-faint tabular-nums">
			{radio.version}
		</span>
	</div>

	<div class="flex flex-col gap-5 p-4 sm:flex-row sm:gap-6 sm:p-5">
		<div class="relative shrink-0">
			<div class="well aspect-square w-full sm:h-52 sm:w-52">
				{#if radio.art}
					<img
						src={radio.art}
						alt="Album art for {radio.title}"
						class="h-full w-full object-cover"
					/>
				{:else}
					<div class="flex h-full w-full items-center justify-center">
						<span class="label">No art</span>
					</div>
				{/if}
			</div>

			{#if playing}
				<div
					class="absolute bottom-2 left-2 flex h-6 items-end gap-[2px] bg-panel-well/85 px-2 py-1.5"
					aria-hidden="true"
				>
					<span class="level-bar h-2.5 w-[2px] bg-level" style="animation-delay:0ms"></span>
					<span class="level-bar h-3.5 w-[2px] bg-level" style="animation-delay:130ms"></span>
					<span class="level-bar h-2 w-[2px] bg-signal" style="animation-delay:260ms"></span>
				</div>
			{/if}
		</div>

		<div class="flex min-w-0 flex-1 flex-col">
			<audio bind:this={audio} preload="none"></audio>

			<!-- Announced to screen readers when the track changes, since the
			     change arrives over the event stream rather than a navigation. -->
			<div aria-live="polite" aria-atomic="true">
				<p class="label">Artist</p>
				<p class="truncate text-lg text-bone-dim" title={radio.artist}>{radio.artist}</p>

				<hr class="my-2.5 border-edge" />

				<p class="label">Now playing</p>
				<h1
					class="truncate font-display text-2xl font-semibold text-bone sm:text-[1.7rem]"
					title={radio.title}
				>
					{radio.title}
				</h1>
				{#if radio.album}
					<p class="truncate text-sm text-bone-faint" title={radio.album}>{radio.album}</p>
				{/if}
			</div>

			<div class="mt-auto flex flex-wrap items-center gap-4 pt-5">
				<button
					onclick={toggle}
					disabled={!radio.connected || loading}
					aria-label={playing ? 'Pause stream' : 'Play stream'}
					class="grid h-11 w-14 shrink-0 place-items-center rounded-[3px] border border-edge-light bg-panel text-signal transition enabled:hover:border-signal enabled:hover:bg-panel-well enabled:active:translate-y-px disabled:cursor-not-allowed disabled:text-bone-faint disabled:opacity-50"
				>
					{#if loading}
						<span
							class="h-4 w-4 animate-spin rounded-full border border-edge-light border-t-signal"
						></span>
					{:else if playing}
						<span class="text-sm tracking-[0.15em]">❚❚</span>
					{:else}
						<span class="text-base">▶</span>
					{/if}
				</button>

				<div class="flex min-w-40 flex-1 items-center gap-3">
					<span class="label shrink-0">Vol</span>
					<input
						type="range"
						min="0"
						max="1"
						step="0.01"
						value={volume}
						oninput={setVolume}
						aria-label="Volume"
						class="fader w-full"
						style="--fill: {volume * 100}%"
					/>
				</div>
			</div>

			<!-- Telemetry, in the mono face with tabular figures so the numbers
			     hold their columns as they change. -->
			<dl class="mt-5 flex flex-wrap gap-x-8 gap-y-2 border-t border-edge pt-4">
				<div>
					<dt class="label">Listeners</dt>
					<dd class="font-mono text-sm text-bone tabular-nums">
						{radio.listeners === -1 ? '—' : radio.listeners}
					</dd>
				</div>
				<div>
					<dt class="label">Bitrate</dt>
					<dd class="font-mono text-sm text-bone tabular-nums">
						{radio.bitrate ? `${radio.bitrate}k` : '—'}
					</dd>
				</div>
				{#if radio.connected}
					<div class="ml-auto self-end">
						<a
							href={radio.streamURL}
							class="label underline decoration-edge-light underline-offset-4 transition hover:text-signal"
						>
							Direct stream
						</a>
					</div>
				{/if}
			</dl>

			{#if error}
				<p class="mt-3 font-mono text-xs text-peak">{error}</p>
			{:else if !radio.connected}
				<p class="mt-3 font-mono text-xs text-bone-faint">No source connected to the broadcaster.</p>
			{/if}
		</div>
	</div>
</section>
