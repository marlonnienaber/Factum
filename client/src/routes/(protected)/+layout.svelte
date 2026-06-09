<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
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
	<div class="protected-layout">
		<nav class="navbar">
			<span class="brand">Factum</span>

			<div aria-hidden="true"></div>

			<div class="account-actions">
				<span class="user-label">User #{authentication.userId}</span>
				<button class="sign-out-button" type="button" disabled={loading} onclick={handleSignOut}>
					{loading ? 'Signing out...' : 'Sign out'}
				</button>
			</div>
		</nav>

		<main class="protected-content">
			{@render children()}
		</main>
	</div>
{/if}

<style>
	.protected-layout {
		display: flex;
		min-height: 100svh;
		flex-direction: column;
	}

	.navbar {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
		align-items: center;
		gap: 1rem;
		height: 3.5rem;
		width: 100%;
		padding: 0 1rem;
		border-bottom: 1px solid #e5e5e5;
	}

	.brand {
		font-size: 1.125rem;
		font-weight: 600;
	}

	.account-actions {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		justify-self: end;
		gap: 0.75rem;
	}

	.user-label {
		color: #737373;
		font-size: 0.875rem;
	}

	.sign-out-button {
		padding: 0.375rem 0.75rem;
		border: 1px solid #d4d4d4;
		border-radius: 0.375rem;
		background: white;
		color: #171717;
		font: inherit;
		cursor: pointer;
	}

	.sign-out-button:hover {
		background: #f5f5f5;
	}

	.sign-out-button:disabled {
		cursor: default;
		opacity: 0.6;
	}

	.protected-content {
		display: flex;
		flex: 1;
	}
</style>
