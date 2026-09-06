<script>
	import { radio } from './radio.svelte.js';

	let audio;
	let playing = $state(false);
	let volume = $state(0.3);

	$effect(() => {
		// Restore the listener's last volume, falling back to 30%.
		const stored = Number(localStorage.getItem('volumeKey'));
		volume = Number.isFinite(stored) && stored > 0 ? stored : 0.3;
	});

	$effect(() => {
		if (audio) audio.volume = volume;
	});

	function setVolume(event) {
		volume = Number(event.currentTarget.value);
		localStorage.setItem('volumeKey', String(volume));
	}

	function toggle() {
		if (!audio) return;
		if (playing) {
			// Dropping the source disconnects from the stream rather than buffering
			// it in the background while paused.
			audio.pause();
			audio.removeAttribute('src');
			audio.load();
			playing = false;
			return;
		}
		audio.src = radio.streamURL;
		audio.load();
		audio.play().then(
			() => (playing = true),
			() => (playing = false)
		);
	}
</script>

<section class="flex flex-col items-center gap-6 py-8 sm:flex-row sm:items-center sm:gap-8">
	<div
		class="h-[50vw] max-h-60 w-[50vw] max-w-60 shrink-0 overflow-hidden rounded-lg border border-neutral-300 bg-neutral-100 dark:border-neutral-700 dark:bg-neutral-800"
	>
		{#if radio.art}
			<img src={radio.art} alt="Album art for {radio.title}" class="h-full w-full object-cover" />
		{:else}
			<div class="flex h-full w-full items-center justify-center text-neutral-400">No art</div>
		{/if}
	</div>

	<div class="flex min-w-0 flex-col items-center text-center sm:items-start sm:text-left">
		<audio bind:this={audio} preload="none"></audio>

		<h1 class="max-w-full truncate text-2xl font-medium">{radio.title}</h1>
		<h2 class="max-w-full truncate text-xl text-neutral-600 dark:text-neutral-400">{radio.artist}</h2>

		<button
			onclick={toggle}
			disabled={!radio.connected}
			aria-label={playing ? 'Pause stream' : 'Play stream'}
			class="my-5 h-11 w-20 rounded-lg border border-neutral-400 text-lg transition enabled:hover:bg-neutral-100 disabled:cursor-not-allowed disabled:opacity-40 dark:border-neutral-600 dark:enabled:hover:bg-neutral-800"
		>
			{playing ? '⏸' : '►'}
		</button>

		<input
			type="range"
			min="0"
			max="1"
			step="0.006"
			value={volume}
			oninput={setVolume}
			aria-label="Volume"
			class="w-48"
		/>

		<p class="mt-4 text-sm">
			{#if radio.connected}
				Connected: <a class="underline" href={radio.streamURL}>{radio.streamURL}</a>
			{:else}
				Disconnected from server.
			{/if}
		</p>
		<p class="text-sm">
			Current Listeners: {radio.listeners === -1 ? 'N/A' : radio.listeners}
		</p>
	</div>
</section>
