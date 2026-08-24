<script lang="ts">
	import { X } from '@lucide/svelte';
	import type { CalendarDate } from '@internationalized/date';
	import { Dialog } from 'bits-ui';
	import Button from '$lib/components/shared/button.svelte';
	import DatePicker from '$lib/components/shared/date-picker.svelte';
	import Input from '$lib/components/shared/input.svelte';
	import CategorySelect from '$lib/components/taskCreation/category-select.svelte';
	import Textarea from '$lib/components/shared/textarea.svelte';
	import { taskService } from '$lib/services/task-service.svelte.js';
	import { createTaskInputSchema } from '$lib/types/inputs/create-task-input';

	type Props = {
		open: boolean;
	};

	let { open = $bindable() }: Props = $props();
	let title = $state('');
	let description = $state('');
	let date = $state<CalendarDate>();
	let categoryId = $state<number>();
	let titleError = $state<string>();
	let descriptionError = $state<string>();

	async function handleCreateTask(): Promise<void> {
		const inputParseResult = createTaskInputSchema.safeParse({
			title,
			description,
			categoryId
		});
		if (!inputParseResult.success) {
			applyValidationErrors(inputParseResult.error.issues);
			return;
		}

		clearValidationErrors();
		if (date) {
			await taskService.createTask({ ...inputParseResult.data, date: date.toString() });
		} else {
			await taskService.createTask(inputParseResult.data);
		}

		title = '';
		description = '';
		date = undefined;
		categoryId = undefined;
		open = false;
	}

	function clearValidationErrors(): void {
		titleError = undefined;
		descriptionError = undefined;
	}

	function applyValidationErrors(issues: Array<{ path: PropertyKey[]; message: string }>): void {
		clearValidationErrors();
		for (const issue of issues) {
			const field = issue.path[0];
			if (field === 'title') titleError = titleError ?? issue.message;
			if (field === 'description') descriptionError = descriptionError ?? issue.message;
		}
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
								<h2 {...titleProps} class="title">Create Task</h2>
							{/snippet}
						</Dialog.Title>
						<Dialog.Close>
							{#snippet child({ props: closeProps })}
								<button {...closeProps} class="close-button" aria-label="Close create task modal">
									<X size={30} strokeWidth={2.5} aria-hidden="true" />
								</button>
							{/snippet}
						</Dialog.Close>
					</div>

					<div class="task-fields">
						<label for="task-title">Title *</label>
						<Input
							id="task-title"
							bind:value={title}
							error={titleError}
							placeholder="None chosen"
							oninput={() => (titleError = undefined)}
						/>

						<label for="task-description">Description</label>
						<Textarea
							id="task-description"
							bind:value={description}
							error={descriptionError}
							placeholder="None chosen"
							oninput={() => (descriptionError = undefined)}
						/>

						<label for="task-date">Date</label>
						<DatePicker id="task-date" bind:value={date} />

						<label for="task-category">Category *</label>
						<CategorySelect id="task-category" onchange={(value) => (categoryId = value)} />
					</div>

					<div class="footer">
						<Button onclick={() => (open = false)}>Cancel</Button>
						<div class="spacer"></div>
						<span>* required</span>
						<Button
							variant="success"
							disabled={categoryId === undefined || taskService.taskCreationLoading}
							onclick={() => void handleCreateTask()}>Create</Button
						>
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
		width: min(32rem, calc(100vw - 2rem));
		box-shadow: 0 1.25rem 4rem rgb(0 0 0 / 30%);
		border: 1px solid var(--border);
		border-radius: 1rem;
		background: var(--background);
		color: var(--text);
		box-sizing: border-box;
		padding: 1.75rem;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;

		.heading {
			display: flex;
			align-items: center;
			justify-content: space-between;

			.title {
				margin: 0;
				font-size: 2rem;
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

		.task-fields {
			display: grid;
			grid-template-columns: 6rem minmax(0, 1fr);
			align-items: start;
			gap: 1rem;
		}

		.footer {
			display: flex;
			align-items: center;
			justify-content: space-between;
			gap: 0.75rem;

			.spacer {
				flex: 1;
			}

			span {
				color: var(--text-subtle);
				font-size: 0.75rem;
			}
		}
	}
</style>
