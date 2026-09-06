// Theme selection. Each theme only redefines tokens, so no component needs to
// know which one is active.
export const THEMES = [
	{
		id: 'chicago',
		name: 'Chicago Evening',
		note: 'Dusk sky, sodium street light',
		logo: 'Cadence',
		swatch: ['#101319', '#ece7de', '#fb9236', '#e2452f']
	},
	{
		id: 'bartender',
		name: 'Cyberpunk Bartender',
		note: 'Cyan and hotpink after dark',
		logo: 'Cadence',
		swatch: ['#0c0a12', '#e9e6f5', '#00ffff', '#ff69b4']
	},
	{
		id: 'electromaster',
		name: 'Electromaster',
		note: 'Modern blue, mid-tone ground',
		// Rendered in a Japanese hand; the same name, transliterated.
		logo: 'ケイデンス',
		swatch: ['#525d70', '#f4f7fc', '#62a0ff', '#ff6b5a']
	}
];

const STORAGE_KEY = 'themeKey';

class Theme {
	current = $state('chicago');

	load() {
		try {
			const stored = localStorage.getItem(STORAGE_KEY);
			if (THEMES.some((t) => t.id === stored)) this.current = stored;
		} catch {
			// A browser with storage blocked still gets the default.
		}
		this.apply();
	}

	set(id) {
		if (!THEMES.some((t) => t.id === id)) return;
		this.current = id;
		try {
			localStorage.setItem(STORAGE_KEY, id);
		} catch {
			// Not persisting is survivable; the change still applies this session.
		}
		this.apply();
	}

	// The wordmark is not the same string in every theme, so it travels with the
	// theme rather than being hardcoded in the footer.
	get logo() {
		return THEMES.find((t) => t.id === this.current)?.logo ?? 'Cadence';
	}

	apply() {
		document.documentElement.setAttribute('data-theme', this.current);
	}
}

export const theme = new Theme();
