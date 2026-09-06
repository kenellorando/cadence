<script>
	// A spectrum analyser drawn from the audio actually coming out of the
	// player. This is only possible because the stream is served from the same
	// origin as the page: Web Audio refuses to expose samples from cross-origin
	// media, and would hand back silence if the audio still came from a separate
	// streaming host.
	let { analyser = null, playing = false } = $props();

	let canvas = $state(null);

	$effect(() => {
		if (!canvas || !analyser || !playing) return;

		const context = canvas.getContext('2d');
		if (!context) return;

		const reduced = window.matchMedia('(prefers-reduced-motion: reduce)');
		const styles = getComputedStyle(canvas);
		const bar = styles.getPropertyValue('--color-signal').trim() || '#fb9236';
		const quiet = styles.getPropertyValue('--color-edge').trim() || '#343d4c';

		const bins = new Uint8Array(analyser.frequencyBinCount);
		let frame;

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

			// The top of the range is mostly empty for a 192 kbps stream, so only
			// the lower bins are drawn; the rest would be a flat dead tail.
			const used = Math.floor(bins.length * 0.7);
			const count = 40;
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
				const level = peak / 255;
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
		class="mt-4 h-8 w-full"
		aria-hidden="true"
	></canvas>
{/if}
