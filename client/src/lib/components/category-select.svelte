<script lang="ts">
	import { Check, ChevronDown, Plus, Trash2 } from '@lucide/svelte';
	import { tick } from 'svelte';
	import { Select as BitsSelect } from 'bits-ui';

	type Item = {
		value: string;
		label: string;
	};

	type Props = {
		items: Item[];
		value?: string;
		onadd?: (label: string) => void;
		ondelete?: (value: string) => void;
	};

	let { items, value = $bindable(), onadd, ondelete }: Props = $props();
	const addCategoryItemValue = $derived(getAvailableAddCategoryItemValue(items));
	let open = $state(false);
	let addingCategory = $state(false);
	let addCategoryHighlighted = $state(false);
	let categoryName = $state('');
	let categoryInput = $state<HTMLInputElement>();

	function getAvailableAddCategoryItemValue(items: Item[]): string {
		const itemValues = new Set(items.map((item) => item.value));
		let candidate = 'add-category';
		let suffix = 1;

		while (itemValues.has(candidate)) {
			candidate = `add-category-${suffix}`;
			suffix += 1;
		}

		return candidate;
	}

	async function showCategoryInput(): Promise<void> {
		addingCategory = true;
		await tick();
		categoryInput?.focus();
	}

	function addCategory(): void {
		if (categoryName === '') return;

		onadd?.(categoryName);
		open = false;
	}

	function handleCategoryNameInput(event: Event): void {
		const input = event.currentTarget as HTMLInputElement;
		const valueWithoutWhitespace = input.value.replace(/\s/g, '');
		input.value = valueWithoutWhitespace;
		categoryName = valueWithoutWhitespace;
	}

	function deleteCategory(event: MouseEvent, itemValue: string): void {
		event.stopPropagation();
		ondelete?.(itemValue);
	}

	function handleCategoryInputKeydown(event: KeyboardEvent): void {
		if (event.key !== 'Enter') return;

		event.preventDefault();
		addCategory();
	}

	function handleTriggerKeydown(event: KeyboardEvent): boolean {
		if (!open || !addCategoryHighlighted || (event.key !== 'Enter' && event.key !== ' ')) return false;

		event.preventDefault();
		void showCategoryInput();
		return true;
	}

	function runEventHandler(handler: unknown, event: Event): void {
		if (typeof handler === 'function') handler(event);
	}

	$effect(() => {
		if (open) return;

		addingCategory = false;
		addCategoryHighlighted = false;
		categoryName = '';
	});
</script>

<BitsSelect.Root type="single" {items} bind:value bind:open>
	<BitsSelect.Trigger>
		{#snippet child({ props })}
			<button
				{...props}
				id="task-category"
				class="select-trigger"
				onkeydown={(event) => {
					if (!handleTriggerKeydown(event)) runEventHandler(props.onkeydown, event);
				}}
			>
				<BitsSelect.Value placeholder="None chosen" />
				<ChevronDown size={18} />
			</button>
		{/snippet}
	</BitsSelect.Trigger>

	{#if items.length > 0 || onadd}
		<BitsSelect.Portal>
			<BitsSelect.Content sideOffset={6}>
				{#snippet child({ props: contentProps, wrapperProps })}
					<div {...wrapperProps}>
						<div {...contentProps} class:adding-category={addingCategory} class="select-content">
							<BitsSelect.Viewport>
								{#snippet child({ props })}
									<div {...props}>
										{#each items as item (item.value)}
											<BitsSelect.Item value={item.value} label={item.label}>
												{#snippet child({ props, selected })}
													<div {...props} class="select-item select-row">
														<span>{item.label}</span>
														<span class="item-actions">
															{#if selected}<Check size={18} aria-hidden="true" />{/if}
															{#if ondelete}
																<button
																	type="button"
																	class="delete-category"
																	aria-label={`Delete ${item.label}`}
																	onpointerdown={(event) => event.stopPropagation()}
																	onpointerup={(event) => event.stopPropagation()}
																	onclick={(event) => deleteCategory(event, item.value)}
																>
																	<Trash2 size={18} aria-hidden="true" />
																</button>
															{/if}
														</span>
													</div>
												{/snippet}
											</BitsSelect.Item>
										{/each}
										{#if onadd}
											<BitsSelect.Item
												value={addCategoryItemValue}
												label="Add category"
												onHighlight={() => (addCategoryHighlighted = true)}
												onUnhighlight={() => (addCategoryHighlighted = false)}
											>
												{#snippet child({ props })}
													<div
														{...props}
														class:editing={addingCategory}
														class="add-category select-item select-row"
														onpointerdown={(event) => {
															if (addingCategory) {
																event.stopPropagation();
																return;
															}
															runEventHandler(props.onpointerdown, event);
														}}
														onpointerup={(event) => {
															event.stopPropagation();
															if (addingCategory) return;

															event.preventDefault();
															void showCategoryInput();
														}}
													>
														{#if addingCategory}
															<input
																bind:this={categoryInput}
																value={categoryName}
																aria-label="Category name"
																placeholder="Category name"
																oninput={handleCategoryNameInput}
																onkeydown={handleCategoryInputKeydown}
															/>
														{:else}
															<span>Add category</span>
															<Plus size={18} aria-hidden="true" />
														{/if}
													</div>
												{/snippet}
											</BitsSelect.Item>
										{/if}
									</div>
								{/snippet}
							</BitsSelect.Viewport>
						</div>
					</div>
				{/snippet}
			</BitsSelect.Content>
		</BitsSelect.Portal>
	{/if}
</BitsSelect.Root>

<style lang="scss">
	.select-trigger {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		height: 2.75rem;
		box-sizing: border-box;
		padding: 0.625rem;
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		outline: none;
		background: transparent;
		color: var(--text);
		cursor: pointer;

		&:focus-visible,
		&[data-state='open'] {
			border-color: var(--border-accented);
		}

		&[data-placeholder] {
			color: var(--text-subtle);
		}
	}

	.select-content {
		z-index: 60;
		width: var(--bits-select-anchor-width);
		box-sizing: border-box;
		padding: 0.25rem;
		border: 1px solid var(--border);
		border-radius: 0.75rem;
		background: var(--background);
		color: var(--text);
		box-shadow: 0 0.75rem 2rem rgb(0 0 0 / 30%);
	}

	.select-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		box-sizing: border-box;
		padding: 0.625rem;
		border: 0;
		border-radius: 0.5rem;
		outline: none;
		background: transparent;
		color: var(--text);
		font: inherit;
	}

	.select-item {
		cursor: pointer;

		&[data-highlighted] {
			background: var(--surface-subtle);
		}
	}

	.select-content.adding-category .select-item:not(.add-category)[data-highlighted] {
		background: transparent;
	}

	.item-actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.delete-category {
		display: flex;
		padding: 0;
		border: 0;
		outline: none;
		background: transparent;
		color: var(--danger);
		cursor: pointer;

		&:focus-visible {
			background: var(--surface-subtle);
		}
	}

	.add-category {
		cursor: pointer;

		&.editing {
			background: var(--surface-subtle);
			cursor: text;
		}

		input {
			width: 100%;
			border: 0;
			outline: none;
			background: transparent;
			color: var(--text);
			font: inherit;
		}
	}
</style>
