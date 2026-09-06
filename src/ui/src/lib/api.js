// Every endpoint Cadence exposes to the browser, in one place. All calls are
// same-origin: the Go server serves this app and the API from the same host.

async function getJSON(path) {
	const response = await fetch(path);
	if (!response.ok) throw new Error(`${path} responded ${response.status}`);
	return response.json();
}

async function postJSON(path, body) {
	const response = await fetch(path, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
	if (!response.ok) throw new Error(`${path} responded ${response.status}`);
	return response;
}

export const getVersion = () => getJSON('/api/version');
export const getNowPlaying = () => getJSON('/api/nowplaying/metadata');
export const getAlbumArt = () => getJSON('/api/nowplaying/albumart');
export const getListenURL = () => getJSON('/api/listenurl');
export const getListeners = () => getJSON('/api/listeners');
export const getBitrate = () => getJSON('/api/bitrate');
export const getHistory = () => getJSON('/api/history');

export const search = (query) => postJSON('/api/search', { search: query }).then((r) => r.json());

// The server decodes ID as a string, so a number here is rejected with a 400.
export const requestByID = (id) => postJSON('/api/request/id', { ID: String(id) });
