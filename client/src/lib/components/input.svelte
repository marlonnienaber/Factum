<script lang="ts">
	import Popover from '$lib/components/popover.svelte';
	import { Info, type LucideIcon } from '@lucide/svelte';
	import type { HTMLInputAttributes } from 'svelte/elements';

	type Props = {
		id?: string;
		value?: string;
		error?: string;
		leadingIcon?: LucideIcon;
		placeholder?: string;
		type?: 'email' | 'password' | 'text';
		autocomplete?: HTMLInputAttributes['autocomplete'];
		disabled?: boolean;
		oninput?: () => void;
		'data-testid'?: string;
	};

	let {
		id,
		value = $bindable(''),
		error,
		leadingIcon: LeadingIcon,
		placeholder,
		type = 'text',
		autocomplete,
		disabled = false,
		oninput,
		'data-testid': dataTestId
	}: Props = $props();

	const invalid = $derived(error !== undefined);
</script>

<div class:invalid class="input">
	{#if LeadingIcon}
		<span class="leading-icon"><LeadingIcon size={18} /></span>
	{/if}
	<input {id} bind:value {type} {placeholder} {autocomplete} {disabled} {oninput} data-testid={dataTestId} />
	{#if invalid}
		<span class="error-icon">
			<Popover side="right" align="center" openOnHover>
				{#snippet trigger()}
					<span data-testid={dataTestId === undefined ? undefined : `${dataTestId}-error-trigger`}>
						<Info size={18} aria-hidden={true} />
					</span>
				{/snippet}
				<span class="error-message">{error}</span>
			</Popover>
		</span>
	{/if}
</div>

<style lang="scss">
	.input {
		display: flex;
		align-items: center;
		width: 100%;
		height: 2.75rem;
		box-sizing: border-box;
		gap: 0.6125rem;
		padding: 0.6125rem 0.625rem;
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		background: var(--background);
		color: var(--text);

		&:focus-within {
			border-color: var(--border-accented);
		}

		&.invalid {
			border-color: var(--danger);
		}

		.leading-icon {
			display: flex;
		}

		input {
			flex: 1;
			min-width: 0;
			border: 0;
			outline: none;
			background: transparent;
			color: var(--text);
			font: inherit;
		}

		.error-icon {
			display: flex;
			align-items: center;
			width: 1.125rem;
			height: 1.125rem;
			color: var(--danger);
			line-height: 0;
		}
	}

	.error-message {
		color: var(--text);
		font-size: 0.865rem;
	}
</style>
