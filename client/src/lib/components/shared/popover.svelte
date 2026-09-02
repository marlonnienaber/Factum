<script lang="ts">
	import { Popover as BitsPopover } from 'bits-ui';
	import type { Snippet } from 'svelte';

	type Props = {
		children: Snippet;
		trigger: Snippet;
		side?: 'top' | 'right' | 'bottom' | 'left';
		align?: 'start' | 'center' | 'end';
		open?: boolean;
		openOnHover?: boolean;
		triggerAriaLabel?: string;
	};

	let {
		children,
		trigger,
		side = 'bottom',
		align = 'center',
		open = $bindable(false),
		openOnHover = false,
		triggerAriaLabel
	}: Props = $props();
</script>

<BitsPopover.Root bind:open>
	<BitsPopover.Trigger {openOnHover} openDelay={0} closeDelay={0}>
		{#snippet child({ props })}
			<button {...props} type="button" class="trigger" aria-label={triggerAriaLabel}>
				{@render trigger()}
			</button>
		{/snippet}
	</BitsPopover.Trigger>
	<BitsPopover.Portal>
		<BitsPopover.Content {side} {align} sideOffset={6}>
			{#snippet child({ props, wrapperProps })}
				<div {...wrapperProps}>
					<div {...props} class="content">
						{@render children()}
					</div>
				</div>
			{/snippet}
		</BitsPopover.Content>
	</BitsPopover.Portal>
</BitsPopover.Root>

<style lang="scss">
	.trigger {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 1.125rem;
		height: 1.125rem;
		border: none;
		padding: 0;
		background: transparent;
		color: inherit;
		cursor: pointer;
	}

	.content {
		z-index: 60;
		box-sizing: border-box;
		width: fit-content;
		padding: 0.5rem;
		border: 1px solid var(--border);
		border-radius: 0.75rem;
		background: var(--background);
	}
</style>
