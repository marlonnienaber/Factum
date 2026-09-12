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
		isShadowItem: boolean;
		selectionEnabled: boolean;
		selected: boolean;
		onSelectedChange: (taskId: number, selected: boolean) => void;
		onedit: () => void;
	};

	const { task, isShadowItem, selectionEnabled, selected, onSelectedChange, onedit }: Props = $props();
	let checked = $state(false);
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

	function toggleSelection(): void {
		if (selectionEnabled) onSelectedChange(task.id, !selected);
	}
</script>

<div
	class="frame"
	role="button"
	tabindex="-1"
	data-selection-enabled={selectionEnabled}
	aria-label={task.title}
	data-is-dnd-shadow-item-hint={isShadowItem}
	onpointerup={toggleSelection}
>
	<div class="header">
		<div class="selection-toggle-and-title">
			<div class="selection-toggle" data-enabled={selectionEnabled}>
				<CheckCircle checked={selected} />
			</div>
			<span class="title">{task.title}</span>
		</div>
		{#if !selectionEnabled}
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
		{/if}
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

		&[data-selection-enabled='true'] {
			cursor: pointer;
		}

		&[data-is-dnd-shadow-item-hint='true'] {
			opacity: 0;
		}

		.header {
			display: flex;
			justify-content: space-between;
			align-items: center;
			min-height: 1.125rem;

			.selection-toggle-and-title {
				display: flex;
				align-items: center;

				.selection-toggle {
					width: 0;
					height: 1rem;
					margin-right: 0;
					overflow: hidden;
					opacity: 0;
					transition:
						width 150ms ease,
						margin-right 150ms ease,
						opacity 75ms ease;

					&[data-enabled='true'] {
						width: 1rem;
						margin-right: 0.5rem;
						opacity: 1;
						transition:
							width 150ms ease,
							margin-right 150ms ease,
							opacity 100ms ease 100ms;
					}
				}

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

	@media (prefers-reduced-motion: reduce) {
		.selection-toggle {
			transition: none !important;
		}
	}
</style>
