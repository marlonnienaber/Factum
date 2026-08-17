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

<div class:invalid class="textarea-container">
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
		position: relative;

		&.invalid textarea {
			padding-right: calc(0.625rem + 1.125rem + 0.6125rem);
		}
	}

	textarea {
		display: block;
		width: 100%;
		min-height: 4.5rem;
		box-sizing: border-box;
		padding: 0.625rem;
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		outline: none;
		background: transparent;
		color: var(--text);
		font: inherit;
		resize: vertical;

		&:focus {
			border-color: var(--border-accented);
		}

		&[aria-invalid='true'] {
			border-color: var(--danger);
		}
	}

	.info-icon {
		position: absolute;
		top: 50%;
		right: 0.625rem;
		display: flex;
		width: 1.125rem;
		height: 1.125rem;
		transform: translateY(-50%);
		color: var(--danger);
		line-height: 0;
	}

	.error-message {
		color: var(--text);
	}
</style>
