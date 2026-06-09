<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { LogOutIcon } from '@lucide/svelte';
	import { Button } from '$lib/components/shadcn/button/index.js';
	import { authenticationStore } from '$lib/stores/authentication-store.svelte';

	const { children } = $props();
	const authentication = $derived(authenticationStore.authentication);
	const attemptedToFetchAuthentication = $derived(authenticationStore.fetchAttempted);
	const authenticationNotLoading = $derived(!authenticationStore.loading);
	const loading = $derived(authenticationStore.loading);

	$effect(() => {
		if (!authentication && authenticationNotLoading) {
			if (attemptedToFetchAuthentication) {
				void goto(resolve('/sign-in'));
			} else {
				void authenticationStore.fetchCurrentAuthenticationIfNecessary();
			}
		}
	});

	async function handleSignOut(): Promise<void> {
		await authenticationStore.signOut();
		await goto(resolve('/sign-in'));
	}
</script>

{#if authentication}
	<div class="flex min-h-svh flex-col">
		<nav class="grid h-14 w-full grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-4 border-b px-4">
			<span class="text-lg font-semibold">Factum</span>

			<div aria-hidden="true"></div>

			<div class="flex items-center justify-end gap-3 justify-self-end">
				<span class="text-sm text-muted-foreground">User #{authentication.userId}</span>
				<Button type="button" variant="outline" size="sm" disabled={loading} onclick={handleSignOut}>
					<LogOutIcon />
					{loading ? 'Signing out...' : 'Sign out'}
				</Button>
			</div>
		</nav>

		<main class="flex flex-1">
			{@render children()}
		</main>
	</div>
{/if}
