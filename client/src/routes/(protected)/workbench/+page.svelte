<script lang="ts">
	import CategoryFilter from '$lib/components/archive/category-filter.svelte';
	import TaskDialog from '$lib/components/taskCreation/task-dialog.svelte';
	import Button from '$lib/components/shared/button.svelte';
	import InfiniteScrollContainer from '$lib/components/shared/infinite-scroll-container.svelte';
	import Task from '$lib/components/taskDisplay/task.svelte';
	import { taskService } from '$lib/services/task-service.svelte';
	import { categoryStore } from '$lib/stores/category-store.svelte';
	import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
	import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
	import type { DatedUntrashedDoneTaskResponse } from '$lib/types/responses/dated-untrashed-done-task-response';
	import type { DatedUntrashedOpenTaskResponse } from '$lib/types/responses/dated-untrashed-open-task-response';
	import type { TaskDragAndDropRequest } from '$lib/types/requests/task-drag-and-drop-request';
	import { getLocalTimeZone, today } from '@internationalized/date';
	import { CheckCheck, Ghost, MousePointer2, Plus, RotateCcw, SquarePen, Trash2, XCircle } from '@lucide/svelte';
	import { dndzone, SHADOW_ITEM_MARKER_PROPERTY_NAME as isShadowTask, TRIGGERS, type DndEvent } from 'svelte-dnd-action';

	type TaskDragItem = (DatedUntrashedOpenTaskResponse | DatedUntrashedDoneTaskResponse) & {
		[isShadowTask]?: boolean;
	};

	const maximumBatchSize = 50;
	const todayDate = today(getLocalTimeZone());

	let taskDialogOpen = $state(false);
	let taskDialogTask = $state<DatedUntrashedOpenTaskResponse | DatedUntrashedDoneTaskResponse>();
	let dragging = $state(false);
	let selectionEnabled = $state(false);
	let selectedTaskIds = $state<number[]>([]);
	let unselectedCategoryIds = $state<number[]>([]);
	let openTaskDragItems = $state<TaskDragItem[]>([...datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks]);
	let doneTaskDragItems = $state<TaskDragItem[]>([...datedUntrashedDoneTaskStore.datedUntrashedDoneTasks]);
	const allSelectedTasksOpen = $derived(areAllSelectedTasksOpen());
	const allSelectedTasksDone = $derived(areAllSelectedTasksDone());
	const batchSizeExceeded = $derived(selectedTaskIds.length > maximumBatchSize);

	$effect.pre(() => {
		if (dragging) return;
		openTaskDragItems = [...datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks];
		doneTaskDragItems = [...datedUntrashedDoneTaskStore.datedUntrashedDoneTasks];
	});

	function handleConsider(event: CustomEvent<DndEvent<TaskDragItem>>, done: boolean): void {
		dragging = true;
		setTaskDragItems(event.detail.items, done);
	}

	async function handleFinalize(event: CustomEvent<DndEvent<TaskDragItem>>, done: boolean): Promise<void> {
		const taskDragItems = event.detail.items;
		setTaskDragItems(taskDragItems, done);

		const taskDragItemId = Number(event.detail.info.id);
		const wasTaskDragItemDroppedIntoZone = event.detail.info.trigger === TRIGGERS.DROPPED_INTO_ZONE;
		const taskExists =
			datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks.some((task) => task.id === taskDragItemId) ||
			datedUntrashedDoneTaskStore.datedUntrashedDoneTasks.some((task) => task.id === taskDragItemId);
		if (wasTaskDragItemDroppedIntoZone && taskExists) {
			const taskDragItemIndex = taskDragItems.findIndex((task) => task.id === taskDragItemId);
			const request: TaskDragAndDropRequest = {
				date: todayDate,
				done,
				trashed: false,
				previousTaskId: taskDragItems[taskDragItemIndex - 1]?.id
			};
			await taskService.dragAndDropTask(taskDragItemId, request);
		}

		dragging = false;
	}

	function setTaskDragItems(taskDragItems: TaskDragItem[], done: boolean): void {
		if (done) {
			doneTaskDragItems = taskDragItems;
		} else {
			openTaskDragItems = taskDragItems;
		}
	}

	function getTaskDragItemKey(task: TaskDragItem): string {
		return `${task.id}-${task[isShadowTask] ? 'shadow' : 'item'}`;
	}

	function openCreateTaskDialog(): void {
		taskDialogTask = undefined;
		taskDialogOpen = true;
	}

	function openEditTaskDialog(task: DatedUntrashedOpenTaskResponse | DatedUntrashedDoneTaskResponse): void {
		taskDialogTask = task;
		taskDialogOpen = true;
	}

	function getTaskActions(task: DatedUntrashedOpenTaskResponse | DatedUntrashedDoneTaskResponse) {
		return [
			{ label: 'Edit', icon: SquarePen, onclick: () => openEditTaskDialog(task) },
			{ label: 'Trash', icon: Trash2, onclick: () => taskService.trashTask(task.id) }
		];
	}

	function cancelTaskSelection(): void {
		selectionEnabled = false;
		selectedTaskIds = [];
	}

	async function completeSelectedTasks(): Promise<void> {
		if (!allSelectedTasksOpen || batchSizeExceeded) return;
		const completed = await taskService.completeTasks(selectedTaskIds);
		if (completed) cancelTaskSelection();
	}

	async function openSelectedTasks(): Promise<void> {
		if (!allSelectedTasksDone || batchSizeExceeded) return;
		const opened = await taskService.openTasks(selectedTaskIds);
		if (opened) cancelTaskSelection();
	}

	async function trashSelectedTasks(): Promise<void> {
		if (selectedTaskIds.length === 0 || batchSizeExceeded) return;
		const trashed = await taskService.trashTasks(selectedTaskIds);
		if (trashed) cancelTaskSelection();
	}

	function updateTaskSelection(taskId: number, selected: boolean): void {
		if (selected && !selectedTaskIds.includes(taskId)) {
			selectedTaskIds = [...selectedTaskIds, taskId];
		} else if (!selected) {
			selectedTaskIds = selectedTaskIds.filter((selectedTaskId) => selectedTaskId !== taskId);
		}
	}

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

	function areAllSelectedTasksOpen(): boolean {
		const atLeastOneTaskSelected = selectedTaskIds.length > 0;
		const everySelectedTaskIsOpen = selectedTaskIds.every((taskId) =>
			datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks.some((task) => task.id === taskId)
		);
		return atLeastOneTaskSelected && everySelectedTaskIsOpen;
	}

	function areAllSelectedTasksDone(): boolean {
		const atLeastOneTaskSelected = selectedTaskIds.length > 0;
		const everySelectedTaskIsDone = selectedTaskIds.every((taskId) =>
			datedUntrashedDoneTaskStore.datedUntrashedDoneTasks.some((task) => task.id === taskId)
		);
		return atLeastOneTaskSelected && everySelectedTaskIsDone;
	}
</script>

<svelte:head>
	<title>Workbench</title>
</svelte:head>

<div class="frame">
	<div class="left-column-container">
		<div class="controls-section">
			{#if selectionEnabled}
				<Button onclick={cancelTaskSelection}>
					Cancel
					{#snippet trailingIcon()}<XCircle size={17} aria-hidden="true" />{/snippet}
				</Button>
				<Button disabled={!allSelectedTasksOpen || batchSizeExceeded} onclick={completeSelectedTasks}>
					Complete
					{#snippet trailingIcon()}<CheckCheck size={17} aria-hidden="true" />{/snippet}
				</Button>
				<Button disabled={!allSelectedTasksDone || batchSizeExceeded} onclick={openSelectedTasks}>
					Open
					{#snippet trailingIcon()}<RotateCcw size={17} aria-hidden="true" />{/snippet}
				</Button>
				<Button disabled={selectedTaskIds.length === 0 || batchSizeExceeded} onclick={trashSelectedTasks}>
					Trash
					{#snippet trailingIcon()}<Trash2 size={17} aria-hidden="true" />{/snippet}
				</Button>
				<span class:limit-exceeded={batchSizeExceeded} class="selection-count">
					{selectedTaskIds.length} / {maximumBatchSize}
				</span>
			{:else}
				<Button onclick={openCreateTaskDialog}>
					Add Task
					{#snippet trailingIcon()}<Plus size={17} strokeWidth={2.5} />{/snippet}
				</Button>
				<Button onclick={() => selectionEnabled = true}>
					Select Tasks
					{#snippet trailingIcon()}<MousePointer2 size={17} aria-hidden="true" />{/snippet}
				</Button>
			{/if}
		</div>
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
		<InfiniteScrollContainer
			hasMore={datedUntrashedOpenTaskStore.todayTasksHasMore}
			disabled={dragging}
			loadMore={loadMoreOpenTasks}
		>
			<div
				class="task-item-drop-zone"
				aria-label="Open tasks"
				use:dndzone={{
					items: openTaskDragItems,
					dragDisabled: selectionEnabled,
					dropTargetStyle: { outline: 'none' }
				}}
				onconsider={(event) => handleConsider(event, false)}
				onfinalize={(event) => handleFinalize(event, false)}
			>
				{#each openTaskDragItems as task (getTaskDragItemKey(task))}
					<Task
						{task}
						actions={getTaskActions(task)}
						isShadowItem={Boolean(task[isShadowTask])}
						{selectionEnabled}
						selected={selectedTaskIds.includes(task.id)}
						onSelectedChange={updateTaskSelection}
					/>
				{/each}
			</div>
		</InfiniteScrollContainer>
	</section>
	<section class="done-tasks-section">
		<h1>Done</h1>
		<InfiniteScrollContainer hasMore={datedUntrashedDoneTaskStore.hasMore} disabled={dragging} loadMore={loadMoreDoneTasks}>
			<div
				class="task-item-drop-zone"
				aria-label="Done tasks"
				use:dndzone={{
					items: doneTaskDragItems,
					dragDisabled: selectionEnabled,
					dropTargetStyle: { outline: 'none' }
				}}
				onconsider={(event) => handleConsider(event, true)}
				onfinalize={(event) => handleFinalize(event, true)}
			>
				{#each doneTaskDragItems as task (getTaskDragItemKey(task))}
					<Task
						{task}
						actions={getTaskActions(task)}
						isShadowItem={Boolean(task[isShadowTask])}
						{selectionEnabled}
						selected={selectedTaskIds.includes(task.id)}
						onSelectedChange={updateTaskSelection}
					/>
				{/each}
			</div>
		</InfiniteScrollContainer>
	</section>
</div>

<TaskDialog bind:open={taskDialogOpen} bind:task={taskDialogTask} defaultDate={todayDate} />

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

		.task-item-drop-zone {
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

	/*noinspection CssUnusedSymbol*/
	:global(#dnd-action-dragged-el) {
		opacity: 1 !important;
	}
</style>
