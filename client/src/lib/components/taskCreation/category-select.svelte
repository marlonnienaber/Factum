<script lang="ts">
	import { categoryStore } from '$lib/stores/category-store.svelte.js';
	import { categoryService } from '$lib/services/category-service.svelte.js';
	import Popover from '$lib/components/shared/popover.svelte';
	import { createNewCategoryInputSchema } from '$lib/types/inputs/create-new-category-input';
	import { Check, ChevronDown, Info, Plus, Trash2 } from '@lucide/svelte';
	import { Select as BitsSelect } from 'bits-ui';

	type Props = {
		id: string;
		categoryId?: number;
		error?: string;
		onblur?: () => void;
	};

	let { id, categoryId = $bindable(), error, onblur }: Props = $props();
	const invalid = $derived(error !== undefined);
	const items = $derived(categoryStore.categories.map((category) => ({ value: String(category.id), label: category.title })));
	const value = $derived(categoryId === undefined ? undefined : String(categoryId));
	let open = $state(false);
	const addCategoryItemValue = $derived(getAvailableAddCategoryItemValue(items));
	let addCategoryItemSelected = $state(false);
	let creatingNewCategory = $state(false);
	let newCategoryName = $state('');
	let newCategoryColor = $state('#3b82f6');

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

	function getRandomCategoryColor(): string {
		return `#${Math.floor(Math.random() * 0x1000000)
			.toString(16)
			.padStart(6, '0')}`;
	}

	function handleNewCategoryNameInput(event: Event): void {
		const input = event.currentTarget as HTMLInputElement;
		newCategoryName = input.value;
	}

	async function deleteCategory(event: MouseEvent, itemValue: string): Promise<void> {
		event.stopPropagation();
		const deletedCategoryId = Number(itemValue);
		await categoryService.deleteCategory(deletedCategoryId);
		if (categoryStore.categories.some((category) => category.id === deletedCategoryId)) return;

		if (value === itemValue) {
			categoryId = undefined;
		}
	}

	function handleCategorySelectKeydown(event: KeyboardEvent, bitsUiHandler: unknown): void {
		const isAddCategoryEvent = open && addCategoryItemSelected && (event.key === 'Enter' || event.key === ' ');
		if (isAddCategoryEvent) {
			event.preventDefault();
			newCategoryColor = getRandomCategoryColor();
			creatingNewCategory = true;
		} else if (typeof bitsUiHandler === 'function') {
			bitsUiHandler(event);
		}
	}

	function handleCategorySelectBlur(event: FocusEvent, bitsUiHandler: unknown): void {
		if (typeof bitsUiHandler === 'function') bitsUiHandler(event);
		if (onblur) onblur();
	}

	function handleCategoryValueChange(selectedValue: string): void {
		const selectedCategoryId = Number(selectedValue);
		if (!Number.isInteger(selectedCategoryId)) return;
		categoryId = selectedCategoryId;
	}

	async function handleNewCategoryNameInputKeydown(event: KeyboardEvent): Promise<void> {
		if (event.key !== 'Enter') return;
		event.preventDefault();
		await createNewCategory();
	}

	function handleCreateNewCategoryButtonClicked(event: MouseEvent): void {
		event.stopPropagation();
		void createNewCategory();
	}

	async function createNewCategory(): Promise<void> {
		const inputParseResult = createNewCategoryInputSchema.safeParse({
			title: newCategoryName
		});
		if (!inputParseResult.success) return;

		const oldCategoryIds = new Set(categoryStore.categories.map((category) => category.id));
		await categoryService.createCategory({
			title: inputParseResult.data.title,
			color: newCategoryColor
		});
		const createdCategory = categoryStore.categories.find((category) => !oldCategoryIds.has(category.id));
		if (createdCategory === undefined) return;

		categoryId = createdCategory.id;
		open = false;
	}

	function handleMouseUpOnCreateNewCategoryItem(event: PointerEvent): void {
		event.stopPropagation();

		if (!creatingNewCategory) {
			newCategoryColor = getRandomCategoryColor();
			creatingNewCategory = true;
		}
	}

	$effect(() => {
		if (!open) {
			creatingNewCategory = false;
			addCategoryItemSelected = false;
			newCategoryName = '';
			newCategoryColor = '#3b82f6';
		}
	});
</script>

<div class:invalid class="select-container">
	<BitsSelect.Root type="single" {items} {value} bind:open onValueChange={handleCategoryValueChange}>
		<BitsSelect.Trigger>
			{#snippet child({ props })}
				<button
					{...props}
					{id}
					role="combobox"
					class="select-trigger"
					aria-invalid={invalid}
					onblur={(event) => handleCategorySelectBlur(event, props.onblur)}
					onkeydown={(event) => handleCategorySelectKeydown(event, props.onkeydown)}
				>
					<BitsSelect.Value placeholder="None chosen" />
					<span class="select-action"><ChevronDown size={18} /></span>
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
													class="create-new-category-item"
													onpointerdown={(event) => event.stopPropagation()}
													onpointerup={handleMouseUpOnCreateNewCategoryItem}
												>
													{#if creatingNewCategory}
														<input
															class="category-name-input"
															{@attach (element) => element.focus()}
															value={newCategoryName}
															aria-label="Category name"
															placeholder="Category name"
															maxlength="50"
															oninput={handleNewCategoryNameInput}
															onkeydown={handleNewCategoryNameInputKeydown}
														/>
														<label
															class="color-pill"
															style:background-color={newCategoryColor}
															title={`Category color: ${newCategoryColor}`}
														>
															<span class="color-pill-label">{newCategoryColor}</span>
															<input
																type="color"
																bind:value={newCategoryColor}
																aria-label="Choose category color"
																onclick={(event) => event.stopPropagation()}
															/>
														</label>
														<button
															type="button"
															class="category-create-button"
															aria-label="Create category"
															disabled={newCategoryName === ''}
															onpointerdown={(event) => event.stopPropagation()}
															onpointerup={(event) => event.stopPropagation()}
															onclick={handleCreateNewCategoryButtonClicked}
														>
															<Plus size={18} aria-hidden="true" />
														</button>
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

	{#if invalid}
		<span class="error-indicator-icon">
			<Popover side="right" align="center" openOnHover>
				{#snippet trigger()}
					<span class="error-indicator-popover-trigger" id={`${id}-error-indicator-popover-trigger`}>
						<Info size={18} aria-hidden="true" />
					</span>
				{/snippet}
				<span class="error-indicator-message">{error}</span>
			</Popover>
		</span>
	{/if}
</div>

<style lang="scss">
	.select-container {
		position: relative;

		&.invalid {
			.select-trigger {
				border-color: var(--danger);
				padding-right: calc(0.625rem + 1.125rem + 0.5rem);
			}
		}

		.error-indicator-icon {
			position: absolute;
			top: 50%;
			right: 0.625rem;
			display: flex;
			align-items: center;
			width: 1.125rem;
			height: 1.125rem;
			transform: translateY(-50%);
			color: var(--danger);
			line-height: 0;

			.error-indicator-popover-trigger {
				display: flex;
				align-items: center;
			}
		}
	}

	.error-indicator-message {
		color: var(--text);
		font-size: 0.865rem;
	}

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

		.select-action {
			display: flex;
			color: var(--text);
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
		font-size: 0.865rem;

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
				background: var(--button-ghost);
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
			display: flex;
			align-items: center;
			justify-content: space-between;
			width: 100%;
			box-sizing: border-box;
			padding: 0.625rem;
			border-radius: 0.5rem;
			outline: none;
			background: transparent;
			color: var(--text);
			cursor: pointer;
			gap: 0.5rem;

			&[data-highlighted] {
				background: var(--button-ghost);
			}

			&.editing {
				background: var(--surface-subtle);
				cursor: text;
			}

			.category-name-input {
				flex: 1;
				min-width: 0;
				border: none;
				outline: none;
				background: transparent;
				color: var(--text);
				font: inherit;
			}

			.color-pill {
				position: relative;
				display: flex;
				align-items: center;
				justify-content: center;
				padding: 0.25rem 0.5rem;
				border-radius: 1rem;
				box-sizing: border-box;
				cursor: pointer;

				.color-pill-label {
					color: var(--text);
					font-size: 0.75rem;
				}

				input[type='color'] {
					position: absolute;
					inset: 0;
					width: 100%;
					height: 100%;
					opacity: 0;
					cursor: pointer;
				}
			}

			.category-create-button {
				display: flex;
				padding: 0;
				border: 0;
				outline: none;
				background: transparent;
				color: var(--text);
				cursor: pointer;

				&:disabled {
					opacity: 0.4;
					cursor: not-allowed;
				}
			}
		}

		&.creating-new-category {
			.select-item[data-highlighted] {
				background: transparent;
			}
		}
	}
</style>
