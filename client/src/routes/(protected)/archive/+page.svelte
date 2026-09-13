<script lang="ts">
	import InfiniteScrollContainer from '$lib/components/shared/infinite-scroll-container.svelte';
	import Task from '$lib/components/taskDisplay/task.svelte';
	import { taskService } from '$lib/services/task-service.svelte';
	import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
	import { trashedTaskStore } from '$lib/stores/trashed-task-store.svelte';
	import { getDateLabel } from '$lib/utils/date';
	import { parseDate } from '@internationalized/date';

	const dateKeyToDatedUntrashedDoneTasksRecord = $derived(
		Object.fromEntries(Map.groupBy(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks, (task) => task.date.toString()))
	);

	function loadMoreDatedUntrashedDoneTasks(): Promise<void> | undefined {
		const lastLoadedTask = datedUntrashedDoneTaskStore.datedUntrashedDoneTasks.at(-1);
		if (lastLoadedTask !== undefined) {
			return taskService.fetchDatedUntrashedDoneTasks({ date: lastLoadedTask.date, rank: lastLoadedTask.rank });
		}
	}

	function loadMoreTrashedTasks(): Promise<void> | undefined {
		const lastLoadedTask = trashedTaskStore.trashedTasks.at(-1);
		if (lastLoadedTask !== undefined) {
			return taskService.fetchTrashedTasks({ trashedDate: lastLoadedTask.trashedDate, id: lastLoadedTask.id });
		}
	}
</script>

<svelte:head>
	<title>Archive</title>
</svelte:head>

<div class="frame">
	<div class="left-column-container">
		<div class="controls-section" aria-hidden="true"></div>
		<section class="category-selection-section">
			<h1>Category Selection</h1>
		</section>
		<section class="time-scope-section">
			<h1>Time Scope</h1>
		</section>
		<section class="activity-section">
			<h1>Activity</h1>
		</section>
	</div>
	<section class="completed-tasks-section">
		<h1>Completed Tasks</h1>
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
									isShadowItem={false}
									selectionEnabled={false}
									selected={false}
									onSelectedChange={() => {}}
									onedit={() => {}}
								/>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		</InfiniteScrollContainer>
	</section>
	<section class="trash-section">
		<h1>Trash</h1>
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
						isShadowItem={false}
						selectionEnabled={false}
						selected={false}
						onSelectedChange={() => {}}
						onedit={() => {}}
					/>
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

		.left-column-container {
			display: grid;
			grid-template-rows: 3.75rem repeat(3, minmax(0, 1fr));
			min-height: 0;
			gap: 1rem;

			.controls-section {
				box-sizing: border-box;
				padding: 0.75rem;
				border: 1px solid var(--border);
				border-radius: 0.75rem;
			}

			.category-selection-section {
				display: flex;
				flex-direction: column;
				min-height: 0;
				box-sizing: border-box;
				padding: 0.75rem;
				border: 1px solid var(--border);
				border-radius: 0.75rem;
			}

			.time-scope-section {
				display: flex;
				flex-direction: column;
				min-height: 0;
				box-sizing: border-box;
				padding: 0.75rem;
				border: 1px solid var(--border);
				border-radius: 0.75rem;
			}

			.activity-section {
				display: flex;
				flex-direction: column;
				min-height: 0;
				box-sizing: border-box;
				padding: 0.75rem;
				border: 1px solid var(--border);
				border-radius: 0.75rem;
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

			.trashed-task-list {
				display: flex;
				flex-direction: column;
				gap: 0.75rem;
			}
		}
	}
</style>
