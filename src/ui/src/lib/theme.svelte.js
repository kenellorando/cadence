// Theme selection. Each theme only redefines tokens, so no component needs to
// know which one is active.
export const THEMES = [
	{
		id: 'transmitter',
		name: 'Transmitter',
		note: 'Anodized black, amber signal',
		swatch: ['#17150f', '#e8e2d2', '#e0912f', '#e2452f']
	},
	{
		id: 'dial',
		name: 'Dial card',
		note: 'Ivory stock, red pointer',
		swatch: ['#efe9db', '#1f1b14', '#a8452a', '#c0281a']
	},
	{
		id: 'tape',
		name: 'Tape',
		note: 'Bone card, teal spot',
		swatch: ['#e6e2d6', '#14130d', '#1d5b66', '#b0231b']
	}
];

const STORAGE_KEY = 'themeKey';

class Theme {
	current = $state('transmitter');

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

	apply() {
		document.documentElement.setAttribute('data-theme', this.current);
	}
}

export const theme = new Theme();
