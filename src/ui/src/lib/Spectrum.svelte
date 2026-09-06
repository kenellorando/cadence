<script>
	// A spectrum analyser drawn from the audio actually coming out of the
	// player. This is only possible because the stream is served from the same
	// origin as the page: Web Audio refuses to expose samples from cross-origin
	// media, and would hand back silence if the audio still came from a separate
	// streaming host.
	import { theme } from './theme.svelte.js';

	let { analyser = null, playing = false, bars = 16 } = $props();

	let canvas = $state(null);

	$effect(() => {
		if (!canvas || !analyser || !playing) return;
		// Read so the loop is rebuilt when the palette changes; the colours below
		// are sampled once per loop, not once per frame.
		theme.current;

		const context = canvas.getContext('2d');
		if (!context) return;

		const reduced = window.matchMedia('(prefers-reduced-motion: reduce)');
		const styles = getComputedStyle(canvas);
		const bar = styles.getPropertyValue('--color-signal').trim() || '#fb9236';
		const quiet = styles.getPropertyValue('--color-edge').trim() || '#343d4c';

		const bins = new Uint8Array(analyser.frequencyBinCount);
		let frame;
		// A graph that builds but never produces samples looks identical to a
		// quiet track. Report it once so the difference is visible.
		let silentFrames = 0;
		let reported = false;

		// Backing store matched to the display, so bars stay crisp rather than
		// being scaled up from a smaller buffer.
		const resize = () => {
			const ratio = window.devicePixelRatio || 1;
			const width = canvas.clientWidth;
			const height = canvas.clientHeight;
			if (canvas.width !== width * ratio || canvas.height !== height * ratio) {
				canvas.width = width * ratio;
				canvas.height = height * ratio;
			}
			return { width, height, ratio };
		};

		const draw = () => {
			const { width, height, ratio } = resize();
			context.setTransform(ratio, 0, 0, ratio, 0, 0);
			context.clearRect(0, 0, width, height);

			analyser.getByteFrequencyData(bins);

			if (!reported) {
				let sum = 0;
				for (let i = 0; i < bins.length; i++) sum += bins[i];
				silentFrames = sum === 0 ? silentFrames + 1 : 0;
				if (silentFrames > 180) {
					console.warn(
						'Cadence: the audio graph is connected but returning no samples.'
					);
					reported = true;
				}
			}

			// The top of the range is mostly empty for a 192 kbps stream, so only
			// the lower bins are drawn; the rest would be a flat dead tail.
			const used = Math.floor(bins.length * 0.7);
			const count = bars;
			const gap = 2;
			const barWidth = Math.max(1, (width - gap * (count - 1)) / count);

			for (let i = 0; i < count; i++) {
				// Logarithmic spacing, so bass does not occupy most of the display.
				const from = Math.floor(Math.pow(i / count, 1.6) * used);
				const to = Math.max(from + 1, Math.floor(Math.pow((i + 1) / count, 1.6) * used));
				let peak = 0;
				for (let b = from; b < to && b < bins.length; b++) {
					if (bins[b] > peak) peak = bins[b];
				}
				// The raw figure spends most of its time low, which reads as a
				// barely moving row. Bending the curve upward trades absolute
				// accuracy -- which nobody is reading off a bar chart -- for
				// movement that actually tracks the music.
				const level = Math.pow(peak / 255, 0.55);
				const barHeight = Math.max(1, level * height);
				context.fillStyle = level > 0.04 ? bar : quiet;
				context.fillRect(i * (barWidth + gap), height - barHeight, barWidth, barHeight);
			}

			if (!reduced.matches) frame = requestAnimationFrame(draw);
		};

		draw();
		return () => cancelAnimationFrame(frame);
	});
</script>

{#if playing}
	<canvas
		bind:this={canvas}
		class="h-6 w-full"
		aria-hidden="true"
	></canvas>
{/if}
