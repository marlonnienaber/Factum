<script lang="ts">
	import '../app.scss';
	import favicon from '$lib/assets/favicon.svg';
	import { themeStore } from '$lib/stores/theme-store.svelte';
	import { Toaster } from 'svelte-sonner';

	let { children } = $props();
	const sonnerTheme = $derived(themeStore.isDarkTheme ? 'dark' : 'light');
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700&display=swap" rel="stylesheet" />
</svelte:head>

<Toaster position="top-center" closeButton theme={sonnerTheme} />

{@render children()}

<style lang="scss">
	:global([data-sonner-toaster] [data-sonner-toast][data-type='error']) {
		border-color: var(--danger);
	}

	:global([data-sonner-toaster] [data-sonner-toast][data-type='success']) {
		border-color: var(--success);
	}
</style>
