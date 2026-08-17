<script lang="ts">
	import CreateTaskButton from '$lib/components/create-task-button.svelte';
</script>

<svelte:head>
	<title>Backlog</title>
</svelte:head>

<div class="frame">
	<div class="left-column-container">
		<div class="controls-section">
			<CreateTaskButton />
		</div>
		<div class="no-date-section"></div>
	</div>
	<div class="date-section"></div>
	<div class="today-section"></div>
</div>

<style lang="scss">
	.frame {
		flex: 1;
		width: 100%;
		display: flex;
		justify-content: center;
		align-items: center;
		box-sizing: border-box;
		padding: 1.5rem;
		gap: 1.5rem;

		.left-column-container {
			display: flex;
			flex-direction: column;
			height: 100%;
			flex: 1;
			gap: 1.5rem;

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
				border: 1px solid var(--border);
				border-radius: 0.75rem;
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
