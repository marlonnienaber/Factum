<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Input from '$lib/components/shared/input.svelte';
	import { authenticationService } from '$lib/services/authentication-service.svelte';
	import { themeStore } from '$lib/stores/theme-store.svelte';
	import { signUpInputSchema } from '$lib/types/inputs/sign-up-input';
	import { Lock, Repeat } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';
	import logoBlack from '$lib/assets/logo-black.svg';
	import logoWhite from '$lib/assets/logo-white.svg';

	type Props = { token: string };

	let { token }: Props = $props();
	let password = $state('');
	let repeatedPassword = $state('');
	let passwordError = $state<string | undefined>(undefined);
	let repeatedPasswordError = $state<string | undefined>(undefined);
	const loading = $derived(authenticationService.loading);
	const isDarkTheme = $derived(themeStore.isDarkTheme);

	async function handleSignUp(): Promise<void> {
		if (loading) return;

		const inputParseResult = signUpInputSchema.safeParse({ token, password, repeatedPassword });
		if (!inputParseResult.success) {
			applyValidationErrors(inputParseResult.error.issues);
			return;
		}

		passwordError = undefined;
		repeatedPasswordError = undefined;
		const success = await authenticationService.signUp({
			token: inputParseResult.data.token,
			password: inputParseResult.data.password
		});
		if (success) {
			toast.success('Sign-up complete. Sign in to continue.');
			await goto(resolve('/sign-in'));
		}
	}

	function applyValidationErrors(issues: Array<{ path: PropertyKey[]; message: string }>): void {
		passwordError = undefined;
		repeatedPasswordError = undefined;
		const tokenIssue = issues.find((issue) => issue.path[0] === 'token');
		if (tokenIssue) toast.error(tokenIssue.message);
		for (const issue of issues) {
			if (issue.path[0] === 'password') passwordError = passwordError ?? issue.message;
			if (issue.path[0] === 'repeatedPassword') repeatedPasswordError = repeatedPasswordError ?? issue.message;
		}
	}
</script>

<div class="card">
	<div class="header">
		<img class="logo" src={isDarkTheme ? logoWhite : logoBlack} alt="Factum" />
		<h1 class="sign-up-text">Sign up</h1>
	</div>

	<div class="inputs">
		<Input
			id="password-input"
			bind:value={password}
			error={passwordError}
			leadingIcon={Lock}
			type="password"
			placeholder="Provide Password"
			autocomplete="new-password"
			disabled={loading}
			oninput={() => (passwordError = undefined)}
		/>
		<Input
			id="repeated-password-input"
			bind:value={repeatedPassword}
			error={repeatedPasswordError}
			leadingIcon={Repeat}
			type="password"
			placeholder="Repeat Password"
			autocomplete="new-password"
			disabled={loading}
			oninput={() => (repeatedPasswordError = undefined)}
		/>
	</div>

	<div class="footer">
		<button class="next-button" disabled={loading} onclick={handleSignUp}>
			{loading ? 'Signing up...' : 'Sign up'}
		</button>

		<div class="sign-in-section">
			<span class="sign-in-text">Already have an account?</span>
			<a class="sign-in-link" href={resolve('/sign-in')}>Sign in</a>
		</div>
	</div>
</div>

<style lang="scss">
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

				&:hover:not(:disabled) {
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
</style>
