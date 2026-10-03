import type { Theme } from '$lib/types/persists/theme';

let theme = $state<Theme>('light');

function setTheme(newTheme: Theme): void {
	theme = newTheme;
}

export const themeStore = {
	get theme(): Theme {
		return theme;
	},
	get isDarkTheme(): boolean {
		return theme === 'dark';
	},
	setTheme
};
