<script lang="ts">
	type Props = {
		checked: boolean;
	};

	let { checked }: Props = $props();
</script>

<span class="check-circle" data-checked={checked} aria-hidden="true">
	{#if checked}
		<span class="check-circle-dot"></span>
	{/if}
</span>

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

		&[data-checked='true'] {
			background: var(--background-check-control-checked);
		}
	}

	.check-circle-dot {
		width: 0.5rem;
		height: 0.5rem;
		border-radius: 50%;
		background: var(--background-check-control-unchecked);
	}
</style>
