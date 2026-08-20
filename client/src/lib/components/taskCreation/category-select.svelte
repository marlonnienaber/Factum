<script lang="ts">
	import { categoryStore } from '$lib/stores/category-store.svelte.js';
	import { Check, ChevronDown, Plus, Trash2 } from '@lucide/svelte';
	import { Select as BitsSelect } from 'bits-ui';

	type Props = {
		id?: string;
		onchange: (value: number | undefined) => void;
	};

	let { id, onchange }: Props = $props();
	const items = $derived(categoryStore.categories.map((category) => ({ value: String(category.id), label: category.title })));
	let value = $state<string>();
	let open = $state(false);

	const addCategoryItemValue = $derived(getAvailableAddCategoryItemValue(items));
	let addCategoryItemSelected = $state(false);

	let creatingNewCategory = $state(false);
	let newCategoryName = $state('');

	function getAvailableAddCategoryItemValue(items: Array<{ value: string }>): string {
		const itemValues = new Set(items.map((item) => item.value));
		let candidate = 'add-category';
		let suffix = 1;
		while (itemValues.has(candidate)) {
			candidate = `add-category-${suffix}`;
			suffix += 1;
		}
		return candidate;
	}

	function handleNewCategoryNameInput(event: Event): void {
		const input = event.currentTarget as HTMLInputElement;
		const valueWithoutWhitespace = input.value.replace(/\s/g, '');
		input.value = valueWithoutWhitespace;
		newCategoryName = valueWithoutWhitespace;
	}

	async function deleteCategory(event: MouseEvent, itemValue: string): Promise<void> {
		event.stopPropagation();
		const categoryId = Number(itemValue);
		await categoryStore.deleteCategory(categoryId);
		if (categoryStore.categories.some((category) => category.id === categoryId)) return;

		if (value === itemValue) {
			value = undefined;
			onchange(value);
		}
	}

	function handleCategorySelectKeydown(event: KeyboardEvent, bitsUiHandler: unknown): void {
		const isAddCategoryEvent = open && addCategoryItemSelected && (event.key === 'Enter' || event.key === ' ');
		if (isAddCategoryEvent) {
			event.preventDefault();
			creatingNewCategory = true;
		} else if (typeof bitsUiHandler === 'function') {
			bitsUiHandler(event);
		}
	}

	async function handleNewCategoryNameInputKeydown(event: KeyboardEvent): Promise<void> {
		if (event.key !== 'Enter') return;
		event.preventDefault();
		if (newCategoryName === '') return;

		const oldCategoryIds = new Set(categoryStore.categories.map((category) => category.id));
		await categoryStore.createCategory(newCategoryName);
		const createdCategory = categoryStore.categories.find((category) => !oldCategoryIds.has(category.id));
		if (createdCategory === undefined) return;

		value = String(createdCategory.id);
		onchange(createdCategory.id);
		open = false;
	}

	function handleMouseUpOnCreateNewCategoryItem(event: PointerEvent): void {
		event.stopPropagation();

		if (!creatingNewCategory) {
			creatingNewCategory = true;
		}
	}

	$effect(() => {
		if (!open) {
			creatingNewCategory = false;
			addCategoryItemSelected = false;
			newCategoryName = '';
		}
	});
</script>

<BitsSelect.Root type="single" {items} bind:value bind:open onValueChange={(value) => onchange(Number(value))}>
	<BitsSelect.Trigger>
		{#snippet child({ props })}
			<button {...props} {id} class="select-trigger" onkeydown={(event) => handleCategorySelectKeydown(event, props.onkeydown)}>
				<BitsSelect.Value placeholder="None chosen" />
				<ChevronDown size={18} />
			</button>
		{/snippet}
	</BitsSelect.Trigger>

	<BitsSelect.Portal>
		<BitsSelect.Content sideOffset={6}>
			{#snippet child({ props: contentProps, wrapperProps })}
				<div {...wrapperProps}>
					<div {...contentProps} class:creating-new-category={creatingNewCategory} class="select-content">
						<BitsSelect.Viewport>
							{#snippet child({ props })}
								<div {...props}>
									{#each items as item (item.value)}
										<BitsSelect.Item value={item.value} label={item.label}>
											{#snippet child({ props, selected })}
												<div {...props} class="select-item">
													<span>{item.label}</span>
													<span class="item-icons">
														{#if selected}<Check size={18} aria-hidden="true" />{/if}
														<button
															type="button"
															class="delete-button"
															aria-label={`Delete ${item.label}`}
															onpointerdown={(event) => event.stopPropagation()}
															onpointerup={(event) => event.stopPropagation()}
															onclick={(event) => void deleteCategory(event, item.value)}
														>
															<Trash2 size={18} aria-hidden="true" />
														</button>
													</span>
												</div>
											{/snippet}
										</BitsSelect.Item>
									{/each}
									<BitsSelect.Item
										value={addCategoryItemValue}
										label="Add category"
										onHighlight={() => (addCategoryItemSelected = true)}
										onUnhighlight={() => (addCategoryItemSelected = false)}
									>
										{#snippet child({ props })}
											<div
												{...props}
												class:editing={creatingNewCategory}
												class="create-new-category-item select-item"
												onpointerdown={(event) => event.stopPropagation()}
												onpointerup={handleMouseUpOnCreateNewCategoryItem}
											>
												{#if creatingNewCategory}
													<input
														{@attach (element) => element.focus()}
														value={newCategoryName}
														aria-label="Category name"
														placeholder="Category name"
														oninput={handleNewCategoryNameInput}
														onkeydown={handleNewCategoryNameInputKeydown}
													/>
												{:else}
													<span>Add category</span>
													<Plus size={18} aria-hidden="true" />
												{/if}
											</div>
										{/snippet}
									</BitsSelect.Item>
								</div>
							{/snippet}
						</BitsSelect.Viewport>
					</div>
				</div>
			{/snippet}
		</BitsSelect.Content>
	</BitsSelect.Portal>
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
		font-size: 0.865rem;

		&:focus-visible {
			border-color: var(--border-accented);
		}

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

		.select-item {
			display: flex;
			align-items: center;
			justify-content: space-between;
			width: 100%;
			box-sizing: border-box;
			padding: 0.625rem;
			border: none;
			border-radius: 0.5rem;
			outline: none;
			background: transparent;
			color: var(--text);
			cursor: pointer;

			&[data-highlighted] {
				background: var(--surface);
			}

			.item-icons {
				display: flex;
				align-items: center;
				gap: 0.5rem;

				.delete-button {
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
			}
		}

		.create-new-category-item {
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

		&.creating-new-category {
			.select-item:not(.create-new-category-item)[data-highlighted] {
				background: transparent;
			}
		}
	}
</style>
