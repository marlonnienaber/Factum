import { fireEvent, render, type RenderResult, waitFor, within } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { categoryStore } from '$lib/stores/category-store.svelte';
import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
import { trashedTaskStore } from '$lib/stores/trashed-task-store.svelte';
import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
import type { DatedUntrashedDoneTaskResponse } from '$lib/types/responses/dated-untrashed-done-task-response';
import type { TrashedTaskResponse } from '$lib/types/responses/trashed-task-response';
import {
	expectRequestBody,
	expectRequestContentType,
	expectRequestCredentialsMode,
	expectRequestMethod,
	expectRequestUrl,
	getResponse
} from '$lib/utils/testing';
import { getLocalTimeZone, today } from '@internationalized/date';
import { generateNKeysBetween } from 'fractional-indexing';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Page from './+page.svelte';

describe('archive page tests', () => {
	beforeEach(() => {
		vi.unstubAllGlobals();
		vi.resetAllMocks();
		vi.stubGlobal('fetch', fetchMock);
		categoryStore.reset();
		undatedTaskStore.reset();
		datedUntrashedDoneTaskStore.reset();
		datedUntrashedOpenTaskStore.reset();
		trashedTaskStore.reset();
		categoryStore.setCategories(categories);
	});

	describe('has no more tasks scenario', () => {
		beforeEach(() => {
			datedUntrashedDoneTaskStore.addOrUpdateTasks([
				...testData.pastCompletedPartition.tasks,
				...testData.futureCompletedPartition.tasks
			]);
			trashedTaskStore.addOrUpdateTasks(testData.trashedPartition.tasks);
		});

		describe('task batch operations', () => {
			it('successfully trashes all untrashed tasks', async () => {
				const tasks = [
					testData.pastCompletedPartition.firstTask,
					testData.pastCompletedPartition.secondTask,
					testData.futureCompletedPartition.thirdTask
				] as const;
				const taskIds = tasks.map((task) => task.id);
				const ranks = generateNKeysBetween(undefined, testData.trashedPartition.undatedOpenTask.rank, tasks.length);
				const response = getResponse(tasks.map((task, index) => ({ id: task.id, rank: ranks[index] })));
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const task of tasks) await clickOnTask(page, task.title);
				const trashAllButton = page.getByRole('button', { name: 'Trash All' });
				await userEvent.click(trashAllButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestUrl(fetchMock, '/api/tasks/batch-trash');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, JSON.stringify({ taskIds }));
				expectRequestContentType(fetchMock, 'application/json');
				const completedColumn = getTaskColumn(page, 'Completed');
				const trashedTaskList = page.getByLabelText('Trashed tasks');
				for (const task of tasks) {
					expect(within(completedColumn).queryByRole('listitem', { name: task.title })).not.toBeInTheDocument();
					expect(within(trashedTaskList).getByRole('listitem', { name: task.title })).toBeInTheDocument();
				}
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toHaveLength(3);
				expect(trashedTaskStore.trashedTasks).toHaveLength(6);
				expect(page.getByRole('button', { name: 'Select Tasks' })).toBeInTheDocument();
			});

			it('successfully opens all untrashed tasks', async () => {
				const tasks = [
					testData.pastCompletedPartition.firstTask,
					testData.pastCompletedPartition.secondTask,
					testData.futureCompletedPartition.thirdTask
				] as const;
				const taskIds = tasks.map((task) => task.id);
				const ranks = generateNKeysBetween(undefined, undefined, tasks.length);
				const response = getResponse(tasks.map((task, index) => ({ id: task.id, rank: ranks[index] })));
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const task of tasks) await clickOnTask(page, task.title);
				const openAllButton = page.getByRole('button', { name: 'Open All' });
				await userEvent.click(openAllButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestUrl(fetchMock, '/api/tasks/batch-open');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, JSON.stringify({ taskIds }));
				expectRequestContentType(fetchMock, 'application/json');
				for (const task of tasks) expectTaskNotToBePresent(page, task.title);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toHaveLength(3);
				expect(datedUntrashedOpenTaskStore.datedUntrashedOpenTasks).toHaveLength(3);
				expect(page.getByRole('button', { name: 'Select Tasks' })).toBeInTheDocument();
			});

			it('successfully cancels a batch operation', async () => {
				const task = testData.pastCompletedPartition.firstTask;

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				await clickOnTask(page, task.title);
				const cancelSelectionButton = page.getByRole('button', { name: 'Cancel Selection' });
				await userEvent.click(cancelSelectionButton);

				expect(page.getByRole('button', { name: 'Select Tasks' })).toBeInTheDocument();
				expect(page.getByRole('checkbox', { name: `Mark "${task.title}" as completed` })).toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});

			it('disables batch actions when no tasks are selected', async () => {
				const page = render(Page);
				await clickOnSelectTasksButton(page);

				expect(page.getByRole('button', { name: 'Open All' })).toBeDisabled();
				expect(page.getByRole('button', { name: 'Trash All' })).toBeDisabled();
				expect(page.getByText('0 / 50')).toBeInTheDocument();
			});

			it('disables batch open when more than 50 tasks are selected', async () => {
				const taskTitles = addTasksExceedingMaximumBatchSize();

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const title of taskTitles) await clickOnTask(page, title);

				expect(page.getByRole('button', { name: 'Open All' })).toBeDisabled();
				expect(page.getByText('51 / 50')).toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});

			it('disables batch trash when more than 50 tasks are selected', async () => {
				const taskTitles = addTasksExceedingMaximumBatchSize();

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const title of taskTitles) await clickOnTask(page, title);

				expect(page.getByRole('button', { name: 'Trash All' })).toBeDisabled();
				expect(page.getByText('51 / 50')).toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});
		});

		describe('task operations', () => {
			it.todo('edits untrashed task');

			it.todo('edits trashed task');

			it.todo('trashes task');

			it.todo('restores task');
		});

		describe('empty trash', () => {
			it.todo('successfully empties trash without applied filters');

			it.todo('successfully empties trash with applied filters');

			it.todo('hides empty trash if there are no trashed tasks');
		});

		it.todo('shows trash info popover on info icon hover');

		describe('filtering', () => {
			it.todo('applies category filter');

			it.todo('applies start date filter');

			it.todo('applies end date filter');

			it.todo('applies date range filter');
		});

		describe('infinite scroll', () => {
			it.todo('does not enable infinite scrolling for exhausted partitions');
		});

		describe('drag and drop', () => {
			describe('within untrashed task partition', () => {
				it.todo('moves a task to the first index');

				it.todo('moves a task to the middle index');

				it.todo('moves a task to the last index');
			});

			describe('within trashed task partition', () => {
				it.todo('moves a task to the first index');

				it.todo('moves a task to the middle index');

				it.todo('moves a task to the last index');
			});

			describe('from trashed task partition', () => {
				describe('to other trashed task partition', () => {
					it.todo('moves a task to the first index');

					it.todo('moves a task to the middle index');

					it.todo('moves a task to the last index');
				});

				describe('to untrashed task partition', () => {
					it.todo('moves a task to the first index');

					it.todo('moves a task to the middle index');

					it.todo('moves a task to the last index');
				});
			});

			describe('from untrashed task partition', () => {
				describe('to untrashed task partition', () => {
					it.todo('moves a task to the first index');

					it.todo('moves a task to the middle index');

					it.todo('moves a task to the last index');
				});
			});
		});
	});

	describe('has more tasks scenario', () => {
		beforeEach(() => {
			datedUntrashedDoneTaskStore.addPage(
				[...testData.pastCompletedPartition.tasks, ...testData.futureCompletedPartition.tasks],
				true
			);
			trashedTaskStore.addPage(testData.trashedPartition.tasks, true);
		});

		describe('infinite scroll', () => {
			it.todo('loads more untrashed tasks without applied filters');

			it.todo('loads more untrashed tasks with applied filters');

			it.todo('loads more trashed tasks without applied filters');

			it.todo('loads more trashed tasks with applied filters');

			it.todo('disables infinite scrolling while a task is being dragged');
		});
	});
});

const fetchMock = vi.fn();
const categories = [
	{ id: 7, title: 'Personal', color: '#22c55e' },
	{ id: 8, title: 'Work', color: '#3b82f6' }
];
const todayDate = today(getLocalTimeZone());
const testData = createTestData();

type TaskPartition<Task> = {
	firstTask: Task;
	secondTask: Task;
	thirdTask: Task;
	tasks: Task[];
};

function createTestData() {
	const pastDate = todayDate.subtract({ days: 1 });
	const futureDate = todayDate.add({ days: 1 });
	let taskId = 1;

	const pastCompletedTaskDetails = [
		{ title: 'Send the signed contract', description: 'Attach the final PDF and copy the project lead.' },
		{ title: 'Book the conference hotel' },
		{ title: 'Submit the travel expenses', description: 'Include the train and taxi receipts.' }
	] as const;
	const pastCompletedTaskRanks = generateNKeysBetween(undefined, undefined, pastCompletedTaskDetails.length);
	const pastCompletedTasks: DatedUntrashedDoneTaskResponse[] = pastCompletedTaskDetails.map((details, index) => ({
		...details,
		id: taskId++,
		categoryId: categories[index % categories.length]!.id,
		date: pastDate,
		rank: pastCompletedTaskRanks[index],
		done: true
	}));

	const futureCompletedTaskDetails = [
		{ title: 'Prepare the workshop agenda' },
		{ title: 'Review the quarterly report', description: 'Check the revenue figures and highlighted risks.' },
		{ title: 'Confirm the catering order' }
	] as const;
	const futureCompletedTaskRanks = generateNKeysBetween(undefined, undefined, futureCompletedTaskDetails.length);
	const futureCompletedTasks: DatedUntrashedDoneTaskResponse[] = futureCompletedTaskDetails.map((details, index) => ({
		...details,
		id: taskId++,
		categoryId: categories[(index + 1) % categories.length]!.id,
		date: futureDate,
		rank: futureCompletedTaskRanks[index],
		done: true
	}));

	const trashedTaskRanks = generateNKeysBetween(undefined, undefined, 3);
	const undatedOpenTrashedTask: TrashedTaskResponse = {
		id: taskId++,
		categoryId: categories[0]!.id,
		title: 'Compare mobile phone plans',
		description: 'Check monthly costs and available data packages.',
		done: false,
		rank: trashedTaskRanks[0]
	};
	const datedOpenTrashedTask: TrashedTaskResponse = {
		id: taskId++,
		categoryId: categories[1]!.id,
		title: 'Schedule the annual checkup',
		date: futureDate,
		done: false,
		rank: trashedTaskRanks[1]
	};
	const datedDoneTrashedTask: TrashedTaskResponse = {
		id: taskId++,
		categoryId: categories[0]!.id,
		title: 'Renew the library books',
		description: 'Extend every loan that expires this week.',
		date: pastDate,
		done: true,
		rank: trashedTaskRanks[2]
	};

	const pastCompletedPartition: TaskPartition<DatedUntrashedDoneTaskResponse> = {
		firstTask: pastCompletedTasks[0]!,
		secondTask: pastCompletedTasks[1]!,
		thirdTask: pastCompletedTasks[2]!,
		tasks: pastCompletedTasks
	};
	const futureCompletedPartition: TaskPartition<DatedUntrashedDoneTaskResponse> = {
		firstTask: futureCompletedTasks[0]!,
		secondTask: futureCompletedTasks[1]!,
		thirdTask: futureCompletedTasks[2]!,
		tasks: futureCompletedTasks
	};
	const trashedTasks = [undatedOpenTrashedTask, datedOpenTrashedTask, datedDoneTrashedTask];

	return {
		pastCompletedPartition,
		pastCompletedPartitionDate: pastDate,
		futureCompletedPartition,
		futureCompletedPartitionDate: futureDate,
		trashedPartition: {
			undatedOpenTask: undatedOpenTrashedTask,
			datedOpenTask: datedOpenTrashedTask,
			datedDoneTask: datedDoneTrashedTask,
			tasks: trashedTasks
		}
	};
}

function addTasksExceedingMaximumBatchSize(): string[] {
	const taskCount = 51;
	const firstTaskId = 10;
	const ranks = generateNKeysBetween(testData.pastCompletedPartition.thirdTask.rank, undefined, taskCount);
	const tasks: DatedUntrashedDoneTaskResponse[] = ranks.map((rank, index) => ({
		id: firstTaskId + index,
		categoryId: categories[index % categories.length]!.id,
		title: `Review completed project note ${index + 1}`,
		description: `Review and organize completed project note ${index + 1}.`,
		date: testData.pastCompletedPartitionDate,
		rank,
		done: true
	}));
	datedUntrashedDoneTaskStore.addOrUpdateTasks(tasks);
	return tasks.map((task) => task.title);
}

async function clickOnSelectTasksButton(page: RenderResult<typeof Page>): Promise<void> {
	const selectTasksButton = page.getByRole('button', { name: 'Select Tasks' });
	await userEvent.click(selectTasksButton);
}

async function clickOnTask(page: RenderResult<typeof Page>, title: string): Promise<void> {
	const task = page.getByRole('listitem', { name: title });
	await fireEvent.pointerUp(task);
}

function expectTaskNotToBePresent(page: RenderResult<typeof Page>, title: string): void {
	const task = page.queryByRole('listitem', { name: title });
	expect(task).not.toBeInTheDocument();
}

function getTaskColumn(page: RenderResult<typeof Page>, heading: string): HTMLElement {
	const taskColumn = page.getByRole('heading', { name: heading }).parentElement;
	if (!(taskColumn instanceof HTMLElement)) throw new Error(`Expected a task column headed "${heading}".`);
	return taskColumn;
}
