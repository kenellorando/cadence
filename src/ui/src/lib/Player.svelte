<script>
	import { radio } from './radio.svelte.js';
	import Spectrum from './Spectrum.svelte';

	let audio = $state(null);
	let playing = $state(false);
	let loading = $state(false);
	let error = $state('');
	let volume = $state(0.3);

	// Artist and album read as one credit line; the dash only earns its place
	// when there is an album to separate.
	// The bar advances locally each second and is corrected against the server
	// every 15, so it moves smoothly without polling once a second.
	$effect(() => {
		const tick = setInterval(() => radio.tick(1), 1000);
		const sync = setInterval(() => radio.loadProgress(), 15000);
		return () => {
			clearInterval(tick);
			clearInterval(sync);
		};
	});

	function clock(seconds) {
		if (!Number.isFinite(seconds) || seconds < 0) seconds = 0;
		const total = Math.floor(seconds);
		const minutes = Math.floor(total / 60);
		return `${minutes}:${String(total % 60).padStart(2, '0')}`;
	}

	const progress = $derived(
		radio.progressKnown && radio.duration > 0
			? Math.min(100, (radio.elapsed / radio.duration) * 100)
			: 0
	);

	const credit = $derived(radio.album ? `${radio.artist} - ${radio.album}` : radio.artist);

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

	let analyser = $state(null);
	let audioContext = null;

	// Routes the element through an analyser so the page can draw what is
	// playing. Everything the listener hears passes through this graph, so it is
	// built in one guarded step and abandoned whole if any part of it fails --
	// a missing visualiser is a cosmetic loss, a broken audio path is not.
	function ensureAnalyser() {
		if (analyser || !audio) return;
		const AudioCtx = window.AudioContext ?? window.webkitAudioContext;
		if (!AudioCtx) return;
		try {
			const context = new AudioCtx();
			const source = context.createMediaElementSource(audio);
			const node = context.createAnalyser();
			node.fftSize = 256;
			node.smoothingTimeConstant = 0.6;
			// The default range is wide enough that ordinary music sits near the
			// bottom of it and barely registers.
			node.minDecibels = -70;
			node.maxDecibels = -25;
			// Connected straight through to the speakers: inserting the analyser
			// must not be audible.
			source.connect(node);
			node.connect(context.destination);
			audioContext = context;
			analyser = node;
			console.info('Cadence: audio visualiser connected.');
		} catch (cause) {
			// Swallowing this silently left no way to tell a browser that refused
			// to build the graph from one that built it and drew nothing.
			console.warn('Cadence: audio visualiser unavailable.', cause);
			analyser = null;
		}
	}

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
		// Built before the source is set. Reordering this on a theory about one
		// browser broke it on the browser where it already worked, so it stays
		// the way that was observed to work.
		ensureAnalyser();
		// A context created before a gesture starts suspended, and once the
		// element is routed through the graph a suspended context means silence.
		audioContext?.resume().catch(() => {});
		// A unique URL per attempt. Playing the same src again lets the browser
		// resume from its media cache, which restarts the audio wherever that
		// cache begins -- potentially a long way behind live -- instead of
		// opening a fresh connection and joining the broadcast where it is now.
		audio.src = `${radio.streamURL}?t=${Date.now()}`;
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
		class="art-wash pointer-events-none fixed inset-0 -z-10 bg-cover bg-center blur-2xl saturate-50"
		style="background-image: url({radio.art})"
	></div>
{/if}
<div aria-hidden="true" class="pointer-events-none fixed inset-0 -z-10 bg-surface/75"></div>

<section class="panel">
	<!-- Header strip: the on-air tally. -->
	<div class="flex items-center gap-3 border-b border-edge px-4 py-2.5 sm:px-5">
		{#if radio.connected}
			<span class="tally" aria-hidden="true"></span>
			<span class="label text-onair">On air</span>
		{:else}
			<span class="tally-off" aria-hidden="true"></span>
			<span class="label">Offline</span>
		{/if}
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
					class="absolute bottom-2 left-2 flex h-6 items-end gap-[2px] bg-surface-inset/85 px-2 py-1.5"
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
				<p class="label">Now playing</p>
				<h1
					class="mt-1 truncate font-display text-2xl font-medium text-ink sm:text-[1.7rem]"
					title={radio.title}
				>
					{radio.title}
				</h1>
				<p class="mt-0.5 truncate text-base text-ink-dim" title={credit}>{credit}</p>
			</div>

			<!-- Position through the track. The audio source reports how much is
			     left; this is that, counted forward. The row is always laid out,
			     even while the position is unknown between tracks, so nothing
			     below it shifts. -->
			<div class="mt-4 flex items-center gap-3">
				<span class="font-mono text-[0.7rem] text-ink-faint tabular-nums">
					{radio.progressKnown ? clock(radio.elapsed) : '--:--'}
				</span>
				<div
					class="h-[3px] flex-1 overflow-hidden rounded-full bg-edge"
					role="progressbar"
					aria-label="Track position"
					aria-valuemin="0"
					aria-valuemax={Math.round(radio.duration)}
					aria-valuenow={radio.progressKnown ? Math.round(radio.elapsed) : 0}
				>
					<div
						class="h-full bg-signal transition-[width] duration-1000 ease-linear"
						style="width: {radio.progressKnown ? progress : 0}%"
					></div>
				</div>
				<span class="font-mono text-[0.7rem] text-ink-faint tabular-nums">
					{radio.progressKnown ? clock(radio.duration) : '--:--'}
				</span>
			</div>

			<div class="mt-auto flex flex-wrap items-center gap-4 pt-5">
				<button
					onclick={toggle}
					disabled={!radio.connected || loading}
					aria-label={playing ? 'Pause stream' : 'Play stream'}
					class="grid h-11 w-14 shrink-0 place-items-center rounded-[3px] border border-edge-light bg-surface text-signal transition enabled:hover:border-signal enabled:hover:bg-surface-inset enabled:active:translate-y-px disabled:cursor-not-allowed disabled:text-ink-faint disabled:opacity-50"
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
					<dd class="font-mono text-sm text-ink tabular-nums">
						{radio.listeners === -1 ? '—' : radio.listeners}
					</dd>
				</div>
				<div>
					<dt class="label">Bitrate</dt>
					<dd class="font-mono text-sm text-ink tabular-nums">
						{radio.bitrate ? `${radio.bitrate}k` : '—'}
					</dd>
				</div>
				{#if playing && analyser}
					<div class="min-w-32 flex-1 self-end pb-1">
						<Spectrum {analyser} {playing} />
					</div>
				{/if}
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
				<p class="mt-3 font-mono text-xs text-ink-faint">No source currently connected.</p>
			{/if}
		</div>
	</div>

</section>
