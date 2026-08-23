<script lang="ts">
	import { Plus } from '@lucide/svelte';
	import Button from '$lib/components/shared/button.svelte';
	import CreateTaskDialog from '$lib/components/taskCreation/create-task-dialog.svelte';
	import Task from '$lib/components/taskDisplay/task.svelte';
	import { undatedTaskStore } from '$lib/stores/undated-tasks-store.svelte';

	let createTaskDialogOpen = $state(false);
</script>

<svelte:head>
	<title>Backlog</title>
</svelte:head>

<div class="frame">
	<div class="left-column-container">
		<div class="controls-section">
			<Button onclick={() => (createTaskDialogOpen = true)}>
				Add Task
				{#snippet trailingIcon()}<Plus size={17} strokeWidth={2.5} />{/snippet}
			</Button>
		</div>
		<div class="no-date-section">
			{#each undatedTaskStore.undatedTasks as task (task.id)}
				<Task {task} />
			{/each}
		</div>
	</div>
	<div class="date-section"></div>
	<div class="today-section"></div>
</div>

<CreateTaskDialog bind:open={createTaskDialogOpen} />

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
			border: 1px solid var(--border);
			border-radius: 0.75rem;
		}

		.today-section {
			display: flex;
			flex-direction: column;
			height: 100%;
			flex: 1;
			border: 1px solid var(--border);
			border-radius: 0.75rem;
		}
	}
</style>
