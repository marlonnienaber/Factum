<script lang="ts">
	import Popover from '$lib/components/shared/popover.svelte';
	import { Info } from '@lucide/svelte';

	type Props = {
		id: string;
		value: string;
		placeholder?: string;
		error?: string;
		oninput?: () => void;
		onblur?: () => void;
	};

	let { id, value = $bindable(), placeholder, error, oninput, onblur }: Props = $props();
	const invalid = $derived(error !== undefined);
</script>

<div class:invalid class="textarea-container">
	<textarea {id} bind:value {placeholder} {oninput} {onblur} aria-invalid={invalid}></textarea>
	{#if invalid}
		<span class="error-indicator-icon">
			<Popover side="right" align="center" openOnHover>
				{#snippet trigger()}
					<span id={`${id}-error-indicator-popover-trigger`}><Info size={18} aria-hidden="true" /></span>
				{/snippet}
				<span class="error-indicator-message">{error}</span>
			</Popover>
		</span>
	{/if}
</div>

<style lang="scss">
	.textarea-container {
		position: relative;

		&.invalid {
			textarea {
				padding-right: calc(0.625rem + 1.125rem + 0.6125rem);
			}
		}

		textarea {
			width: 100%;
			min-height: 6rem;
			box-sizing: border-box;
			padding: 0.625rem;
			border: 1px solid var(--border);
			border-radius: 0.5rem;
			outline: none;
			background: transparent;
			color: var(--text);
			font: inherit;
			font-size: 0.865rem;
			resize: vertical;

			&:focus {
				border-color: var(--border-accented);
			}

			&[aria-invalid='true'] {
				border-color: var(--danger);
			}
		}

		.error-indicator-icon {
			position: absolute;
			top: 50%;
			right: 0.625rem;
			display: flex;
			width: 1.125rem;
			height: 1.125rem;
			transform: translateY(-50%);
			color: var(--danger);

			.error-indicator-message {
				color: var(--text);
			}
		}
	}
</style>
