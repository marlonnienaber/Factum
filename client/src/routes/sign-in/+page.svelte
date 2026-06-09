<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
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

	void handleSignIn;
</script>

<svelte:head>
	<title>Sign in</title>
</svelte:head>
