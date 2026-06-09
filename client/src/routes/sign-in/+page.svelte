<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { LockIcon, MailIcon } from '@lucide/svelte';
	import AuthenticationInput from '$lib/components/authentication-input/authentication-input.svelte';
	import { Button } from '$lib/components/shadcn/button/index.js';
	import { authenticationStore } from '$lib/stores/authentication-store.svelte';
	import { signInInputSchema } from '$lib/types/inputs/sign-in-input';

	let email = $state('');
	let password = $state('');
	let emailError = $state<string | undefined>(undefined);
	let passwordError = $state<string | undefined>(undefined);
	const authenticated = $derived(authenticationStore.authentication !== undefined);
	const loading = $derived(authenticationStore.loading);

	$effect(() => {
		if (authenticated) {
			void goto(resolve('/workbench'));
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

<main class="flex min-h-svh items-center justify-center bg-background px-4">
	<section class="flex w-full max-w-sm flex-col gap-6 rounded-lg border bg-card p-7 shadow-sm">
		<div class="text-center">
			<h1 class="text-2xl font-semibold">Sign in</h1>
		</div>

		<div class="flex flex-col gap-4">
			<AuthenticationInput
				icon={MailIcon}
				type="email"
				placeholder="Provide Email"
				autocomplete="email"
				bind:value={email}
				bind:error={emailError}
			/>

			<AuthenticationInput
				icon={LockIcon}
				type="password"
				placeholder="Provide Password"
				autocomplete="current-password"
				bind:value={password}
				bind:error={passwordError}
			/>
		</div>

		<div class="flex flex-col gap-3">
			<Button type="button" size="lg" class="w-full" disabled={loading} onclick={handleSignIn}>
				{loading ? 'Signing in...' : 'Next'}
			</Button>

			<div class="text-center text-sm">
				<span class="text-muted-foreground">or </span>
				<a href={resolve('/sign-up')} class="font-medium text-foreground underline-offset-4 hover:underline"> Sign up </a>
			</div>
		</div>
	</section>
</main>
