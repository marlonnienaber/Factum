<script lang="ts">
	import Popover from '$lib/components/popover.svelte';
	import { Info, type LucideIcon } from '@lucide/svelte';
	import type { HTMLInputAttributes } from 'svelte/elements';

	type Props = {
		value: string;
		error?: string;
		icon: LucideIcon;
		placeholder: string;
		type?: 'email' | 'password' | 'text';
		autocomplete?: HTMLInputAttributes['autocomplete'];
		disabled?: boolean;
		'data-testid'?: string;
	};

	let {
		value = $bindable(),
		error = $bindable(),
		icon: Icon,
		placeholder,
		type = 'text',
		autocomplete,
		disabled = false,
		'data-testid': dataTestId
	}: Props = $props();

	const invalid = $derived(error !== undefined);

	function clearError() {
		error = undefined;
	}
</script>

<div class:invalid class="authentication-input">
	<span class="field-icon">
		<Icon size={18} />
	</span>
	<input bind:value {type} {placeholder} {autocomplete} {disabled} data-testid={dataTestId} oninput={clearError} />
	{#if invalid}
		<span class="error-icon">
			<Popover side="right" align="center" openOnHover>
				{#snippet trigger()}
					<span data-testid={dataTestId === undefined ? undefined : `${dataTestId}-error-trigger`}>
						<Info size={18} aria-hidden={true} />
					</span>
				{/snippet}
				<span class="error-popover-content">{error}</span>
			</Popover>
		</span>
	{/if}
</div>

<style lang="scss">
	.authentication-input {
		width: 20rem;
		display: flex;
		align-items: center;
		box-sizing: border-box;
		gap: 0.6125rem;
		padding: 0.6125rem 0.875rem;
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		color: var(--text);
		background-color: var(--background);

		&.invalid {
			border-color: var(--danger);
		}

		.field-icon {
			display: flex;
		}

		.error-icon {
			display: flex;
			align-items: center;
			width: 1.125rem;
			height: 1.125rem;
			color: var(--danger);
			line-height: 0;
		}

		input {
			flex: 1;
			border: none;
			outline: none;
			background: transparent;
			font-size: 1rem;
			color: var(--text);
		}
	}

	.error-popover-content {
		color: var(--text);
		font-size: 0.865rem;
	}
</style>
