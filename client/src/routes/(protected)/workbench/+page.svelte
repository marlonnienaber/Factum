<script lang="ts">
	import CategoryFilter from '$lib/components/archive/category-filter.svelte';
	import InfiniteScrollContainer from '$lib/components/shared/infinite-scroll-container.svelte';
	import Task from '$lib/components/taskDisplay/task.svelte';
	import { taskService } from '$lib/services/task-service.svelte';
	import { categoryStore } from '$lib/stores/category-store.svelte';
	import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
	import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
	import { Ghost } from '@lucide/svelte';

	let unselectedCategoryIds = $state<number[]>([]);

	function handleCategorySelectionChange(nextUnselectedCategoryIds: number[]): void {
		unselectedCategoryIds = nextUnselectedCategoryIds;
		const selectedCategoryIds = getSelectedCategoryIds();
		if (selectedCategoryIds.length === 0) {
			datedUntrashedOpenTaskStore.clearTodayPages();
			datedUntrashedDoneTaskStore.reset();
		} else {
			void taskService.fetchDatedUntrashedOpenTodayTasks(undefined, selectedCategoryIds);
			void taskService.fetchDatedUntrashedDoneTodayTasks(undefined, selectedCategoryIds);
		}
	}

	function getSelectedCategoryIds(): number[] {
		return categoryStore.categories
			.filter((category) => !unselectedCategoryIds.includes(category.id))
			.map((category) => category.id);
	}

	function loadMoreOpenTasks(): Promise<void> | undefined {
		const lastLoadedTask = datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks.at(-1);
		if (lastLoadedTask !== undefined) {
			return taskService.fetchDatedUntrashedOpenTodayTasks(lastLoadedTask.rank, getSelectedCategoryIds());
		}
	}

	function loadMoreDoneTasks(): Promise<void> | undefined {
		const lastLoadedTask = datedUntrashedDoneTaskStore.datedUntrashedDoneTasks.at(-1);
		if (lastLoadedTask !== undefined) {
			return taskService.fetchDatedUntrashedDoneTodayTasks(lastLoadedTask.rank, getSelectedCategoryIds());
		}
	}
</script>

<svelte:head>
	<title>Workbench</title>
</svelte:head>

<div class="frame">
	<div class="left-column-container">
		<div class="controls-section"></div>
		<section class="category-selection-section">
			<h1>Category Filter</h1>
			<div class="category-filter-container">
				{#if categoryStore.categories.length === 0}
					<div class="no-categories">
						<Ghost size={40} aria-hidden="true" />
						<span>No Categories Available</span>
					</div>
				{:else}
					<CategoryFilter {unselectedCategoryIds} onSelectionChange={handleCategorySelectionChange} />
				{/if}
			</div>
		</section>
		<section class="progress-section">
			<h1>Progress</h1>
		</section>
	</div>
	<section class="open-tasks-section">
		<h1>Open</h1>
		<InfiniteScrollContainer hasMore={datedUntrashedOpenTaskStore.todayTasksHasMore} loadMore={loadMoreOpenTasks}>
			<div class="task-list" aria-label="Open tasks">
				{#each datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks as task (task.id)}
					<Task {task} actions={[]} isShadowItem={false} selectionEnabled={false} selected={false} onSelectedChange={() => {}} />
				{/each}
			</div>
		</InfiniteScrollContainer>
	</section>
	<section class="done-tasks-section">
		<h1>Done</h1>
		<InfiniteScrollContainer hasMore={datedUntrashedDoneTaskStore.hasMore} loadMore={loadMoreDoneTasks}>
			<div class="task-list" aria-label="Done tasks">
				{#each datedUntrashedDoneTaskStore.datedUntrashedDoneTasks as task (task.id)}
					<Task {task} actions={[]} isShadowItem={false} selectionEnabled={false} selected={false} onSelectedChange={() => {}} />
				{/each}
			</div>
		</InfiniteScrollContainer>
	</section>
</div>

<style lang="scss">
	.frame {
		flex: 1;
		min-height: 0;
		width: 100%;
		display: grid;
		grid-template-columns: repeat(3, minmax(min-content, 1fr));
		box-sizing: border-box;
		padding: 1rem;
		gap: 1rem;
		overflow-x: auto;
		overflow-y: hidden;

		h1 {
			margin: 0;
			font-size: 1.25rem;
			font-weight: 500;
		}

		.task-list {
			display: flex;
			flex-direction: column;
			gap: 0.75rem;
		}

		.left-column-container {
			display: grid;
			grid-template-rows: 3.75rem repeat(2, minmax(0, 1fr));
			min-height: 0;
			gap: 1rem;

			.controls-section {
				display: flex;
				align-items: center;
				gap: 0.75rem;
				box-sizing: border-box;
				padding: 0.75rem;
				border: 1px solid var(--border);
				border-radius: 0.75rem;
			}

			.category-selection-section {
				display: grid;
				grid-template-rows: auto minmax(0, 1fr);
				min-height: 0;
				box-sizing: border-box;
				padding: 0.75rem;
				row-gap: 0.75rem;
				border: 1px solid var(--border);
				border-radius: 0.75rem;

				.category-filter-container {
					display: flex;
					align-items: safe center;
					justify-content: center;
					min-height: 0;

					.no-categories {
						display: flex;
						flex-direction: column;
						align-items: center;
						gap: 0.75rem;
						color: var(--text);
						font-size: 0.875rem;

						span {
							color: var(--text-subtle);
						}
					}
				}
			}

			.progress-section {
				min-height: 0;
				box-sizing: border-box;
				padding: 0.75rem;
				border: 1px solid var(--border);
				border-radius: 0.75rem;
			}
		}

		.open-tasks-section {
			display: flex;
			flex-direction: column;
			height: 100%;
			min-height: 0;
			box-sizing: border-box;
			padding: 0.75rem;
			border: 1px solid var(--border);
			border-radius: 0.75rem;
			gap: 0.75rem;
		}

		.done-tasks-section {
			display: flex;
			flex-direction: column;
			height: 100%;
			min-height: 0;
			box-sizing: border-box;
			padding: 0.75rem;
			border: 1px solid var(--border);
			border-radius: 0.75rem;
			gap: 0.75rem;
		}
	}
</style>
