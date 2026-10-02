<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { MouseEventHandler } from 'svelte/elements';

	type Variant = 'default' | 'success' | 'danger';
	type Props = {
		children: Snippet;
		trailingIcon?: Snippet;
		variant?: Variant;
		disabled?: boolean;
		onclick: MouseEventHandler<HTMLButtonElement>;
	};

	let { children, trailingIcon, variant = 'default', disabled = false, onclick }: Props = $props();
</script>

<button type="button" class="button" data-variant={variant} {onclick} {disabled}>
	{@render children()}
	{#if trailingIcon}
		{@render trailingIcon()}
	{/if}
</button>

<style lang="scss">
	.button {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.25rem;
		padding: 0.375rem 0.5rem;
		border: none;
		border-radius: 0.5rem;
		font-size: 0.865rem;
		cursor: pointer;
		color: var(--text-inverse);

		&:hover:not(:disabled) {
			opacity: 0.8;
		}

		&:disabled {
			opacity: 0.65;
			cursor: default;
		}

		&[data-variant='default'] {
			background: var(--button);
		}

		&[data-variant='success'] {
			background: var(--success);
		}

		&[data-variant='danger'] {
			background: var(--danger);
		}
	}
</style>
