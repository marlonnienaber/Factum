<script lang="ts">
	import InfiniteScrollContainer from '$lib/components/shared/infinite-scroll-container.svelte';
	import Button from '$lib/components/shared/button.svelte';
	import CategoryFilter from '$lib/components/archive/category-filter.svelte';
	import TimeFilter from '$lib/components/archive/time-filter.svelte';
	import Popover from '$lib/components/shared/popover.svelte';
	import TaskDialog from '$lib/components/taskCreation/task-dialog.svelte';
	import Task from '$lib/components/taskDisplay/task.svelte';
	import { taskService } from '$lib/services/task-service.svelte';
	import { categoryStore } from '$lib/stores/category-store.svelte';
	import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
	import { trashedTaskStore } from '$lib/stores/trashed-task-store.svelte';
	import { getDateLabel } from '$lib/utils/date';
	import { parseDate, type DateValue } from '@internationalized/date';
	import type { DatedUntrashedDoneTaskResponse } from '$lib/types/responses/dated-untrashed-done-task-response';
	import type { TrashedTaskResponse } from '$lib/types/responses/trashed-task-response';
	import { Info, MousePointer2, RotateCcw, SquarePen, Trash2, XCircle } from '@lucide/svelte';

	const maximumBatchSize = 50;

	let taskDialogOpen = $state(false);
	let taskDialogTask = $state<DatedUntrashedDoneTaskResponse | TrashedTaskResponse>();
	let selectionEnabled = $state(false);
	let selectedTaskIds = $state<number[]>([]);
	let unselectedCategoryIds = $state<number[]>([]);
	let startDate = $state<DateValue>();
	let endDate = $state<DateValue>();
	const dateKeyToDatedUntrashedDoneTasksRecord = $derived(createDateKeyToDatedUntrashedDoneTasksRecord());

	function createDateKeyToDatedUntrashedDoneTasksRecord(): Record<string, DatedUntrashedDoneTaskResponse[]> {
		const dateKeyToDatedUntrashedDoneTasksMap = Map.groupBy(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks, (task) =>
			task.date.toString()
		);
		return Object.fromEntries(dateKeyToDatedUntrashedDoneTasksMap);
	}

	function loadMoreDatedUntrashedDoneTasks(): Promise<void> | undefined {
		const lastLoadedTask = datedUntrashedDoneTaskStore.datedUntrashedDoneTasks.at(-1);
		if (lastLoadedTask !== undefined) {
			const cursor = { date: lastLoadedTask.date, rank: lastLoadedTask.rank };
			const selectedCategoryIds = getSelectedCategoryIds();
			return taskService.fetchDatedUntrashedDoneTasks(cursor, selectedCategoryIds, startDate, endDate);
		}
	}

	function loadMoreTrashedTasks(): Promise<void> | undefined {
		const lastLoadedTask = trashedTaskStore.trashedTasks.at(-1);
		if (lastLoadedTask !== undefined) {
			const selectedCategoryIds = getSelectedCategoryIds();
			return taskService.fetchTrashedTasks(lastLoadedTask.rank, selectedCategoryIds, startDate, endDate);
		}
	}

	function handleCategorySelectionChange(nextUnselectedCategoryIds: number[]): void {
		unselectedCategoryIds = nextUnselectedCategoryIds;
		const selectedCategoryIds = getSelectedCategoryIds();
		if (selectedCategoryIds.length === 0) {
			datedUntrashedDoneTaskStore.reset();
			trashedTaskStore.reset();
		} else {
			void taskService.fetchDatedUntrashedDoneTasks(undefined, selectedCategoryIds, startDate, endDate);
			void taskService.fetchTrashedTasks(undefined, selectedCategoryIds, startDate, endDate);
		}
	}

	function handleTimeSelectionChange(nextStartDate: DateValue | undefined, nextEndDate: DateValue | undefined): void {
		startDate = nextStartDate;
		endDate = nextEndDate;
		const selectedCategoryIds = getSelectedCategoryIds();
		if (selectedCategoryIds.length === 0) {
			datedUntrashedDoneTaskStore.reset();
			trashedTaskStore.reset();
		} else {
			void taskService.fetchDatedUntrashedDoneTasks(undefined, selectedCategoryIds, startDate, endDate);
			void taskService.fetchTrashedTasks(undefined, selectedCategoryIds, startDate, endDate);
		}
	}

	function openEditTaskDialog(task: DatedUntrashedDoneTaskResponse | TrashedTaskResponse): void {
		taskDialogTask = task;
		taskDialogOpen = true;
	}

	function getTaskActions(task: DatedUntrashedDoneTaskResponse | TrashedTaskResponse, trashed: boolean) {
		const editAction = { label: 'Edit', icon: SquarePen, onclick: () => openEditTaskDialog(task) };
		if (trashed) {
			return [
				editAction,
				{
					label: 'Restore',
					icon: RotateCcw,
					variant: 'success' as const,
					disabled: taskService.taskMutationBlocked,
					onclick: () => taskService.restoreTask(task.id)
				}
			];
		}
		return [
			editAction,
			{
				label: 'Trash',
				icon: Trash2,
				variant: 'danger' as const,
				disabled: taskService.taskMutationBlocked,
				onclick: () => taskService.trashTask(task.id)
			}
		];
	}

	function getSelectedCategoryIds(): number[] {
		return categoryStore.categories
			.filter((category) => !unselectedCategoryIds.includes(category.id))
			.map((category) => category.id);
	}

	function startTaskSelection(): void {
		selectionEnabled = true;
	}

	function cancelTaskSelection(): void {
		selectionEnabled = false;
		selectedTaskIds = [];
	}

	async function openSelectedTasks(): Promise<void> {
		const opened = await taskService.openTasks(selectedTaskIds);
		if (opened) cancelTaskSelection();
	}

	async function trashSelectedTasks(): Promise<void> {
		const trashed = await taskService.trashTasks(selectedTaskIds);
		if (trashed) cancelTaskSelection();
	}

	function emptyTrash(): void {
		void taskService.deleteTrashedTasks();
	}

	function updateTaskSelection(taskId: number, selected: boolean): void {
		if (selected && !selectedTaskIds.includes(taskId)) {
			selectedTaskIds = [...selectedTaskIds, taskId];
		} else if (!selected) {
			selectedTaskIds = selectedTaskIds.filter((selectedTaskId) => selectedTaskId !== taskId);
		}
	}
</script>

<svelte:head>
	<title>Archive</title>
</svelte:head>

<div class="frame">
	<div class="left-column-container">
		<div class="controls-section">
			{#if selectionEnabled}
				<Button onclick={cancelTaskSelection}>
					Cancel Selection
					{#snippet trailingIcon()}<XCircle size={17} aria-hidden="true" />{/snippet}
				</Button>
				<Button
					variant="success"
					disabled={selectedTaskIds.length === 0 || selectedTaskIds.length > maximumBatchSize || taskService.taskMutationBlocked}
					onclick={openSelectedTasks}
				>
					Open All
					{#snippet trailingIcon()}<RotateCcw size={17} aria-hidden="true" />{/snippet}
				</Button>
				<Button
					variant="danger"
					disabled={selectedTaskIds.length === 0 || selectedTaskIds.length > maximumBatchSize || taskService.taskMutationBlocked}
					onclick={trashSelectedTasks}
				>
					Trash All
					{#snippet trailingIcon()}<Trash2 size={17} aria-hidden="true" />{/snippet}
				</Button>
				<span class:limit-exceeded={selectedTaskIds.length > maximumBatchSize} class="selection-count">
					{selectedTaskIds.length} / {maximumBatchSize}
				</span>
			{:else}
				<Button onclick={startTaskSelection}>
					Select Tasks
					{#snippet trailingIcon()}<MousePointer2 size={17} aria-hidden="true" />{/snippet}
				</Button>
				<Button
					variant="danger"
					disabled={trashedTaskStore.trashedTasks.length === 0 || taskService.taskMutationBlocked}
					onclick={emptyTrash}
				>
					Empty Trash
					{#snippet trailingIcon()}<Trash2 size={17} aria-hidden="true" />{/snippet}
				</Button>
			{/if}
		</div>
		<section class="category-selection-section">
			<h1>Category Filter</h1>
			<div class="category-filter-container">
				<CategoryFilter {unselectedCategoryIds} onSelectionChange={handleCategorySelectionChange} />
			</div>
		</section>
		<section class="time-scope-section">
			<h1>Date Filter</h1>
			<div class="time-scope-container">
				<TimeFilter onSelectionChange={handleTimeSelectionChange} />
			</div>
		</section>
	</div>
	<section class="completed-tasks-section">
		<h1>Completed</h1>
		<InfiniteScrollContainer
			hasMore={datedUntrashedDoneTaskStore.hasMore}
			loading={taskService.datedUntrashedDoneTasksLoading}
			disabled={false}
			loadMore={loadMoreDatedUntrashedDoneTasks}
		>
			<div class="date-groups">
				{#each Object.entries(dateKeyToDatedUntrashedDoneTasksRecord) as [groupKey, tasks] (groupKey)}
					{@const groupDate = parseDate(groupKey)}
					<div class="date-group">
						<h2>{getDateLabel(groupDate)}</h2>
						<div class="task-list" aria-label={`Completed tasks dated ${getDateLabel(groupDate)}`}>
							{#each tasks as task (task.id)}
								<Task
									{task}
									actions={getTaskActions(task, false)}
									isShadowItem={false}
									{selectionEnabled}
									selected={selectedTaskIds.includes(task.id)}
									onSelectedChange={updateTaskSelection}
								/>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		</InfiniteScrollContainer>
	</section>
	<section class="trash-section">
		<div class="trash-heading">
			<h1>Trashed</h1>
			<Popover openOnHover side="right" triggerAriaLabel="About automatic task deletion">
				{#snippet trigger()}<Info size={17} aria-hidden="true" />{/snippet}
				<span class="trash-information">Trashed tasks are automatically deleted every day at midnight German time.</span>
			</Popover>
		</div>
		<InfiniteScrollContainer
			hasMore={trashedTaskStore.hasMore}
			loading={taskService.trashedTasksLoading}
			disabled={false}
			loadMore={loadMoreTrashedTasks}
		>
			<div class="trashed-task-list" aria-label="Trashed tasks">
				{#each trashedTaskStore.trashedTasks as task (task.id)}
					<Task
						{task}
						actions={getTaskActions(task, true)}
						isShadowItem={false}
						selectionEnabled={false}
						selected={false}
						displayDate
						onSelectedChange={() => {}}
					/>
				{/each}
			</div>
		</InfiniteScrollContainer>
	</section>
</div>

<TaskDialog bind:open={taskDialogOpen} bind:task={taskDialogTask} />

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
			font-size: 1.25rem;
			margin: 0;
			font-weight: 500;
		}

		h2 {
			font-size: 1rem;
			margin: 0;
			font-weight: 500;
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

				.selection-count {
					margin-left: auto;
					color: var(--text-subtle);
					font-size: 0.865rem;

					&.limit-exceeded {
						color: var(--danger);
					}
				}
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
				}
			}

			.time-scope-section {
				display: grid;
				grid-template-rows: auto minmax(0, 1fr);
				min-height: 0;
				box-sizing: border-box;
				padding: 0.75rem;
				row-gap: 0.75rem;
				border: 1px solid var(--border);
				border-radius: 0.75rem;

				.time-scope-container {
					display: flex;
					align-items: safe center;
					justify-content: center;
					min-height: 0;
					overflow: auto;
				}
			}
		}

		.completed-tasks-section {
			display: flex;
			flex-direction: column;
			height: 100%;
			min-height: 0;
			box-sizing: border-box;
			padding: 0.75rem;
			border: 1px solid var(--border);
			border-radius: 0.75rem;
			gap: 0.75rem;

			.date-groups {
				display: flex;
				flex-direction: column;
				gap: 0.75rem;

				.date-group {
					display: flex;
					flex-direction: column;
					gap: 0.75rem;

					.task-list {
						display: flex;
						flex-direction: column;
						gap: 0.75rem;
					}
				}
			}
		}

		.trash-section {
			display: flex;
			flex-direction: column;
			height: 100%;
			min-height: 0;
			box-sizing: border-box;
			padding: 0.75rem;
			border: 1px solid var(--border);
			border-radius: 0.75rem;
			gap: 0.75rem;

			.trash-heading {
				display: flex;
				align-items: center;
				gap: 0.375rem;

				.trash-information {
					font-size: 0.75rem;
					white-space: nowrap;
				}
			}

			.trashed-task-list {
				display: flex;
				flex-direction: column;
				gap: 0.75rem;
			}
		}
	}
</style>
