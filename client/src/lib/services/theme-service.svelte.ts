import { themeStore } from '$lib/stores/theme-store.svelte';
import { themeSchema, type Theme } from '$lib/types/persists/theme';
import { toast } from 'svelte-sonner';

const THEME_STORAGE_KEY = 'theme';

function initialize(): void {
	if (typeof document === 'undefined') return;
	const systemThemeDarkPreference = window.matchMedia('(prefers-color-scheme: dark)');
	const systemPrefersDarkTheme = systemThemeDarkPreference.matches;
	const storedTheme = readThemeFromLocalStorage();
	const fallbackTheme = systemPrefersDarkTheme ? 'dark' : 'light';
	applyTheme(storedTheme ?? fallbackTheme);
	systemThemeDarkPreference.addEventListener('change', applySystemTheme);
}

function toggleTheme(): void {
	if (typeof document === 'undefined') return;
	const nextTheme: Theme = themeStore.theme === 'dark' ? 'light' : 'dark';
	applyTheme(nextTheme);
	persistThemeToLocalStorage(nextTheme);
}

function applyTheme(theme: Theme): void {
	document.documentElement.dataset.theme = theme;
	themeStore.setTheme(theme);
}

function applySystemTheme(systemThemeDarkPreferenceChangeEvent: MediaQueryListEvent): void {
	if (readThemeFromLocalStorage() !== undefined) return;
	const systemPrefersDarkTheme = systemThemeDarkPreferenceChangeEvent.matches;
	applyTheme(systemPrefersDarkTheme ? 'dark' : 'light');
}

function readThemeFromLocalStorage(): Theme | undefined {
	try {
		const storedThemeParseResult = themeSchema.safeParse(localStorage.getItem(THEME_STORAGE_KEY));
		return storedThemeParseResult.success ? storedThemeParseResult.data : undefined;
	} catch {
		return undefined;
	}
}

function persistThemeToLocalStorage(theme: Theme): void {
	try {
		localStorage.setItem(THEME_STORAGE_KEY, theme);
	} catch {
		toast.warning(
			'Your browser does not support saving the selected theme. You may need to select it again the next time you use Factum.'
		);
	}
}

export const themeService = {
	initialize,
	toggleTheme
};
