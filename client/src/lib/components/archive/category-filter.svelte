<script lang="ts">
	import Category from '$lib/components/shared/category.svelte';
	import { categoryStore } from '$lib/stores/category-store.svelte';
	import { ArrowLeft, ArrowRight } from '@lucide/svelte';

	type Props = {
		unselectedCategoryIds: number[];
		onSelectionChange: (unselectedCategoryIds: number[]) => void;
	};

	let { unselectedCategoryIds, onSelectionChange }: Props = $props();

	function categoryIncluded(categoryId: number): boolean {
		return !unselectedCategoryIds.includes(categoryId);
	}

	function toggleCategory(categoryId: number): void {
		const nextUnselectedCategoryIds = categoryIncluded(categoryId)
			? [...unselectedCategoryIds, categoryId]
			: unselectedCategoryIds.filter((id) => id !== categoryId);
		onSelectionChange(nextUnselectedCategoryIds);
	}
</script>

<div class="frame">
	<div class="heading-row">
		<span>display</span>
		<div class="separator" aria-hidden="true"></div>
		<span>hide</span>
	</div>
	<div class="category-scroll-area">
		<div class="category-columns">
			<div class="category-list">
				{#each categoryStore.categories as category (category.id)}
					{@const isIncluded = categoryIncluded(category.id)}
					<div class="category-entry">
						<div class="category-spacer"></div>
						{#if isIncluded}
							<button type="button" aria-label={`Exclude ${category.title}`} onclick={() => toggleCategory(category.id)}>
								<Category {category} />
							</button>
						{:else}
							<span class="category-placeholder" aria-hidden="true"><span>{category.title}</span></span>
						{/if}
						<div class="category-spacer">
							{#if isIncluded}
								<button type="button" aria-label={`Exclude ${category.title}`} onclick={() => toggleCategory(category.id)}>
									<ArrowRight size={18} />
								</button>
							{/if}
						</div>
					</div>
				{/each}
			</div>
			<div class="separator" aria-hidden="true"></div>
			<div class="category-list">
				{#each categoryStore.categories as category (category.id)}
					{@const isIncluded = categoryIncluded(category.id)}
					<div class="category-entry">
						<div class="category-spacer">
							{#if !isIncluded}
								<button type="button" aria-label={`Include ${category.title}`} onclick={() => toggleCategory(category.id)}>
									<ArrowLeft size={18} />
								</button>
							{/if}
						</div>
						{#if isIncluded}
							<span class="category-placeholder" aria-hidden="true"><span>{category.title}</span></span>
						{:else}
							<button type="button" aria-label={`Include ${category.title}`} onclick={() => toggleCategory(category.id)}>
								<Category {category} />
							</button>
						{/if}
						<div class="category-spacer"></div>
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>

<style lang="scss">
	.frame {
		display: flex;
		flex-direction: column;
		max-height: 100%;
		min-height: 0;
		overflow: hidden;

		.separator {
			width: 1px;
			background-color: var(--border);
		}

		.heading-row {
			display: grid;
			grid-template-columns: 1fr 0 1fr;

			span {
				padding-bottom: 1.125rem;
				color: var(--text-subtle);
				font-size: 1rem;
				font-weight: 500;
				text-align: center;
			}
		}

		.category-scroll-area {
			min-height: 0;
			overflow-y: auto;

			.category-columns {
				display: grid;
				grid-template-columns: 1fr 0 1fr;

				.category-list {
					display: flex;
					flex-direction: column;
					gap: 1rem;

					.category-entry {
						display: grid;
						grid-template-columns: 2.5rem auto 2.5rem;
						align-items: center;
						justify-items: center;
						width: 100%;

						button {
							display: flex;
							padding: 0;
							border: 0;
							background: transparent;
							color: var(--text);
							cursor: pointer;
						}

						.category-spacer {
							display: flex;
							justify-content: center;
							width: 100%;
						}

						.category-placeholder {
							padding: 0.25rem 0.625rem;
							border-radius: 1rem;
							outline: 1px dashed var(--border);
							outline-offset: -1px;
							font-size: 0.75rem;

							span {
								visibility: hidden;
							}
						}
					}
				}
			}
		}
	}
</style>
