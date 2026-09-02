<script lang="ts">
	import { Plus } from '@lucide/svelte';
	import Button from '$lib/components/shared/button.svelte';
	import TaskDialog from '$lib/components/taskCreation/task-dialog.svelte';
	import Task from '$lib/components/taskDisplay/task.svelte';
	import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
	import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
	import { taskService } from '$lib/services/task-service.svelte';
	import { getLocalTimeZone, parseDate, today, type CalendarDate } from '@internationalized/date';
	import type { DatedUntrashedOpenTaskResponse } from '$lib/types/responses/dated-untrashed-open-task-response';
	import type { UndatedTaskResponse } from '$lib/types/responses/undated-task-response';
	import { dndzone, SHADOW_ITEM_MARKER_PROPERTY_NAME as isShadowTask, TRIGGERS, type DndEvent } from 'svelte-dnd-action';

	type TaskDragItem = (DatedUntrashedOpenTaskResponse | UndatedTaskResponse) & {
		[isShadowTask]?: boolean;
	};

	const dropTargetStyle = { outline: 'none' };
	const todayDate = today(getLocalTimeZone());
	const todayDateKey = todayDate.toString();

	let taskDialogOpen = $state(false);
	let dragging = $state(false);
	let taskDialogTask = $state<DatedUntrashedOpenTaskResponse | UndatedTaskResponse>();
	let undatedTaskDragItems = $state<TaskDragItem[]>([...undatedTaskStore.undatedTasks]);
	let dateKeyToDatedTaskDragItemsRecord = $state<Record<string, TaskDragItem[]>>(createGroupKeyToDatedTaskItemsRecord());

	$effect.pre(() => {
		if (dragging) return;
		undatedTaskDragItems = [...undatedTaskStore.undatedTasks];
		dateKeyToDatedTaskDragItemsRecord = createGroupKeyToDatedTaskItemsRecord();
	});

	function createGroupKeyToDatedTaskItemsRecord(): Record<string, TaskDragItem[]> {
		const dateKeyToTaskDragItemsMap = Map.groupBy(datedUntrashedOpenTaskStore.datedUntrashedOpenTasks, (task) =>
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
			await taskService.changeTaskDateAndOrRank(
				taskDragItemId,
				dateKey ? parseDate(dateKey) : undefined,
				taskDragItems[taskDragItemIndex - 1]?.rank,
				taskDragItems[taskDragItemIndex + 1]?.rank
			);
		}

		dragging = false;
	}

	function setTaskDragItems(taskDragItems: TaskDragItem[], dateKey?: string): void {
		if (dateKey) {
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
</script>

<svelte:head>
	<title>Backlog</title>
</svelte:head>

<div class="frame">
	<div class="left-column-container">
		<div class="controls-section">
			<Button onclick={openCreateTaskDialog}>
				Add Task
				{#snippet trailingIcon()}<Plus size={17} strokeWidth={2.5} />{/snippet}
			</Button>
		</div>
		<div class="no-date-section">
			<h1>No Date Assigned</h1>
			<div
				class="task-item-drop-zone"
				aria-label="Tasks without a date"
				use:dndzone={{
					items: undatedTaskDragItems,
					dragDisabled: taskService.taskMutationBlocked,
					dropTargetStyle: dropTargetStyle
				}}
				onconsider={handleConsider}
				onfinalize={handleFinalize}
			>
				{#each undatedTaskDragItems as taskDragItem (getTaskDragItemKey(taskDragItem))}
					<Task
						task={taskDragItem}
						isShadowItem={Boolean(taskDragItem[isShadowTask])}
						onedit={() => openEditTaskDialog(taskDragItem)}
					/>
				{/each}
			</div>
		</div>
	</div>
	<div class="date-section">
		<h1>Date Assigned</h1>
		{#each Object.entries(dateKeyToDatedTaskDragItemsRecord) as [groupKey, taskDragItems] (groupKey)}
			{#if groupKey !== todayDateKey}
				{@const groupDate = parseDate(groupKey)}
				<div class="date-group">
					<h2>{formatDate(groupDate)}</h2>
					<div
						class="task-item-drop-zone"
						aria-label={`Tasks dated ${formatDate(groupDate)}`}
						use:dndzone={{
							items: taskDragItems,
							dragDisabled: taskService.taskMutationBlocked,
							dropTargetStyle: dropTargetStyle
						}}
						onconsider={(event) => handleConsider(event, groupKey)}
						onfinalize={(event) => handleFinalize(event, groupKey)}
					>
						{#each taskDragItems as taskDragItem (getTaskDragItemKey(taskDragItem))}
							<Task
								task={taskDragItem}
								isShadowItem={Boolean(taskDragItem[isShadowTask])}
								onedit={() => openEditTaskDialog(taskDragItem)}
							/>
						{/each}
					</div>
				</div>
			{/if}
		{/each}
	</div>
	<div class="today-section">
		<h1>Today</h1>
		<div
			class="task-item-drop-zone"
			aria-label="Tasks dated Today"
			use:dndzone={{
				items: dateKeyToDatedTaskDragItemsRecord[todayDateKey] ?? [],
				dragDisabled: taskService.taskMutationBlocked,
				dropTargetStyle: dropTargetStyle
			}}
			onconsider={(event) => handleConsider(event, todayDateKey)}
			onfinalize={(event) => handleFinalize(event, todayDateKey)}
		>
			{#each dateKeyToDatedTaskDragItemsRecord[todayDateKey] as taskDragItem (getTaskDragItemKey(taskDragItem))}
				<Task
					task={taskDragItem}
					isShadowItem={Boolean(taskDragItem[isShadowTask])}
					onedit={() => openEditTaskDialog(taskDragItem)}
				/>
			{/each}
		</div>
	</div>
</div>

<TaskDialog bind:open={taskDialogOpen} bind:task={taskDialogTask} />

<style lang="scss">
	.frame {
		flex: 1;
		width: 100%;
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		box-sizing: border-box;
		padding: 1rem;
		gap: 1rem;

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
			flex: 1;
			gap: 1rem;

			.controls-section {
				display: flex;
				align-items: center;
				width: 100%;
				box-sizing: border-box;
				padding: 0.75rem;
				border: 1px solid var(--border);
				border-radius: 0.75rem;
			}

			.no-date-section {
				display: flex;
				flex-direction: column;
				width: 100%;
				flex: 1;
				box-sizing: border-box;
				padding: 0.75rem;
				border: 1px solid var(--border);
				border-radius: 0.75rem;
				gap: 0.75rem;

				.task-item-drop-zone {
					flex: 1;
				}
			}
		}

		.date-section {
			display: flex;
			flex-direction: column;
			height: 100%;
			flex: 1;
			box-sizing: border-box;
			padding: 0.75rem;
			border: 1px solid var(--border);
			border-radius: 0.75rem;
			gap: 0.75rem;

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
			flex: 1;
			box-sizing: border-box;
			padding: 0.75rem;
			border: 1px solid var(--border);
			border-radius: 0.75rem;
			gap: 0.75rem;

			.task-item-drop-zone {
				flex: 1;
			}
		}
	}

	/*noinspection CssUnusedSymbol*/
	:global(#dnd-action-dragged-el) {
		opacity: 1 !important;
	}
</style>
