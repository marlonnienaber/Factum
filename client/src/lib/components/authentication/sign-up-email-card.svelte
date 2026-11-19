<script lang="ts">
	import { resolve } from '$app/paths';
	import Input from '$lib/components/shared/input.svelte';
	import { authenticationService } from '$lib/services/authentication-service.svelte';
	import { themeStore } from '$lib/stores/theme-store.svelte';
	import { verifyEmailInputSchema, type VerifyEmailInput } from '$lib/types/inputs/verify-email-input';
	import type { ZodSafeParseResult } from 'zod';
	import { Mail } from '@lucide/svelte';
	import logoBlack from '$lib/assets/logo-black.svg';
	import logoWhite from '$lib/assets/logo-white.svg';

	type Props = { onverified: (email: string) => void };

	let { onverified }: Props = $props();
	let email = $state('');
	let emailError = $state<string | undefined>(undefined);
	const loading = $derived(authenticationService.loading);
	const isDarkTheme = $derived(themeStore.isDarkTheme);

	async function handleVerifyEmail(): Promise<void> {
		if (loading) return;

		const inputParseResult: ZodSafeParseResult<VerifyEmailInput> = verifyEmailInputSchema.safeParse({ email });
		if (!inputParseResult.success) {
			emailError = inputParseResult.error.issues[0]?.message;
			return;
		}

		emailError = undefined;
		const success = await authenticationService.verifyEmail(inputParseResult.data);
		if (success) onverified(inputParseResult.data.email);
	}
</script>

<div class="card">
	<div class="header">
		<img class="logo" src={isDarkTheme ? logoWhite : logoBlack} alt="Factum" />
		<h1 class="sign-up-text">Sign up</h1>
	</div>

	<Input
		id="email-input"
		bind:value={email}
		error={emailError}
		leadingIcon={Mail}
		type="email"
		placeholder="Provide Email"
		autocomplete="email"
		disabled={loading}
		oninput={() => (emailError = undefined)}
	/>

	<div class="footer">
		<button class="next-button" disabled={loading} onclick={handleVerifyEmail}>
			{loading ? 'Requesting link...' : 'Next'}
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
		width: 20rem;
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
