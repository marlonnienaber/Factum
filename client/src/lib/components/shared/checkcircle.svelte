<script lang="ts">
	import { Checkbox as BitsCheckbox } from 'bits-ui';

	type Props = {
		checked: boolean;
		ariaLabel: string;
		disabled?: boolean;
		onCheckedChange?: (checked: boolean) => void;
	};

	let { checked, ariaLabel, disabled = false, onCheckedChange }: Props = $props();
</script>

<BitsCheckbox.Root {checked} {disabled} {onCheckedChange} aria-label={ariaLabel}>
	{#snippet child({ props, checked })}
		<button {...props} class="check-circle" onpointerdown={(event) => event.stopPropagation()}>
			{#if checked}
				<span class="check-circle-dot"></span>
			{/if}
		</button>
	{/snippet}
</BitsCheckbox.Root>

<style lang="scss">
	.check-circle {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 1rem;
		height: 1rem;
		box-sizing: border-box;
		padding: 0;
		border: 1px solid var(--border-check-control);
		border-radius: 50%;
		background: var(--background-check-control-unchecked);
		cursor: pointer;

		&[data-state='checked'] {
			background: var(--background-check-control-checked);
		}

		&:focus-visible {
			outline: 2px solid var(--border-accented);
			outline-offset: 2px;
		}

		&:disabled {
			cursor: default;
			opacity: 0.5;
		}
	}

	.check-circle-dot {
		width: 0.5rem;
		height: 0.5rem;
		border-radius: 50%;
		background: var(--background-check-control-unchecked);
	}
</style>
