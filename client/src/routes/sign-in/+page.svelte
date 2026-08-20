<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Input from '$lib/components/shared/input.svelte';
	import { authenticationStore } from '$lib/stores/authentication-store.svelte';
	import { themeStore } from '$lib/stores/theme-store.svelte';
	import { signInInputSchema } from '$lib/types/inputs/sign-in-input';
	import { Lock, Mail } from '@lucide/svelte';
	import logoBlack from '$lib/assets/logo-black.svg';
	import logoWhite from '$lib/assets/logo-white.svg';

	let email = $state('');
	let password = $state('');
	let emailError = $state<string | undefined>(undefined);
	let passwordError = $state<string | undefined>(undefined);
	const authenticated = $derived(authenticationStore.authentication !== undefined);
	const loading = $derived(authenticationStore.loading);
	const isDarkTheme = $derived(themeStore.isDarkTheme);

	$effect(() => {
		if (authenticated) {
			void goto(resolve('/backlog'));
		}
	});

	async function handleSignIn(): Promise<void> {
		if (loading) return;

		const inputParseResult = signInInputSchema.safeParse({ email, password });
		if (!inputParseResult.success) {
			applyValidationErrors(inputParseResult.error.issues);
			return;
		}

		clearValidationErrors();
		await authenticationStore.signIn(inputParseResult.data);
	}

	function clearValidationErrors(): void {
		emailError = undefined;
		passwordError = undefined;
	}

	function applyValidationErrors(issues: Array<{ path: PropertyKey[]; message: string }>): void {
		clearValidationErrors();
		for (const issue of issues) {
			const field = issue.path[0];
			if (field === 'email') {
				emailError = emailError ?? issue.message;
			}
			if (field === 'password') {
				passwordError = passwordError ?? issue.message;
			}
		}
	}
</script>

<svelte:head>
	<title>Sign in</title>
</svelte:head>

<main class="page">
	<div class="card">
		<div class="header">
			<img class="logo" src={isDarkTheme ? logoWhite : logoBlack} alt="Factum" />
			<h1 class="sign-in-text">Sign in</h1>
		</div>

		<div class="inputs">
			<input
				bind:value={email}
				error={emailError}
				leadingIcon={Mail}
				type="email"
				placeholder="Provide Email"
				autocomplete="email"
				disabled={loading}
				oninput={() => (emailError = undefined)}
			/>
			<input
				bind:value={password}
				error={passwordError}
				leadingIcon={Lock}
				type="password"
				placeholder="Provide Password"
				autocomplete="current-password"
				disabled={loading}
				oninput={() => (passwordError = undefined)}
			/>
		</div>

		<div class="footer">
			<button class="next-button" disabled={loading} onclick={handleSignIn}>
				{loading ? 'Signing in...' : 'Next'}
			</button>

			<div class="sign-up-section">
				<span class="sign-up-text">or </span>
				<a class="sign-up-link" href={resolve('/sign-up')}>Sign up</a>
			</div>
		</div>
	</div>
</main>

<style lang="scss">
	.page {
		height: 100svh;
		width: 100%;
		display: flex;
		justify-content: center;
		align-items: center;

		.card {
			display: flex;
			flex-direction: column;
			padding: 1.5rem;
			gap: 1.5rem;
			border-radius: 1rem;
			border: solid 1px var(--border);
			background-color: var(--surface-subtle);

			.header {
				display: flex;
				flex-direction: column;
				align-items: center;
				gap: 1rem;

				.logo {
					height: 5rem;
				}

				.sign-in-text {
					color: var(--text);
					margin: 0;
					line-height: 1.25;
				}
			}

			.inputs {
				display: flex;
				flex-direction: column;
				align-items: center;
				width: 20rem;
				gap: 1rem;
			}

			.footer {
				display: flex;
				flex-direction: column;
				align-items: center;
				gap: 0.75rem;

				.next-button {
					width: 100%;
					padding: 0.5rem;
					background-color: var(--button);
					color: var(--text-inverse);
					border-radius: 0.5rem;
					font-size: 1rem;
					border: none;

					&:hover {
						background-color: var(--button-hover);
						cursor: pointer;
					}

					&:disabled {
						cursor: default;
					}
				}

				.sign-up-section {
					display: flex;
					gap: 0.25rem;

					.sign-up-text {
						font-size: 0.75rem;
						color: var(--text-subtle);
					}

					.sign-up-link {
						font-size: 0.75rem;
						color: var(--text);
					}
				}
			}
		}
	}
</style>
