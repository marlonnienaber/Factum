<script lang="ts">
	import Button from '$lib/components/shared/button.svelte';
	import { taskService } from '$lib/services/task-service.svelte';
	import { X } from '@lucide/svelte';
	import { Dialog } from 'bits-ui';

	type Props = {
		open: boolean;
	};

	let { open = $bindable() }: Props = $props();

	async function confirm(): Promise<void> {
		await taskService.deleteTrashedTasks();
		open = false;
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay>
			{#snippet child({ props })}
				<div {...props} class="overlay"></div>
			{/snippet}
		</Dialog.Overlay>
		<Dialog.Content>
			{#snippet child({ props })}
				<div {...props} class="content">
					<div class="heading">
						<Dialog.Title>
							{#snippet child({ props: titleProps })}
								<h1 {...titleProps} class="title">Empty Trash?</h1>
							{/snippet}
						</Dialog.Title>
						<Dialog.Close>
							{#snippet child({ props: closeProps })}
								<button {...closeProps} class="close-button" aria-label="Close empty trash dialog">
									<X size={30} strokeWidth={2.5} aria-hidden="true" />
								</button>
							{/snippet}
						</Dialog.Close>
					</div>
					<Dialog.Description>
						{#snippet child({ props: descriptionProps })}
							<p {...descriptionProps} class="description">
								Filters are currently applied. Emptying the trash will permanently delete all trashed tasks, including tasks
								hidden by these filters.
							</p>
						{/snippet}
					</Dialog.Description>
					<div class="footer">
						<Button onclick={() => (open = false)}>Cancel</Button>
						<Button variant="danger" disabled={taskService.taskMutationBlocked} onclick={confirm}>Delete All</Button>
					</div>
				</div>
			{/snippet}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style lang="scss">
	.overlay {
		position: fixed;
		inset: 0;
		z-index: 50;
		background: var(--overlay);
	}

	.content {
		position: fixed;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		z-index: 51;
		width: min(28rem, calc(100vw - 2rem));
		box-shadow: 0 1.25rem 4rem rgb(0 0 0 / 30%);
		border: 1px solid var(--border);
		border-radius: 1rem;
		background: var(--background);
		color: var(--text);
		box-sizing: border-box;
		padding: 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;

		.heading {
			display: flex;
			align-items: center;
			justify-content: space-between;

			.title {
				margin: 0;
				font-size: 1.5rem;
				font-weight: 500;
			}

			.close-button {
				display: flex;
				align-items: center;
				justify-content: center;
				padding: 0;
				border: none;
				background: transparent;
				color: var(--text);
				cursor: pointer;
			}
		}

		.description {
			margin: 0;
			font-size: 0.875rem;
			line-height: 1.5;
		}

		.footer {
			display: flex;
			justify-content: space-between;
		}
	}
</style>
