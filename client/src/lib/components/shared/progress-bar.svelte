<script lang="ts">
	import { Progress } from 'bits-ui';

	type Props = {
		percentage: number;
		label: string;
		valueText: string;
	};

	const red = 0;
	const orange = 30;
	const yellow = 60;
	const green = 120;
	const progressColorStops = [
		{ percentage: 0, hue: red },
		{ percentage: 55, hue: orange },
		{ percentage: 85, hue: yellow },
		{ percentage: 100, hue: green }
	];

	let { percentage, label, valueText }: Props = $props();
	const progressColor = $derived(getProgressColor(percentage));

	/**
	 * Returns an HSL color by interpolating the hue between the configured percentage stops surrounding the value.
	 */
	function getProgressColor(value: number): string {
		const nextStop = progressColorStops.find((stop) => stop.percentage >= value) ?? progressColorStops.at(-1)!;
		const previousStop = progressColorStops.findLast((stop) => stop.percentage <= value) ?? progressColorStops[0];
		if (previousStop === nextStop) return `hsl(${previousStop.hue} 75% 50%)`;

		const progressBetweenStops = (value - previousStop.percentage) / (nextStop.percentage - previousStop.percentage);
		const hue = previousStop.hue + (nextStop.hue - previousStop.hue) * progressBetweenStops;
		return `hsl(${hue} 75% 50%)`;
	}
</script>

<div class="progress-bar-container">
	<Progress.Root value={percentage} max={100} aria-label={label} aria-valuetext={valueText}>
		<div class="progress-bar-value" style:width={`${percentage}%`} style:background-color={progressColor}></div>
	</Progress.Root>
	<span class="progress-percentage">{percentage}%</span>
</div>

<style lang="scss">
	.progress-bar-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.75rem;
		width: 100%;

		:global([data-progress-root]) {
			width: 100%;
			height: 0.75rem;
			overflow: hidden;
			border-radius: 0.375rem;
			background: var(--surface);
		}

		.progress-bar-value {
			height: 100%;
			border-radius: inherit;
			transition:
				width 200ms ease,
				background-color 200ms ease;
		}

		.progress-percentage {
			color: var(--text-subtle);
			font-size: 1rem;
			font-weight: 500;
			line-height: 1;
		}
	}
</style>
