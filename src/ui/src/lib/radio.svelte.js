import * as api from './api.js';

// Live radio state, shared by every component. The server pushes changes over
// SSE; the REST endpoints are only used to fill in the initial page load and to
// fetch the things SSE announces but does not carry (art, history).
class Radio {
	title = $state('-');
	artist = $state('-');
	album = $state('');
	art = $state('');
	listeners = $state(-1);
	bitrate = $state(0);
	listenURL = $state('');
	version = $state('-');
	history = $state([]);
	searchResults = $state([]);
	searchStatus = $state('');
	// Null until the first check, so the search pane can stay quiet rather than
	// claiming an empty library before it knows.
	indexing = $state(null);
	tracks = $state(0);
	// Track position. The server is asked occasionally; the page counts the
	// seconds in between, so the bar moves smoothly rather than in steps.
	elapsed = $state(0);
	duration = $state(0);
	progressKnown = $state(false);

	// "-/-" is what the server reports when Icecast has no source connected.
	get connected() {
		return this.listenURL !== '' && this.listenURL !== '-/-';
	}

	// The server hands back a same-origin path, so it is played as-is. Building
	// a URL here is what used to drop the port and guess at the scheme.
	get streamURL() {
		return this.connected ? this.listenURL : '';
	}

	async loadNowPlaying() {
		try {
			const data = await api.getNowPlaying();
			this.title = data.Title;
			this.artist = data.Artist;
			this.album = data.Album ?? '';
		} catch {
			this.title = '-';
			this.artist = '-';
		}
	}

	async loadArt() {
		try {
			const data = await api.getAlbumArt();
			this.art = data?.Picture ? `data:image/jpeg;base64,${data.Picture}` : '';
		} catch {
			this.art = '';
		}
	}

	async loadListenURL() {
		try {
			this.listenURL = (await api.getListenURL()).ListenURL;
		} catch {
			this.listenURL = '';
		}
	}

	async loadListeners() {
		try {
			this.listeners = (await api.getListeners()).Listeners;
		} catch {
			this.listeners = -1;
		}
	}

	async loadBitrate() {
		try {
			this.bitrate = (await api.getBitrate()).Bitrate;
		} catch {
			this.bitrate = 0;
		}
	}

	async loadHistory() {
		try {
			this.history = (await api.getHistory()) ?? [];
		} catch {
			this.history = [];
		}
	}

	async loadLibrary() {
		try {
			const status = await api.getLibrary();
			this.indexing = status.Indexing;
			this.tracks = status.Tracks;
		} catch {
			this.indexing = false;
		}
	}

	async loadProgress() {
		try {
			const progress = await api.getProgress();
			this.elapsed = progress.Elapsed;
			this.duration = progress.Duration;
			this.progressKnown = progress.Known;
		} catch {
			this.progressKnown = false;
		}
	}

	// Advances the local count between server readings. Stops at the duration
	// rather than running past it while waiting for the next track.
	tick(seconds) {
		if (!this.progressKnown) return;
		this.elapsed = Math.min(this.elapsed + seconds, this.duration);
	}

	async loadVersion() {
		try {
			this.version = (await api.getVersion()).Version;
		} catch {
			this.version = '(N/A)';
		}
	}

	async runSearch(query) {
		try {
			const results = (await api.search(query)) ?? [];
			this.searchResults = results;
			this.searchStatus = `Results: ${results.length}`;
		} catch {
			this.searchResults = [];
			this.searchStatus = 'Error. Could not execute search.';
		}
	}

	async request(id) {
		try {
			await api.requestByID(id);
			this.searchStatus = 'Request accepted!';
		} catch {
			this.searchStatus = 'Sorry, your request was not accepted. You may be rate limited.';
		}
	}

	// Subscribes to the server event stream, reconnecting if it drops. Returns a
	// teardown function for the caller's $effect.
	connect() {
		let source;
		let retry;
		let watchdog;
		let closed = false;

		// The server sends a keepalive comment every 20 seconds. Comments do not
		// surface as events, but any traffic resets the browser's read state, so a
		// stream that goes quiet for much longer than that has died in a way the
		// connection itself will not report. Only the client can notice this: a
		// half-open connection looks identical to an idle one from the server side.
		const silenceLimit = 50000;

		const reconnect = () => {
			if (closed) return;
			source?.close();
			clearTimeout(retry);
			retry = setTimeout(open, 5000);
		};

		const kick = () => {
			clearTimeout(watchdog);
			watchdog = setTimeout(reconnect, silenceLimit);
		};

		const open = () => {
			if (closed) return;
			source = new EventSource('/api/radiodata/sse');

			// Events only fire on change, so a client that reconnects mid-song would
			// otherwise keep showing whatever was playing when it dropped.
			source.onopen = () => {
				kick();
				this.loadAll();
			};

			source.addEventListener('title', (event) => {
				kick();
				this.title = event.data;
				// A new track restarts the clock, so read it rather than waiting
				// for the next poll to notice.
				this.elapsed = 0;
				this.loadProgress();
				// The stream announces the change; the art and the rest of the
				// metadata still have to be fetched.
				this.loadArt();
				this.loadNowPlaying();
			});
			source.addEventListener('artist', (event) => {
				kick();
				this.artist = event.data;
			});
			source.addEventListener('listeners', (event) => {
				kick();
				this.listeners = Number(event.data);
			});
			source.addEventListener('listenurl', (event) => {
				kick();
				this.listenURL = event.data;
			});
			source.addEventListener('history', () => {
				kick();
				this.loadHistory();
			});
			// The library finished being read; results that were empty a moment
			// ago are now available.
			source.addEventListener('library', () => {
				kick();
				this.loadLibrary();
				this.runSearch('');
			});

			source.onerror = reconnect;
		};

		open();
		return () => {
			closed = true;
			clearTimeout(retry);
			clearTimeout(watchdog);
			source?.close();
		};
	}

	async loadAll() {
		await Promise.all([
			this.loadNowPlaying(),
			this.loadArt(),
			this.loadListenURL(),
			this.loadListeners(),
			this.loadBitrate(),
			this.loadHistory(),
			this.loadVersion(),
			this.loadLibrary(),
			this.loadProgress(),
			this.runSearch('')
		]);
	}
}

export const radio = new Radio();
