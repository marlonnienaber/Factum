<script lang="ts">
	import { X } from '@lucide/svelte';
	import { Dialog } from 'bits-ui';
	import Button from '$lib/components/button.svelte';
	import DatePicker from '$lib/components/date-picker.svelte';
	import Input from '$lib/components/input.svelte';
	import CategorySelect from '$lib/components/category-select.svelte';
	import Textarea from '$lib/components/textarea.svelte';

	type Props = {
		open?: boolean;
	};

	let { open = $bindable(false) }: Props = $props();

	let categories = $state([
		{ value: 'work', label: 'Work' },
		{ value: 'personal', label: 'Personal' },
		{ value: 'shopping', label: 'Shopping' }
	]);

	let title = $state('');
	let description = $state('');
	let category = $state<string>();
	const titleError = $derived(title.length > 100 ? 'Maximum character count of 100 violated.' : undefined);
	const descriptionError = $derived(description.length > 500 ? 'Maximum character count of 500 violated.' : undefined);

	function handleCreateTask(): void {
		open = false;
	}

	function addCategory(label: string): void {
		const newCategory = { value: crypto.randomUUID(), label };
		categories = [...categories, newCategory];
		category = newCategory.value;
	}

	function deleteCategory(value: string): void {
		if (category === value) category = undefined;
		categories = categories.filter((category) => category.value !== value);
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay>
			{#snippet child({ props })}
				<div {...props} class="dialog-overlay"></div>
			{/snippet}
		</Dialog.Overlay>
		<Dialog.Content>
			{#snippet child({ props })}
				<div {...props} class="dialog-content">
					<div class="dialog-heading">
						<Dialog.Title>
							{#snippet child({ props: titleProps })}
								<h2 {...titleProps} class="dialog-title">Create Task</h2>
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
						<Input id="task-title" bind:value={title} error={titleError} placeholder="None chosen" />

						<label for="task-description">Description</label>
						<Textarea id="task-description" bind:value={description} error={descriptionError} placeholder="None chosen" />

						<label for="task-date">Date</label>
						<DatePicker />

						<label for="task-category">Category *</label>
						<CategorySelect items={categories} bind:value={category} onadd={addCategory} ondelete={deleteCategory} />

						<div class="form-actions">
							<Button onclick={() => (open = false)}>Cancel</Button>
							<div class="create-action">
								<span><sup>*</sup> required</span>
								<Button variant="success" onclick={handleCreateTask}>Create</Button>
							</div>
						</div>
					</div>
				</div>
			{/snippet}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style lang="scss">
	.dialog-overlay {
		position: fixed;
		inset: 0;
		z-index: 50;
		background: rgb(0 0 0 / 55%);
	}

	.dialog-content {
		position: fixed;
		top: 50%;
		left: 50%;
		z-index: 51;
		width: min(28rem, calc(100vw - 2rem));
		box-sizing: border-box;
		transform: translate(-50%, -50%);
		padding: 1.75rem;
		border: 1px solid var(--border);
		border-radius: 1rem;
		background: var(--background);
		color: var(--text);
		box-shadow: 0 1.25rem 4rem rgb(0 0 0 / 30%);
	}

	.dialog-heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 1.5rem;
	}

	.dialog-title {
		margin: 0;
		font-size: 2rem;
		font-weight: 500;
		line-height: 1.2;
	}

	.close-button {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0.125rem;
		border: 0;
		background: transparent;
		color: var(--text);
		cursor: pointer;
	}

	.task-fields {
		display: grid;
		grid-template-columns: 6rem minmax(0, 1fr);
		align-items: center;
		gap: 1rem;

		label {
			font-size: 0.925rem;
			font-weight: 500;
		}

		.form-actions {
			grid-column: 1 / -1;
			display: flex;
			align-items: center;
			justify-content: space-between;
			margin-top: 0.75rem;
		}

		.create-action {
			display: flex;
			align-items: center;
			gap: 0.75rem;

			span {
				color: var(--text-subtle);
				font-size: 0.75rem;
			}
		}
	}

	@media (max-width: 35rem) {
		.task-fields {
			grid-template-columns: 1fr;
			gap: 0.5rem;
		}
	}
</style>
