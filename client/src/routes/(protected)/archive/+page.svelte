<script>
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
	</section>
	<section class="trash-section">
		<h1>Trash</h1>
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

			.category-selection-section,
			.time-scope-section,
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

		.completed-tasks-section,
		.trash-section {
			display: flex;
			flex-direction: column;
			height: 100%;
			min-height: 0;
			box-sizing: border-box;
			padding: 0.75rem;
			border: 1px solid var(--border);
			border-radius: 0.75rem;
		}
	}
</style>
