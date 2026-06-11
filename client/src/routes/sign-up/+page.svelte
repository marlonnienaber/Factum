<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import AuthenticationInput from '$lib/components/authentication-input.svelte';
	import { authenticationStore } from '$lib/stores/authentication-store.svelte';
	import { themeStore } from '$lib/stores/theme-store.svelte';
	import { signUpInputSchema } from '$lib/types/inputs/sign-up-input';
	import { Lock, Mail, Repeat } from '@lucide/svelte';
	import logoBlack from '$lib/assets/logo-black.svg';
	import logoWhite from '$lib/assets/logo-white.svg';

	let email = $state('');
	let password = $state('');
	let repeatedPassword = $state('');
	let emailError = $state<string | undefined>(undefined);
	let passwordError = $state<string | undefined>(undefined);
	let repeatedPasswordError = $state<string | undefined>(undefined);
	const authenticated = $derived(authenticationStore.authentication !== undefined);
	const loading = $derived(authenticationStore.loading);
	const isDarkTheme = $derived(themeStore.isDarkTheme);

	$effect(() => {
		if (authenticated) {
			void goto(resolve('/workbench'));
		}
	});

	async function handleSignUp() {
		const inputParseResult = signUpInputSchema.safeParse({ email, password, repeatedPassword });
		if (!inputParseResult.success) {
			applyValidationErrors(inputParseResult.error.issues);
			return;
		}

		clearValidationErrors();
		await authenticationStore.signUp({
			email: inputParseResult.data.email,
			password: inputParseResult.data.password
		});
	}

	function clearValidationErrors() {
		emailError = undefined;
		passwordError = undefined;
		repeatedPasswordError = undefined;
	}

	function applyValidationErrors(issues: Array<{ path: PropertyKey[]; message: string }>) {
		clearValidationErrors();
		for (const issue of issues) {
			const field = issue.path[0];
			if (field === 'email') {
				emailError = emailError ?? issue.message;
			}
			if (field === 'password') {
				passwordError = passwordError ?? issue.message;
			}
			if (field === 'repeatedPassword') {
				repeatedPasswordError = repeatedPasswordError ?? issue.message;
			}
		}
	}
</script>

<svelte:head>
	<title>Sign up</title>
</svelte:head>

<main class="page">
	<div class="card">
		<div class="header">
			<img class="logo" src={isDarkTheme ? logoWhite : logoBlack} alt="Factum" />
			<h1 class="sign-up-text">Sign up</h1>
		</div>

		<div class="inputs">
			<AuthenticationInput
				bind:value={email}
				bind:error={emailError}
				icon={Mail}
				type="email"
				placeholder="Provide Email"
				autocomplete="email"
				disabled={loading}
				data-testid="email-input"
			/>
			<AuthenticationInput
				bind:value={password}
				bind:error={passwordError}
				icon={Lock}
				type="password"
				placeholder="Provide Password"
				autocomplete="new-password"
				disabled={loading}
				data-testid="password-input"
			/>
			<AuthenticationInput
				bind:value={repeatedPassword}
				bind:error={repeatedPasswordError}
				icon={Repeat}
				type="password"
				placeholder="Repeat Password"
				autocomplete="new-password"
				disabled={loading}
				data-testid="repeated-password-input"
			/>
		</div>

		<div class="footer">
			<button class="next-button" data-testid="next-button" disabled={loading} onclick={handleSignUp}>
				{loading ? 'Signing up...' : 'Next'}
			</button>

			<div class="sign-in-section">
				<span class="sign-in-text">or </span>
				<a class="sign-in-link" href={resolve('/sign-in')}>Sign in</a>
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

				.sign-up-text {
					color: var(--text);
					margin: 0;
					line-height: 1.25;
				}
			}

			.inputs {
				display: flex;
				flex-direction: column;
				align-items: center;
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

				.sign-in-section {
					display: flex;
					gap: 0.25rem;

					.sign-in-text {
						font-size: 0.75rem;
						color: var(--text-subtle);
					}

					.sign-in-link {
						font-size: 0.75rem;
						color: var(--text);
					}
				}
			}
		}
	}
</style>
