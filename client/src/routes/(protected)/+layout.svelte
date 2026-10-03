<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { authenticationService } from '$lib/services/authentication-service.svelte';
	import { themeService } from '$lib/services/theme-service.svelte';
	import { authenticationStore } from '$lib/stores/authentication-store.svelte';
	import { categoryStore } from '$lib/stores/category-store.svelte';
	import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
	import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
	import { trashedTaskStore } from '$lib/stores/trashed-task-store.svelte';
	import { themeStore } from '$lib/stores/theme-store.svelte';
	import { todayProgressStore } from '$lib/stores/today-progress-store.svelte';
	import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
	import { Moon, Sun } from '@lucide/svelte';
	import logoBlack from '$lib/assets/logo-black.svg';
	import logoWhite from '$lib/assets/logo-white.svg';

	const { children } = $props();
	const authentication = $derived(authenticationStore.authentication);
	const notAttemptedToFetchAuthentication = $derived(!authenticationService.fetchAttempted);
	const loading = $derived(authenticationService.loading);
	const isDarkTheme = $derived(themeStore.isDarkTheme);

	$effect(() => {
		if (!authentication) {
			if (!loading) {
				if (notAttemptedToFetchAuthentication) {
					void authenticationService.fetchCurrentAuthenticationIfNecessary();
				} else {
					categoryStore.reset();
					datedUntrashedDoneTaskStore.reset();
					datedUntrashedOpenTaskStore.reset();
					undatedTaskStore.reset();
					trashedTaskStore.reset();
					todayProgressStore.reset();
					void goto(resolve('/sign-in'));
				}
			}
		}
	});
</script>

{#if authentication}
	<div class="frame">
		<nav class="navbar">
			<div class="brand">
				<img class="logo" src={isDarkTheme ? logoWhite : logoBlack} alt="Factum" />
				<span class="name">actum</span>
			</div>
			<div class="account-actions">
				<a href={resolve('/workbench')}>Workbench</a>
				<a href={resolve('/backlog')}>Backlog</a>
				<a href={resolve('/archive')}>Archive</a>
				<button class="theme-toggle-button" type="button" onclick={() => themeService.toggleTheme()}>
					{#if isDarkTheme}
						<Sun size={18} aria-hidden={true} />
					{:else}
						<Moon size={18} aria-hidden={true} />
					{/if}
				</button>
				<button class="sign-out-button" type="button" disabled={loading} onclick={() => void authenticationService.signOut()}>
					{loading ? 'Signing out...' : 'Sign out'}
				</button>
			</div>
		</nav>

		<main class="content">
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
			padding: 1rem 1rem 0 1rem;

			.brand {
				display: flex;
				align-items: center;
				gap: 0.375rem;

				.logo {
					height: 2.25rem;
				}

				.name {
					font-size: 1.5rem;
					font-weight: 600;
					color: var(--text);
					line-height: 1;
				}
			}

			.account-actions {
				display: flex;
				align-items: center;
				justify-content: flex-end;
				justify-self: end;
				gap: 0.75rem;

				a {
					color: var(--text);
					font-size: 0.865rem;
					text-decoration: none;

					&:hover {
						text-decoration: underline;
					}
				}

				.theme-toggle-button {
					display: flex;
					justify-content: center;
					align-items: center;
					width: 1.75rem;
					height: 1.75rem;
					border: none;
					background: none;
					color: var(--button);
					cursor: pointer;
					border-radius: 0.5rem;

					&:hover {
						background-color: var(--button-ghost);
					}
				}

				.sign-out-button {
					padding: 0.375rem 0.5rem;
					border: none;
					border-radius: 0.5rem;
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

		.content {
			flex: 1;
			min-height: 0;
			display: flex;
			flex-direction: column;
		}
	}
</style>
