<script lang="ts">
	import Checkbox from '$lib/components/shared/checkbox.svelte';
	import CheckCircle from '$lib/components/shared/checkcircle.svelte';
	import Button from '$lib/components/shared/button.svelte';
	import Category from '$lib/components/shared/category.svelte';
	import Popover from '$lib/components/shared/popover.svelte';
	import { taskService } from '$lib/services/task-service.svelte';
	import { categoryStore } from '$lib/stores/category-store.svelte';
	import type { DatedUntrashedDoneTaskResponse } from '$lib/types/responses/dated-untrashed-done-task-response';
	import type { DatedUntrashedOpenTaskResponse } from '$lib/types/responses/dated-untrashed-open-task-response';
	import type { UndatedTaskResponse } from '$lib/types/responses/undated-task-response';
	import type { TrashedTaskResponse } from '$lib/types/responses/trashed-task-response';
	import type { ZonedDateTime } from '@internationalized/date';
	import { Ellipsis, Trash2, type LucideIcon } from '@lucide/svelte';

	type TaskAction = {
		label: string;
		icon: LucideIcon;
		variant?: 'default' | 'success' | 'danger';
		disabled?: boolean;
		onclick: () => void | Promise<void>;
	};

	type Props = {
		task: DatedUntrashedDoneTaskResponse | DatedUntrashedOpenTaskResponse | UndatedTaskResponse | TrashedTaskResponse;
		actions: TaskAction[];
		isShadowItem: boolean;
		selectionEnabled: boolean;
		selected: boolean;
		onSelectedChange: (taskId: number, selected: boolean) => void;
	};

	const { task, actions, isShadowItem, selectionEnabled, selected, onSelectedChange }: Props = $props();
	let menuOpen = $state(false);
	const category = $derived(categoryStore.categories.find((category) => category.id === task.categoryId));
	const trashedDateLabel = $derived('trashedDate' in task ? formatTrashedDate(task.trashedDate) : undefined);

	function handleAction(action: TaskAction): void {
		menuOpen = false;
		void action.onclick();
	}

	function handleCheckedChange(): void {
		void taskService.toggleTaskDone(task.id);
	}

	function toggleSelection(): void {
		if (selectionEnabled) onSelectedChange(task.id, !selected);
	}

	function formatTrashedDate(trashedDate: ZonedDateTime): string {
		const day = String(trashedDate.day).padStart(2, '0');
		const month = String(trashedDate.month).padStart(2, '0');
		const hour = String(trashedDate.hour).padStart(2, '0');
		const minute = String(trashedDate.minute).padStart(2, '0');
		return `${day}.${month}.${trashedDate.year}, ${hour}:${minute}`;
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
				{#if actions.length > 0}
					<Popover bind:open={menuOpen} side="bottom" align="end" triggerAriaLabel={`More options for "${task.title}"`}>
						{#snippet trigger()}<Ellipsis size={18} aria-hidden="true" />{/snippet}
						{#each actions as action (action.label)}
							{@const ActionIcon = action.icon}
							<Button variant={action.variant} disabled={action.disabled} onclick={() => handleAction(action)}>
								{action.label}
								{#snippet trailingIcon()}<ActionIcon size={17} aria-hidden="true" />{/snippet}
							</Button>
						{/each}
					</Popover>
				{/if}
				<Checkbox checked={task.done} onCheckedChange={handleCheckedChange} ariaLabel={`Mark "${task.title}" as completed`} />
			</div>
		{/if}
	</div>
	{#if task.description}
		<span class="description">{task.description}</span>
	{/if}
	{#if category || trashedDateLabel}
		<div class="metadata">
			{#if category}
				<Category {category} />
			{/if}
			{#if trashedDateLabel}
				<span class="trashed-date" aria-label={`Trashed ${trashedDateLabel}`}>
					<Trash2 size={12} aria-hidden="true" />
					{trashedDateLabel}
				</span>
			{/if}
		</div>
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

		.metadata {
			display: flex;
			align-items: center;
			gap: 0.5rem;

			.trashed-date {
				display: flex;
				align-items: center;
				gap: 0.25rem;
				color: var(--text-subtle);
				font-size: 0.75rem;
			}
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.selection-toggle {
			transition: none !important;
		}
	}
</style>
