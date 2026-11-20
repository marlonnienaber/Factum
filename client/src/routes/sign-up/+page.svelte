<script lang="ts">
	import { page } from '$app/state';
	import SignUpEmailCard from '$lib/components/authentication/sign-up-email-card.svelte';
	import SignUpCheckEmailCard from '$lib/components/authentication/sign-up-check-email-card.svelte';
	import SignUpPasswordCard from '$lib/components/authentication/sign-up-password-card.svelte';

	type SignUpStep = { kind: 'email' } | { kind: 'check-email'; email: string } | { kind: 'password'; token: string };

	const token = $derived(page.url.searchParams.get('token') ?? undefined);
	let step = $derived<SignUpStep>(getInitialStep());
	const title = $derived(getTitle());

	function getInitialStep(): SignUpStep {
		if (token !== undefined) {
			return { kind: 'password', token };
		} else {
			return { kind: 'email' };
		}
	}

	function getTitle(): string {
		if (step.kind === 'password') return 'Choose your password';
		else if (step.kind === 'check-email') return 'Check your email';
		else return 'Sign up';
	}

	function setStepToEmail(): void {
		step = { kind: 'email' };
	}

	function setStepToCheckEmail(email: string): void {
		step = { kind: 'check-email', email };
	}
</script>

<svelte:head>
	<title>{title}</title>
	<meta name="referrer" content="no-referrer" />
</svelte:head>

<main class="page">
	{#if step.kind === 'password'}
		<SignUpPasswordCard token={step.token} />
	{:else if step.kind === 'check-email'}
		<SignUpCheckEmailCard email={step.email} onback={setStepToEmail} />
	{:else}
		<SignUpEmailCard onverified={setStepToCheckEmail} />
	{/if}
</main>

<style lang="scss">
	.page {
		height: 100svh;
		width: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
	}
</style>
