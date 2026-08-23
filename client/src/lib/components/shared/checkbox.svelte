<script lang="ts">
	import { Check } from '@lucide/svelte';
	import { Checkbox as BitsCheckbox } from 'bits-ui';

	type Props = {
		checked: boolean;
		ariaLabel: string;
	};

	let { checked = $bindable(), ariaLabel }: Props = $props();
</script>

<BitsCheckbox.Root bind:checked aria-label={ariaLabel}>
	{#snippet child({ props, checked })}
		<button {...props} class="checkbox">
			{#if checked}
				<Check size={12} strokeWidth={3} aria-hidden={true} />
			{/if}
		</button>
	{/snippet}
</BitsCheckbox.Root>

<style lang="scss">
	.checkbox {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 1rem;
		height: 1rem;
		box-sizing: border-box;
		padding: 0;
		border: 1px solid var(--border-check-control);
		border-radius: 0.25rem;
		background: var(--background-check-control-unchecked);
		color: var(--text-inverse);
		cursor: pointer;

		&[data-state='checked'] {
			background: var(--background-check-control-checked);
		}
	}
</style>
