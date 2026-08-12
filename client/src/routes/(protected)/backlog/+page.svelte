<script>
</script>

<svelte:head>
	<title>Backlog</title>
</svelte:head>

<div class="frame">
	<div class="left-column-container">
		<div class="controls-section"></div>
		<div class="no-date-section"></div>
	</div>
	<div class="date-section"></div>
	<div class="today-section"></div>
</div>

<style>
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
				flex-direction: column;
				width: 100%;
				height: 63px;
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
