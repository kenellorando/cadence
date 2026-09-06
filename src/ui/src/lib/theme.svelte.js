// Theme selection. Each theme only redefines tokens, so no component needs to
// know which one is active.
export const THEMES = [
	{
		id: 'chicago',
		name: 'Chicago Evening',
		logo: 'Cadence',
		swatch: ['#101319', '#ece7de', '#fb9236', '#e2452f']
	},
	{
		id: 'bartender',
		name: 'Cyberpunk Bartender',
		logo: 'Cadence',
		swatch: ['#0c0a12', '#e9e6f5', '#00ffff', '#ff69b4']
	},
	{
		id: 'lightmage',
		name: 'Light Mage',
		logo: 'Cadence',
		swatch: ['#faf7ef', '#302c20', '#a8760f', '#dd5140']
	},
	{
		id: 'electromaster',
		name: 'Electromaster',
		// Rendered in a Japanese hand; the same name, transliterated.
		logo: 'ケイデンス',
		swatch: ['#ece5d5', '#1c1f26', '#0b5fd4', '#c9372c']
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
