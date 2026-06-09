<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { authenticationStore } from '$lib/stores/authentication-store.svelte';
	import { signUpInputSchema } from '$lib/types/inputs/sign-up-input';

	let email = $state('');
	let password = $state('');
	let repeatedPassword = $state('');
	let emailError = $state<string | undefined>(undefined);
	let passwordError = $state<string | undefined>(undefined);
	let repeatedPasswordError = $state<string | undefined>(undefined);
	const authenticated = $derived(authenticationStore.authentication !== undefined);
	const loading = $derived(authenticationStore.loading);

	$effect(() => {
		if (authenticated) {
			void goto(resolve('/workbench'));
		}
	});

	async function handleSignUp(): Promise<void> {
		if (loading) return;

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

	function clearValidationErrors(): void {
		emailError = undefined;
		passwordError = undefined;
		repeatedPasswordError = undefined;
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
			if (field === 'repeatedPassword') {
				repeatedPasswordError = repeatedPasswordError ?? issue.message;
			}
		}
	}

	void handleSignUp;
</script>

<svelte:head>
	<title>Sign up</title>
</svelte:head>
