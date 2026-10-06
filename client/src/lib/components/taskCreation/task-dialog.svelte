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
	import { taskInputSchema } from '$lib/types/inputs/task-input';
	import type { CreateTaskRequest } from '$lib/types/requests/create-task-request';
	import type { DatedUntrashedDoneTaskResponse } from '$lib/types/responses/dated-untrashed-done-task-response';
	import type { DatedUntrashedOpenTaskResponse } from '$lib/types/responses/dated-untrashed-open-task-response';
	import type { UndatedTaskResponse } from '$lib/types/responses/undated-task-response';
	import type { TrashedTaskResponse } from '$lib/types/responses/trashed-task-response';

	type Props = {
		open: boolean;
		task?: DatedUntrashedDoneTaskResponse | DatedUntrashedOpenTaskResponse | UndatedTaskResponse | TrashedTaskResponse;
		defaults?: Partial<CreateTaskRequest>;
	};

	let { open = $bindable(), task = $bindable(), defaults }: Props = $props();
	let title = $state('');
	let description = $state('');
	let date = $state<CalendarDate>();
	let categoryId = $state<number>();

	let titleTouched = $state(false);
	let descriptionTouched = $state(false);
	let categoryTouched = $state(false);
	const titleError = $derived(titleTouched ? getValidationError('title') : undefined);
	const descriptionError = $derived(descriptionTouched ? getValidationError('description') : undefined);
	const categoryError = $derived(categoryTouched ? getValidationError('categoryId') : undefined);

	const editing = $derived(task !== undefined);
	const inputParseResult = $derived(taskInputSchema.safeParse({ title, description, categoryId }));
	let saving = $state(false);

	async function handleSubmit(): Promise<void> {
		if (saving || !inputParseResult.success) return;
		saving = true;
		try {
			const request = date ? { ...inputParseResult.data, date } : inputParseResult.data;
			if (task) {
				await taskService.updateTask({ ...request, id: task.id });
			} else {
				await taskService.createTask(request);
			}
			open = false;
			task = undefined;
		} finally {
			saving = false;
		}
	}

	function getValidationError(field: 'title' | 'description' | 'categoryId'): string | undefined {
		if (inputParseResult.success) return undefined;
		return inputParseResult.error.issues.find((issue) => issue.path[0] === field)?.message;
	}

	$effect(() => {
		if (open) {
			if (task) {
				title = task.title;
				description = task.description ?? '';
				date = 'date' in task ? task.date : undefined;
				categoryId = task.categoryId;
			} else {
				title = defaults?.title ?? '';
				description = defaults?.description ?? '';
				date = defaults?.date;
				categoryId = defaults?.categoryId;
			}
			titleTouched = false;
			descriptionTouched = false;
			categoryTouched = false;
		}
	});
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
								<h2 {...titleProps} class="title">{editing ? 'Edit Task' : 'Create Task'}</h2>
							{/snippet}
						</Dialog.Title>
						<Dialog.Close>
							{#snippet child({ props: closeProps })}
								<button {...closeProps} class="close-button" aria-label={`Close ${editing ? 'edit' : 'create'} task dialog`}>
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
							onblur={() => (titleTouched = true)}
						/>

						<label for="task-description">Description</label>
						<Textarea
							id="task-description"
							bind:value={description}
							error={descriptionError}
							placeholder="None chosen"
							onblur={() => (descriptionTouched = true)}
						/>

						<label for="task-date">Date</label>
						<DatePicker id="task-date" bind:value={date} />

						<label for="task-category">Category *</label>
						<CategorySelect id="task-category" bind:categoryId error={categoryError} onblur={() => (categoryTouched = true)} />
					</div>

					<div class="footer">
						<Button onclick={() => (open = false)}>Cancel</Button>
						<div class="spacer"></div>
						<span>* required</span>
						<Button variant="success" disabled={!inputParseResult.success || saving} onclick={handleSubmit}>
							{saving ? (editing ? 'Saving...' : 'Creating...') : editing ? 'Save' : 'Create'}
						</Button>
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
