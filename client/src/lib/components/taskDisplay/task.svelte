<script lang="ts">
	import Checkbox from '$lib/components/shared/checkbox.svelte';
	import CheckCircle from '$lib/components/shared/checkcircle.svelte';
	import Button from '$lib/components/shared/button.svelte';
	import Category from '$lib/components/shared/category.svelte';
	import Popover from '$lib/components/shared/popover.svelte';
	import { taskService } from '$lib/services/task-service.svelte';
	import { categoryStore } from '$lib/stores/category-store.svelte';
	import type { DatedUntrashedOpenTaskResponse } from '$lib/types/responses/dated-untrashed-open-task-response';
	import type { UndatedTaskResponse } from '$lib/types/responses/undated-task-response';
	import { Ellipsis, SquarePen, Trash2 } from '@lucide/svelte';

	type Props = {
		task: DatedUntrashedOpenTaskResponse | UndatedTaskResponse;
		isShadowItem?: boolean;
		onedit: () => void;
	};

	const { task, isShadowItem = false, onedit }: Props = $props();
	let checked = $state(false);
	let selected = $state(false);
	let menuOpen = $state(false);
	const category = $derived(categoryStore.categories.find((category) => category.id === task.categoryId));

	function handleEdit(): void {
		menuOpen = false;
		onedit();
	}

	function handleTrash(): void {
		menuOpen = false;
		void taskService.trashTask(task.id);
	}

	function handleCheckedChange(nextChecked: boolean): void {
		if (!nextChecked) return;
		checked = true;
		void taskService.completeTask(task.id);
	}
</script>

<div class="frame" aria-label={task.title} data-is-dnd-shadow-item-hint={isShadowItem}>
	<div class="header">
		<div class="selection-toggle-and-title">
			<CheckCircle checked={selected} onCheckedChange={(checked) => (selected = checked)} ariaLabel={`Select "${task.title}"`} />
			<span class="title">{task.title}</span>
		</div>
		<div class="actions-and-done-toggle">
			<Popover bind:open={menuOpen} side="bottom" align="end" triggerAriaLabel={`More options for "${task.title}"`}>
				{#snippet trigger()}<Ellipsis size={18} aria-hidden="true" />{/snippet}
				<Button onclick={handleEdit}>
					Edit
					{#snippet trailingIcon()}<SquarePen size={17} aria-hidden="true" />{/snippet}
				</Button>
				<Button variant="danger" disabled={taskService.taskMutationBlocked} onclick={handleTrash}>
					Trash
					{#snippet trailingIcon()}<Trash2 size={17} aria-hidden="true" />{/snippet}
				</Button>
			</Popover>
			<Checkbox {checked} onCheckedChange={handleCheckedChange} ariaLabel={`Mark "${task.title}" as completed`} />
		</div>
	</div>
	{#if task.description}
		<span class="description">{task.description}</span>
	{/if}
	{#if category}
		<Category {category} />
	{/if}
</div>

<style lang="scss">
	.frame {
		display: flex;
		flex-direction: column;
		background-color: var(--surface);
		box-sizing: border-box;
		padding: 1rem;
		border-radius: 0.5rem;
		gap: 0.75rem;
		outline: none;

		&[data-is-dnd-shadow-item-hint='true'] {
			opacity: 0;
		}

		.header {
			display: flex;
			justify-content: space-between;
			align-items: center;

			.selection-toggle-and-title {
				display: flex;
				align-items: center;
				gap: 0.5rem;

				.title {
					line-height: 1rem;
					font-size: 1rem;
				}
			}

			.actions-and-done-toggle {
				display: flex;
				align-items: center;
				gap: 0.5rem;
			}
		}

		.description {
			font-size: 0.75rem;
		}
	}
</style>
