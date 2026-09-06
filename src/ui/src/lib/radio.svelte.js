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
	listenURL = $state('');
	version = $state('-');
	history = $state([]);
	searchResults = $state([]);
	searchStatus = $state('');

	// "-/-" is what the server reports when Icecast has no source connected.
	get connected() {
		return this.listenURL !== '' && this.listenURL !== '-/-';
	}

	get streamURL() {
		return this.connected ? `${location.protocol}//${this.listenURL}` : '';
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

	async loadHistory() {
		try {
			this.history = (await api.getHistory()) ?? [];
		} catch {
			this.history = [];
		}
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
		let closed = false;

		const open = () => {
			source = new EventSource('/api/radiodata/sse');

			source.addEventListener('title', (event) => {
				this.title = event.data;
				// The stream announces the change; the art and the rest of the
				// metadata still have to be fetched.
				this.loadArt();
				this.loadNowPlaying();
			});
			source.addEventListener('artist', (event) => {
				this.artist = event.data;
			});
			source.addEventListener('listeners', (event) => {
				this.listeners = Number(event.data);
			});
			source.addEventListener('listenurl', (event) => {
				this.listenURL = event.data;
			});
			source.addEventListener('history', () => {
				this.loadHistory();
			});

			source.onerror = () => {
				source.close();
				if (!closed) retry = setTimeout(open, 10000);
			};
		};

		open();
		return () => {
			closed = true;
			clearTimeout(retry);
			source?.close();
		};
	}

	async loadAll() {
		await Promise.all([
			this.loadNowPlaying(),
			this.loadArt(),
			this.loadListenURL(),
			this.loadListeners(),
			this.loadHistory(),
			this.loadVersion(),
			this.runSearch('')
		]);
	}
}

export const radio = new Radio();
