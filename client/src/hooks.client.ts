import { themeService } from '$lib/services/theme-service.svelte';
import type { ClientInit } from '@sveltejs/kit';

export const init: ClientInit = () => {
	themeService.initialize();
};
