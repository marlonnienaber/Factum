<script lang="ts">
	import { Plus } from '@lucide/svelte';
	import Button from '$lib/components/shared/button.svelte';
	import TaskDialog from '$lib/components/taskCreation/task-dialog.svelte';
	import Task from '$lib/components/taskDisplay/task.svelte';
	import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
	import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
	import { getLocalTimeZone, today, type CalendarDate } from '@internationalized/date';
	import type { DatedUntrashedOpenTaskResponse } from '$lib/types/responses/dated-untrashed-open-task-response';
	import type { UndatedTaskResponse } from '$lib/types/responses/undated-task-response';

	let taskDialogOpen = $state(false);
	let selectedTask = $state<DatedUntrashedOpenTaskResponse | UndatedTaskResponse>();
	const todayDate = today(getLocalTimeZone());
	const datedTasks = $derived(datedUntrashedOpenTaskStore.datedUntrashedOpenTasks);
	const todayTasks = $derived(
		datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.filter((task) => task.date.compare(todayDate) === 0)
	);

	function openCreateTaskDialog(): void {
		selectedTask = undefined;
		taskDialogOpen = true;
	}

	function openEditTaskDialog(task: DatedUntrashedOpenTaskResponse | UndatedTaskResponse): void {
		selectedTask = task;
		taskDialogOpen = true;
	}

	function formatDate(date: CalendarDate): string {
		if (date.compare(todayDate) === 0) return 'Today';
		return `${String(date.day).padStart(2, '0')}.${String(date.month).padStart(2, '0')}.${date.year}`;
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
			{#each undatedTaskStore.undatedTasks as task (task.id)}
				<Task {task} onedit={() => openEditTaskDialog(task)} />
			{/each}
		</div>
	</div>
	<div class="date-section">
		<h1>Date Assigned</h1>
		{#each datedTasks as task, index (task.id)}
			{#if index === 0 || task.date.compare(datedTasks[index - 1].date) !== 0}
				<h2>{formatDate(task.date)}</h2>
			{/if}
			<Task {task} onedit={() => openEditTaskDialog(task)} />
		{/each}
	</div>
	<div class="today-section">
		<h1>Today</h1>
		{#each todayTasks as task (task.id)}
			<Task {task} onedit={() => openEditTaskDialog(task)} />
		{/each}
	</div>
</div>

<TaskDialog bind:open={taskDialogOpen} bind:task={selectedTask} />

<style lang="scss">
	.frame {
		flex: 1;
		width: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
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
		}
	}
</style>
