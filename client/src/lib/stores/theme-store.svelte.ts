enum Theme {
	Light = 'light',
	Dark = 'dark'
}

let theme = $state(Theme.Light);
let observer: MutationObserver | undefined;

function readThemeFromDocument() {
	if (typeof document === 'undefined') return;
	theme = document.documentElement.dataset.theme === Theme.Dark ? Theme.Dark : Theme.Light;
}

function initialize() {
	if (observer !== undefined || typeof document === 'undefined') return;

	readThemeFromDocument();

	observer = new MutationObserver(readThemeFromDocument);
	observer.observe(document.documentElement, {
		attributes: true,
		attributeFilter: ['data-theme']
	});
}

function toggleTheme(): void {
	if (typeof document === 'undefined') return;
	document.documentElement.dataset.theme = theme === Theme.Dark ? Theme.Light : Theme.Dark;
}

initialize();

export const themeStore = {
	get isDarkTheme() {
		return theme === Theme.Dark;
	},
	toggleTheme
};
