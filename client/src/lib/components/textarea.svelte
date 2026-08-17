<script lang="ts">
	import Popover from '$lib/components/popover.svelte';
	import { Info } from '@lucide/svelte';

	type Props = {
		id: string;
		value?: string;
		placeholder?: string;
		error?: string;
	};

	let { id, value = $bindable(''), placeholder, error }: Props = $props();
	const invalid = $derived(error !== undefined);
</script>

<div class="textarea-container">
	<textarea {id} bind:value {placeholder} aria-invalid={invalid}></textarea>
	{#if invalid}
		<span class="info-icon">
			<Popover side="right" align="center" openOnHover>
				{#snippet trigger()}<Info size={18} aria-hidden="true" />{/snippet}
				<span class="error-message">{error}</span>
			</Popover>
		</span>
	{/if}
</div>

<style lang="scss">
	.textarea-container {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	textarea {
		flex: 1;
		min-width: 0;
		min-height: 4.5rem;
		box-sizing: border-box;
		padding: 0.625rem;
		border: 2px solid var(--border);
		border-radius: 0.5rem;
		outline: none;
		background: transparent;
		color: var(--text);
		font: inherit;
		font-size: 0.9rem;
		resize: vertical;

		&:focus {
			border-color: var(--text-subtle);
		}

		&[aria-invalid='true'] {
			border-color: var(--danger);
		}
	}

	.info-icon {
		display: flex;
		color: var(--danger);
	}

	.error-message {
		color: var(--text);
	}
</style>
