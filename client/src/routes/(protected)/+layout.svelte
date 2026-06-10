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
	<div class="frame">
		<nav class="navbar">
			<span class="brand">Factum</span>
			<div class="account-actions">
				<span class="user-label">User #{authentication.userId}</span>
				<button class="sign-out-button" type="button" disabled={loading} onclick={handleSignOut}>
					{loading ? 'Signing out...' : 'Sign out'}
				</button>
			</div>
		</nav>

		<main>
			{@render children()}
		</main>
	</div>
{/if}

<style lang="scss">
	.frame {
		width: 100%;
		height: 100svh;
		display: flex;
		flex-direction: column;

		.navbar {
			width: 100%;
			box-sizing: border-box;
			display: flex;
			justify-content: space-between;
			align-items: center;
			gap: 1rem;
			padding: 0.865rem 1rem;
			border-bottom: 1px solid var(--border);

			.brand {
				font-size: 1.75rem;
				font-weight: 600;
				color: var(--text);
				line-height: 1.2;
			}

			.account-actions {
				display: flex;
				align-items: center;
				justify-content: flex-end;
				justify-self: end;
				gap: 0.75rem;

				.user-label {
					font-size: 0.865rem;
					color: var(--text);
				}

				.sign-out-button {
					padding: 0.375rem 0.5rem;
					border: none;
					border-radius: 0.375rem;
					font-size: 0.865rem;
					background: var(--button);
					background-color: var(--button);
					cursor: pointer;
					color: var(--text-inverse);

					&:hover {
						background: var(--button-hover);
					}
				}
			}
		}
	}
</style>
