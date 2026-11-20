<script lang="ts">
	import { resolve } from '$app/paths';
	import { authenticationService } from '$lib/services/authentication-service.svelte';
	import { themeStore } from '$lib/stores/theme-store.svelte';
	import { ArrowLeft } from '@lucide/svelte';
	import { now as getCurrentTime, type ZonedDateTime } from '@internationalized/date';
	import logoBlack from '$lib/assets/logo-black.svg';
	import logoWhite from '$lib/assets/logo-white.svg';
	type VerifyEmailTimeWindow = {
		startedAt: ZonedDateTime;
		requestCount: number;
	};

	type Props = { email: string; onback: () => void };

	let { email, onback }: Props = $props();
	const mountedAt = getCurrentTime('UTC');
	let minutelyWindowStartedAt = $state(mountedAt);
	let hourlyWindow = $state<VerifyEmailTimeWindow>({ startedAt: mountedAt, requestCount: 1 });
	let dailyWindow = $state<VerifyEmailTimeWindow>({ startedAt: mountedAt, requestCount: 1 });
	let now = $state(mountedAt);
	const secondsUntilNextRequestPossible = $derived(getSecondsUntilNextRequestPossible());
	const hourlyLimitActive = $derived(isHourlyLimitActive());
	const dailyLimitActive = $derived(isDailyLimitActive());
	const loading = $derived(authenticationService.loading);
	const isDarkTheme = $derived(themeStore.isDarkTheme);
	const buttonText = $derived(getButtonText());

	$effect(() => {
		const interval = window.setInterval(() => (now = getCurrentTime('UTC')), 1000);
		return () => window.clearInterval(interval);
	});

	function isHourlyLimitActive(): boolean {
		return hourlyWindow.requestCount >= 5 && now.compare(hourlyWindow.startedAt.add({ hours: 1 })) < 0;
	}

	function isDailyLimitActive(): boolean {
		return dailyWindow.requestCount >= 10 && now.compare(dailyWindow.startedAt.add({ hours: 24 })) < 0;
	}

	function getSecondsUntilNextRequestPossible(): number {
		const nextRequestPossibleDate = getNextRequestPossibleDate();
		const millisecondsUntilNextRequestPossible = Math.max(0, nextRequestPossibleDate.compare(now));
		const secondsUntilNextRequestPossible = millisecondsUntilNextRequestPossible / 1000;
		const roundedSecondsUntilNextRequestPossible = Math.ceil(secondsUntilNextRequestPossible);
		return roundedSecondsUntilNextRequestPossible;
	}

	function getNextRequestPossibleDate(): ZonedDateTime {
		const minutelyWindowEnd = minutelyWindowStartedAt.add({ minutes: 1 });
		let nextRequestPossibleDate = minutelyWindowEnd;

		if (hourlyLimitActive) {
			const hourlyWindowEnd = hourlyWindow.startedAt.add({ hours: 1 });
			if (hourlyWindowEnd.compare(nextRequestPossibleDate) > 0) {
				nextRequestPossibleDate = hourlyWindowEnd;
			}
		}

		if (dailyLimitActive) {
			const dailyWindowEnd = dailyWindow.startedAt.add({ hours: 24 });
			if (dailyWindowEnd.compare(nextRequestPossibleDate) > 0) {
				nextRequestPossibleDate = dailyWindowEnd;
			}
		}
		return nextRequestPossibleDate;
	}

	function recordEmailRequest(triggeredAt: ZonedDateTime): void {
		minutelyWindowStartedAt = triggeredAt;
		if (triggeredAt.compare(hourlyWindow.startedAt.add({ hours: 1 })) < 0) {
			hourlyWindow = { ...hourlyWindow, requestCount: hourlyWindow.requestCount + 1 };
		} else {
			hourlyWindow = { startedAt: triggeredAt, requestCount: 1 };
		}
		if (triggeredAt.compare(dailyWindow.startedAt.add({ hours: 24 })) < 0) {
			dailyWindow = { ...dailyWindow, requestCount: dailyWindow.requestCount + 1 };
		} else {
			dailyWindow = { startedAt: triggeredAt, requestCount: 1 };
		}
	}

	async function handleResend(): Promise<void> {
		if (loading) return;
		const triggeredAt = getCurrentTime('UTC');
		now = triggeredAt;
		if (triggeredAt.compare(getNextRequestPossibleDate()) < 0) return;
		await authenticationService.verifyEmail({ email });
		const completedAt = getCurrentTime('UTC');
		now = completedAt;
		recordEmailRequest(completedAt);
	}

	function getButtonText(): string {
		if (loading) return 'Requesting link...';
		if (secondsUntilNextRequestPossible > 0) return `Resend in ${formatWait(secondsUntilNextRequestPossible)}`;
		return 'Resend link';
	}

	function formatWait(seconds: number): string {
		const hours = Math.floor(seconds / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);
		const remainingSeconds = seconds % 60;
		if (hours > 0) return `${hours}h ${minutes}m`;
		if (minutes > 0) return `${minutes}m ${remainingSeconds}s`;
		return `${remainingSeconds}s`;
	}
</script>

<div class="card">
	<button class="back-button" type="button" aria-label="Back to email entry" disabled={loading} onclick={onback}>
		<ArrowLeft size={20} aria-hidden={true} />
	</button>

	<div class="header">
		<img class="logo" src={isDarkTheme ? logoWhite : logoBlack} alt="Factum" />
		<h1 class="sign-up-text">Sign up</h1>
	</div>

	<p class="text">
		Check the inbox of <span class="email-address">{email}</span>. Open the link in the latest mail to complete the sign up. Check
		spam too.
	</p>

	<div class="footer">
		<button class="resend-button" disabled={loading || secondsUntilNextRequestPossible > 0} onclick={handleResend}>
			{buttonText}
		</button>

		<div class="sign-in-section">
			<span class="sign-in-text">Already have an account?</span>
			<a class="sign-in-link" href={resolve('/sign-in')}>Sign in</a>
		</div>
	</div>
</div>

<style lang="scss">
	.card {
		position: relative;
		display: flex;
		flex-direction: column;
		padding: 1.5rem;
		gap: 1.5rem;
		border-radius: 1rem;
		border: solid 1px var(--border);
		background-color: var(--surface-subtle);

		.back-button {
			position: absolute;
			top: 1rem;
			left: 1rem;
			display: flex;
			align-items: center;
			justify-content: center;
			width: 2rem;
			height: 2rem;
			padding: 0;
			border: none;
			border-radius: 0.5rem;
			background: transparent;
			color: var(--text);

			&:hover:not(:disabled) {
				background: var(--button-ghost);
				cursor: pointer;
			}

			&:disabled {
				cursor: default;
			}
		}

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

		.text {
			width: 20rem;
			margin: 0;
			color: var(--text-subtle);
			font-size: 0.865rem;
			line-height: 1.6;
			text-align: center;
			overflow-wrap: anywhere;

			.email-address {
				color: var(--text);
			}
		}

		.footer {
			display: flex;
			flex-direction: column;
			align-items: center;
			gap: 0.75rem;

			.resend-button {
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
