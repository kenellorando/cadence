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

	function setVolume(event) {
		volume = Number(event.currentTarget.value);
		localStorage.setItem('volumeKey', String(volume));
	}

	async function toggle() {
		if (!audio) return;
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
			// unreachable stream or a codec the browser refuses looks identical to
			// a dead button unless it is said out loud.
			error = 'Playback failed. The stream may be unreachable or unsupported here.';
		} finally {
			loading = false;
		}
	}
</script>

<!-- Ambient wash pulled from the album art, sitting behind the whole page. -->
{#if radio.art}
	<div
		aria-hidden="true"
		class="pointer-events-none fixed inset-0 -z-10 bg-cover bg-center opacity-25 blur-3xl saturate-150 transition-[background-image] duration-1000"
		style="background-image: url({radio.art})"
	></div>
{/if}
<div aria-hidden="true" class="pointer-events-none fixed inset-0 -z-10 bg-gradient-to-b from-surface/40 via-surface/80 to-surface"></div>

<section
	class="mt-8 flex flex-col items-center gap-7 rounded-3xl border border-white/10 bg-white/5 p-6 backdrop-blur-xl sm:flex-row sm:items-end sm:gap-8 sm:p-8"
>
	<div class="relative shrink-0">
		<div
			class="h-52 w-52 overflow-hidden rounded-2xl bg-surface-raised shadow-2xl shadow-black/60 ring-1 ring-white/10 sm:h-56 sm:w-56"
		>
			{#if radio.art}
				<img
					src={radio.art}
					alt="Album art for {radio.title}"
					class="h-full w-full object-cover"
				/>
			{:else}
				<div class="flex h-full w-full items-center justify-center text-4xl text-white/15">♪</div>
			{/if}
		</div>

		{#if playing}
			<div
				class="absolute bottom-3 left-3 flex h-8 items-end gap-[3px] rounded-full bg-black/60 px-3 py-2 backdrop-blur"
				aria-hidden="true"
			>
				<span class="bar h-3 w-[3px] rounded-full bg-accent" style="animation-delay:0ms"></span>
				<span class="bar h-4 w-[3px] rounded-full bg-accent" style="animation-delay:150ms"></span>
				<span class="bar h-3 w-[3px] rounded-full bg-accent" style="animation-delay:300ms"></span>
			</div>
		{/if}
	</div>

	<div class="flex min-w-0 flex-1 flex-col items-center text-center sm:items-start sm:text-left">
		<audio bind:this={audio} preload="none"></audio>

		<div class="flex items-center gap-2 text-[11px] font-semibold tracking-[0.14em] uppercase">
			{#if radio.connected}
				<span class="onair-dot h-2 w-2 rounded-full bg-emerald-400"></span>
				<span class="text-emerald-400">On air</span>
			{:else}
				<span class="h-2 w-2 rounded-full bg-white/25"></span>
				<span class="text-white/40">Offline</span>
			{/if}
		</div>

		<h1 class="mt-2 max-w-full truncate text-3xl font-semibold text-white" title={radio.title}>
			{radio.title}
		</h1>
		<p class="max-w-full truncate text-lg text-white/55" title={radio.artist}>{radio.artist}</p>
		{#if radio.album}
			<p class="max-w-full truncate text-sm text-white/35" title={radio.album}>{radio.album}</p>
		{/if}

		<div class="mt-6 flex w-full flex-col items-center gap-5 sm:flex-row sm:items-center">
			<button
				onclick={toggle}
				disabled={!radio.connected || loading}
				aria-label={playing ? 'Pause stream' : 'Play stream'}
				class="grid h-14 w-14 shrink-0 place-items-center rounded-full bg-gradient-to-br from-accent to-accent-strong text-xl text-white shadow-lg shadow-accent-strong/30 transition enabled:hover:scale-105 enabled:hover:shadow-accent-strong/50 enabled:active:scale-95 disabled:cursor-not-allowed disabled:opacity-30 disabled:shadow-none"
			>
				{#if loading}
					<span class="h-5 w-5 animate-spin rounded-full border-2 border-white/30 border-t-white"></span>
				{:else if playing}
					<span class="text-2xl leading-none">❚❚</span>
				{:else}
					<span class="ml-1 text-2xl leading-none">▶</span>
				{/if}
			</button>

			<div class="flex w-full max-w-56 items-center gap-3">
				<span class="text-white/35" aria-hidden="true">🔈</span>
				<input
					type="range"
					min="0"
					max="1"
					step="0.006"
					value={volume}
					oninput={setVolume}
					aria-label="Volume"
					class="volume w-full"
					style="--fill: {volume * 100}%"
				/>
			</div>
		</div>

		<div class="mt-5 flex flex-wrap items-center justify-center gap-2 sm:justify-start">
			<span
				class="rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs text-white/60"
				title="Listeners tuned in right now"
			>
				{radio.listeners === -1 ? 'N/A' : radio.listeners}
				{radio.listeners === 1 ? 'listener' : 'listeners'}
			</span>
			{#if radio.connected}
				<a
					href={radio.streamURL}
					class="rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs text-white/60 transition hover:border-accent/50 hover:text-white"
				>
					Direct stream ↗
				</a>
			{/if}
		</div>

		{#if error}
			<p class="mt-4 text-sm text-rose-400">{error}</p>
		{:else if !radio.connected}
			<p class="mt-4 text-sm text-white/40">Disconnected from server.</p>
		{/if}
	</div>
</section>
