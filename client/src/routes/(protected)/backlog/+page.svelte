<script lang="ts">
	import { CheckCheck, MousePointer2, Plus, Trash2, XCircle } from '@lucide/svelte';
	import Button from '$lib/components/shared/button.svelte';
	import InfiniteScrollContainer from '$lib/components/shared/infinite-scroll-container.svelte';
	import TaskDialog from '$lib/components/taskCreation/task-dialog.svelte';
	import Task from '$lib/components/taskDisplay/task.svelte';
	import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
	import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
	import { taskService } from '$lib/services/task-service.svelte';
	import type { TaskDragAndDropRequest } from '$lib/types/requests/task-drag-and-drop-request';
	import { getLocalTimeZone, parseDate, today, type CalendarDate } from '@internationalized/date';
	import type { DatedUntrashedOpenTaskResponse } from '$lib/types/responses/dated-untrashed-open-task-response';
	import type { UndatedTaskResponse } from '$lib/types/responses/undated-task-response';
	import { dndzone, SHADOW_ITEM_MARKER_PROPERTY_NAME as isShadowTask, TRIGGERS, type DndEvent } from 'svelte-dnd-action';

	type TaskDragItem = (DatedUntrashedOpenTaskResponse | UndatedTaskResponse) & {
		[isShadowTask]?: boolean;
	};

	const dropTargetStyle = { outline: 'none' };
	const maximumBatchSize = 50;
	const todayDate = today(getLocalTimeZone());
	const todayDateKey = todayDate.toString();

	let taskDialogOpen = $state(false);
	let dragging = $state(false);
	let selectionEnabled = $state(false);
	let selectedTaskIds = $state<number[]>([]);
	let taskDialogTask = $state<DatedUntrashedOpenTaskResponse | UndatedTaskResponse>();
	let undatedTaskDragItems = $state<TaskDragItem[]>([...undatedTaskStore.undatedTasks]);
	let todayTaskDragItems = $state<TaskDragItem[]>([...datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks]);
	let dateKeyToDatedTaskDragItemsRecord = $state<Record<string, TaskDragItem[]>>(createGroupKeyToDatedTaskItemsRecord());

	$effect.pre(() => {
		if (dragging) return;
		undatedTaskDragItems = [...undatedTaskStore.undatedTasks];
		todayTaskDragItems = [...datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks];
		dateKeyToDatedTaskDragItemsRecord = createGroupKeyToDatedTaskItemsRecord();
	});

	$effect(() => {
		console.log('Selected task IDs:', $state.snapshot(selectedTaskIds));
	});

	function createGroupKeyToDatedTaskItemsRecord(): Record<string, TaskDragItem[]> {
		const dateKeyToTaskDragItemsMap = Map.groupBy(datedUntrashedOpenTaskStore.datedUntrashedOpenNonTodayTasks, (task) =>
			task.date.toString()
		);
		return Object.fromEntries(dateKeyToTaskDragItemsMap);
	}

	function handleConsider(event: CustomEvent<DndEvent<TaskDragItem>>, dateKey?: string): void {
		dragging = true;
		setTaskDragItems(event.detail.items, dateKey);
	}

	async function handleFinalize(event: CustomEvent<DndEvent<TaskDragItem>>, dateKey?: string): Promise<void> {
		const taskDragItems = event.detail.items;
		setTaskDragItems(taskDragItems, dateKey);

		const taskDragItemId = Number(event.detail.info.id);
		const wasTaskDragItemDroppedIntoZoneOfHandler = event.detail.info.trigger === TRIGGERS.DROPPED_INTO_ZONE;
		const doesTaskForTaskDragItemIdExist =
			undatedTaskStore.undatedTasks.some((task) => task.id === taskDragItemId) ||
			datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.some((task) => task.id === taskDragItemId);
		if (wasTaskDragItemDroppedIntoZoneOfHandler && doesTaskForTaskDragItemIdExist) {
			const taskDragItemIndex = taskDragItems.findIndex((task) => task.id === taskDragItemId);
			const request: TaskDragAndDropRequest = {
				date: dateKey ? parseDate(dateKey) : undefined,
				previousTaskId: taskDragItems[taskDragItemIndex - 1]?.id
			};
			await taskService.dragAndDropTask(taskDragItemId, request);
		}

		dragging = false;
	}

	function setTaskDragItems(taskDragItems: TaskDragItem[], dateKey?: string): void {
		if (dateKey === todayDateKey) {
			todayTaskDragItems = taskDragItems;
		} else if (dateKey) {
			dateKeyToDatedTaskDragItemsRecord = {
				...dateKeyToDatedTaskDragItemsRecord,
				[dateKey]: taskDragItems
			};
		} else {
			undatedTaskDragItems = taskDragItems;
		}
	}

	function getTaskDragItemKey(task: TaskDragItem): string {
		return `${task.id}-${task[isShadowTask] ? 'shadow' : 'item'}`;
	}

	function formatDate(date: CalendarDate): string {
		if (date.compare(todayDate) === 0) return 'Today';
		return `${String(date.day).padStart(2, '0')}.${String(date.month).padStart(2, '0')}.${date.year}`;
	}

	function openCreateTaskDialog() {
		taskDialogTask = undefined;
		taskDialogOpen = true;
	}

	function openEditTaskDialog(task: DatedUntrashedOpenTaskResponse | UndatedTaskResponse) {
		taskDialogTask = task;
		taskDialogOpen = true;
	}

	function loadMoreUndatedTasks(): Promise<void> | undefined {
		const lastLoadedTaskRank = undatedTaskStore.undatedTasks.at(-1)?.rank;
		if (lastLoadedTaskRank !== undefined) return taskService.fetchUndatedTasks(lastLoadedTaskRank);
	}

	function loadMoreDatedUntrashedOpenNonTodayTasks(): Promise<void> | undefined {
		const lastLoadedTask = datedUntrashedOpenTaskStore.datedUntrashedOpenNonTodayTasks.at(-1);
		if (lastLoadedTask !== undefined) {
			return taskService.fetchDatedUntrashedOpenNonTodayTasks({ date: lastLoadedTask.date, rank: lastLoadedTask.rank });
		}
	}

	function loadMoreDatedUntrashedOpenTodayTasks(): Promise<void> | undefined {
		const lastLoadedTaskRank = datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks.at(-1)?.rank;
		if (lastLoadedTaskRank !== undefined) return taskService.fetchDatedUntrashedOpenTodayTasks(lastLoadedTaskRank);
	}

	function startTaskSelection(): void {
		selectionEnabled = true;
	}

	function cancelTaskSelection(): void {
		selectionEnabled = false;
		selectedTaskIds = [];
	}

	async function completeSelectedTasks(): Promise<void> {
		const completed = await taskService.completeTasks(selectedTaskIds);
		if (completed) cancelTaskSelection();
	}

	async function trashSelectedTasks(): Promise<void> {
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
</script>

<svelte:head>
	<title>Backlog</title>
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
					onclick={completeSelectedTasks}
				>
					Complete All
					{#snippet trailingIcon()}<CheckCheck size={17} aria-hidden="true" />{/snippet}
				</Button>
				<Button
					variant="danger"
					disabled={selectedTaskIds.length === 0 || selectedTaskIds.length > maximumBatchSize || taskService.taskMutationBlocked}
					onclick={trashSelectedTasks}
				>
					Trash All
					{#snippet trailingIcon()}<Trash2 size={17} aria-hidden="true" />{/snippet}
				</Button>
			{:else}
				<Button onclick={openCreateTaskDialog}>
					Add Task
					{#snippet trailingIcon()}<Plus size={17} strokeWidth={2.5} />{/snippet}
				</Button>
				<Button onclick={startTaskSelection}>
					Select Tasks
					{#snippet trailingIcon()}<MousePointer2 size={17} aria-hidden="true" />{/snippet}
				</Button>
			{/if}
			{#if selectionEnabled}
				<span class:limit-exceeded={selectedTaskIds.length > maximumBatchSize} class="selection-count">
					{selectedTaskIds.length} / {maximumBatchSize}
				</span>
			{/if}
		</div>
		<div class="no-date-section">
			<h1>No Date Assigned</h1>
			<InfiniteScrollContainer
				hasMore={undatedTaskStore.hasMore}
				loading={taskService.undatedTasksLoading}
				disabled={dragging}
				loadMore={loadMoreUndatedTasks}
			>
				<div
					class="task-item-drop-zone"
					aria-label="Tasks without a date"
					use:dndzone={{
						items: undatedTaskDragItems,
						dragDisabled: selectionEnabled || taskService.taskMutationBlocked,
						dropTargetStyle: dropTargetStyle
					}}
					onconsider={handleConsider}
					onfinalize={handleFinalize}
				>
					{#each undatedTaskDragItems as taskDragItem (getTaskDragItemKey(taskDragItem))}
						<Task
							task={taskDragItem}
							isShadowItem={Boolean(taskDragItem[isShadowTask])}
							{selectionEnabled}
							selected={selectedTaskIds.includes(taskDragItem.id)}
							onSelectedChange={updateTaskSelection}
							onedit={() => openEditTaskDialog(taskDragItem)}
						/>
					{/each}
				</div>
			</InfiniteScrollContainer>
		</div>
	</div>
	<div class="date-section">
		<h1>Date Assigned</h1>
		<InfiniteScrollContainer
			hasMore={datedUntrashedOpenTaskStore.nonTodayTasksHasMore}
			loading={taskService.datedUntrashedOpenNonTodayTasksLoading}
			disabled={dragging}
			loadMore={loadMoreDatedUntrashedOpenNonTodayTasks}
		>
			<div class="date-groups">
				{#each Object.entries(dateKeyToDatedTaskDragItemsRecord) as [groupKey, taskDragItems] (groupKey)}
					{@const groupDate = parseDate(groupKey)}
					<div class="date-group">
						<h2>{formatDate(groupDate)}</h2>
						<div
							class="task-item-drop-zone"
							aria-label={`Tasks dated ${formatDate(groupDate)}`}
							use:dndzone={{
								items: taskDragItems,
								dragDisabled: selectionEnabled || taskService.taskMutationBlocked,
								dropTargetStyle: dropTargetStyle
							}}
							onconsider={(event) => handleConsider(event, groupKey)}
							onfinalize={(event) => handleFinalize(event, groupKey)}
						>
							{#each taskDragItems as taskDragItem (getTaskDragItemKey(taskDragItem))}
								<Task
									task={taskDragItem}
									isShadowItem={Boolean(taskDragItem[isShadowTask])}
									{selectionEnabled}
									selected={selectedTaskIds.includes(taskDragItem.id)}
									onSelectedChange={updateTaskSelection}
									onedit={() => openEditTaskDialog(taskDragItem)}
								/>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		</InfiniteScrollContainer>
	</div>
	<div class="today-section">
		<h1>Today</h1>
		<InfiniteScrollContainer
			hasMore={datedUntrashedOpenTaskStore.todayTasksHasMore}
			loading={taskService.datedUntrashedOpenTodayTasksLoading}
			disabled={dragging}
			loadMore={loadMoreDatedUntrashedOpenTodayTasks}
		>
			<div
				class="task-item-drop-zone"
				aria-label="Tasks dated Today"
				use:dndzone={{
					items: todayTaskDragItems,
					dragDisabled: selectionEnabled || taskService.taskMutationBlocked,
					dropTargetStyle: dropTargetStyle
				}}
				onconsider={(event) => handleConsider(event, todayDateKey)}
				onfinalize={(event) => handleFinalize(event, todayDateKey)}
			>
				{#each todayTaskDragItems as taskDragItem (getTaskDragItemKey(taskDragItem))}
					<Task
						task={taskDragItem}
						isShadowItem={Boolean(taskDragItem[isShadowTask])}
						{selectionEnabled}
						selected={selectedTaskIds.includes(taskDragItem.id)}
						onSelectedChange={updateTaskSelection}
						onedit={() => openEditTaskDialog(taskDragItem)}
					/>
				{/each}
			</div>
		</InfiniteScrollContainer>
	</div>
</div>

<TaskDialog bind:open={taskDialogOpen} bind:task={taskDialogTask} />

<style lang="scss">
	.frame {
		flex: 1;
		min-height: 0;
		width: 100%;
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		box-sizing: border-box;
		padding: 1rem;
		gap: 1rem;
		overflow: hidden;

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

		.task-item-drop-zone {
			display: flex;
			flex-direction: column;
			gap: 0.75rem;
		}

		.left-column-container {
			display: flex;
			flex-direction: column;
			height: 100%;
			min-height: 0;
			flex: 1;
			gap: 1rem;

			.controls-section {
				display: flex;
				align-items: center;
				gap: 0.75rem;
				width: 100%;
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

			.no-date-section {
				display: flex;
				flex-direction: column;
				width: 100%;
				flex: 1;
				min-height: 0;
				box-sizing: border-box;
				padding: 0.75rem;
				border: 1px solid var(--border);
				border-radius: 0.75rem;
				gap: 0.75rem;
			}
		}

		.date-section {
			display: flex;
			flex-direction: column;
			height: 100%;
			min-height: 0;
			flex: 1;
			box-sizing: border-box;
			padding: 0.75rem;
			border: 1px solid var(--border);
			border-radius: 0.75rem;
			gap: 0.75rem;

			.date-groups {
				display: flex;
				flex-direction: column;
				gap: 0.75rem;
			}

			.date-group {
				display: flex;
				flex-direction: column;
				gap: 0.75rem;

				&:last-child {
					flex: 1;

					.task-item-drop-zone {
						flex: 1;
					}
				}
			}
		}

		.today-section {
			display: flex;
			flex-direction: column;
			height: 100%;
			min-height: 0;
			flex: 1;
			box-sizing: border-box;
			padding: 0.75rem;
			border: 1px solid var(--border);
			border-radius: 0.75rem;
			gap: 0.75rem;

			.task-item-drop-zone {
				flex: 1;
				min-height: 0;
			}
		}
	}

	/*noinspection CssUnusedSymbol*/
	:global(#dnd-action-dragged-el) {
		opacity: 1 !important;
	}
</style>
