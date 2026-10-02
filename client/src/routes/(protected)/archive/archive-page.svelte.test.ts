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
	expectInputToBeInvalid,
	expectRequestBody,
	expectRequestContentType,
	expectRequestCredentialsMode,
	expectRequestMethod,
	expectRequestUrl,
	expectTextareaToBeInvalid,
	getInputById,
	getResponse,
	getTextareaById
} from '$lib/utils/testing';
import { getLocalTimeZone, today } from '@internationalized/date';
import { generateKeyBetween, generateNKeysBetween } from 'fractional-indexing';
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
					const category = categories.find((category) => category.id === task.categoryId)!;
					expectTaskNotToBePresent(completedColumn, task.title);
					expectTaskToBePresent(trashedTaskList, task.title, task.description, category.title);
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
				const completedColumn = getTaskColumn(page, 'Completed');
				for (const task of tasks) expectTaskNotToBePresent(completedColumn, task.title);
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
			describe('task editing', () => {
				it('reacts correctly to missing task title', async () => {
					const task = testData.pastCompletedPartition.firstTask;

					const page = render(Page);
					await openEditTaskDialog(page, task.title);
					const titleInput = await enterTaskTitle(page, '');
					await fireEvent.blur(titleInput);

					const saveButton = page.getByRole('button', { name: 'Save' });
					expect(saveButton).toBeDisabled();
					await expectInputToBeInvalid(page, 'task-title', 'Title must have at least 1 character');
				});

				it('reacts correctly to too long task title', async () => {
					const task = testData.pastCompletedPartition.firstTask;

					const page = render(Page);
					await openEditTaskDialog(page, task.title);
					const tooLongTitle = 'Plan the weekend hiking route '.repeat(4);
					const titleInput = await enterTaskTitle(page, tooLongTitle);
					await fireEvent.blur(titleInput);

					const saveButton = page.getByRole('button', { name: 'Save' });
					expect(saveButton).toBeDisabled();
					await expectInputToBeInvalid(page, 'task-title', 'Title must have at most 100 characters');
				});

				it('reacts correctly to too long task description', async () => {
					const task = testData.pastCompletedPartition.firstTask;

					const page = render(Page);
					await openEditTaskDialog(page, task.title);
					const tooLongDescription = 'Compare the available trails, travel times, and weather forecast. '.repeat(9);
					const descriptionInput = await enterTaskDescription(page, tooLongDescription);
					await fireEvent.blur(descriptionInput);

					const saveButton = page.getByRole('button', { name: 'Save' });
					expect(saveButton).toBeDisabled();
					await expectTextareaToBeInvalid(page, 'task-description', 'Description must have at most 500 characters');
				});

				it('cancels and closes the edit-task dialog', async () => {
					const task = testData.pastCompletedPartition.firstTask;

					const page = render(Page);
					await openEditTaskDialog(page, task.title);
					const cancelButton = page.getByRole('button', { name: 'Cancel' });
					await userEvent.click(cancelButton);

					const editTaskDialog = page.queryByRole('dialog', { name: 'Edit Task' });
					expect(editTaskDialog).not.toBeInTheDocument();
				});

				it('edits untrashed task', async () => {
					const task = testData.pastCompletedPartition.firstTask;
					const category = categories.find((category) => category.id === task.categoryId)!;
					const newTitle = 'Archive the signed contract';
					const newDescription = 'Store the signed copy with the completed project documents.';

					const response = getResponse({
						id: task.id,
						categoryId: task.categoryId,
						title: newTitle,
						description: newDescription,
						date: task.date.toString(),
						rank: task.rank,
						done: true,
						trashed: false
					});
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					await openEditTaskDialog(page, task.title);
					await enterTaskTitle(page, newTitle);
					await enterTaskDescription(page, newDescription);
					const saveButton = page.getByRole('button', { name: 'Save' });
					await userEvent.click(saveButton);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const editTaskDialog = page.queryByRole('dialog', { name: 'Edit Task' });
					expect(editTaskDialog).not.toBeInTheDocument();
					const requestBody = JSON.stringify({
						title: newTitle,
						description: newDescription,
						categoryId: task.categoryId,
						date: task.date.toString(),
						id: task.id
					});
					expectRequestUrl(fetchMock, '/api/tasks');
					expectRequestMethod(fetchMock, 'PUT');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					const completedTaskList = page.getByLabelText(`Completed tasks dated ${testData.pastCompletedPartitionLabel}`);
					expectTaskToBePresent(completedTaskList, newTitle, newDescription, category.title);
				});

				it('edits trashed task', async () => {
					const task = testData.trashedPartition.undatedOpenTask;
					const category = categories.find((category) => category.id === task.categoryId)!;
					const newTitle = 'Compare family mobile phone plans';
					const newDescription = 'Compare monthly costs, roaming, and shared data packages.';

					const response = getResponse({
						id: task.id,
						categoryId: task.categoryId,
						title: newTitle,
						description: newDescription,
						rank: task.rank,
						done: false,
						trashed: true
					});
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					await openEditTaskDialog(page, task.title);
					await enterTaskTitle(page, newTitle);
					await enterTaskDescription(page, newDescription);
					const saveButton = page.getByRole('button', { name: 'Save' });
					await userEvent.click(saveButton);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const editTaskDialog = page.queryByRole('dialog', { name: 'Edit Task' });
					expect(editTaskDialog).not.toBeInTheDocument();
					const requestBody = JSON.stringify({
						title: newTitle,
						description: newDescription,
						categoryId: task.categoryId,
						id: task.id
					});
					expectRequestUrl(fetchMock, '/api/tasks');
					expectRequestMethod(fetchMock, 'PUT');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					const trashedTaskList = page.getByLabelText('Trashed tasks');
					expectTaskToBePresent(trashedTaskList, newTitle, newDescription, category.title);
				});
			});

			it('trashes task', async () => {
				const task = testData.futureCompletedPartition.firstTask;
				const category = categories.find((category) => category.id === task.categoryId)!;
				const trashedTaskRank = generateKeyBetween(undefined, testData.trashedPartition.undatedOpenTask.rank);

				const response = getResponse({ rank: trashedTaskRank });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const moreOptionsButton = page.getByRole('button', { name: `More options for "${task.title}"` });
				await fireEvent.click(moreOptionsButton);
				const trashButton = await page.findByRole('button', { name: 'Trash' });
				await userEvent.click(trashButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const requestUrl = `/api/tasks/${task.id}/trash`;
				expectRequestUrl(fetchMock, requestUrl);
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				const completedTaskList = page.getByLabelText(`Completed tasks dated ${testData.futureCompletedPartitionLabel}`);
				const trashedTaskList = page.getByLabelText('Trashed tasks');
				expectTaskNotToBePresent(completedTaskList, task.title);
				expectTaskToBePresent(trashedTaskList, task.title, task.description, category.title);
			});

			it('restores task', async () => {
				const task = testData.trashedPartition.datedDoneTask;
				if (task.date === undefined) throw new Error('Expected the trashed task to have a date.');
				const category = categories.find((category) => category.id === task.categoryId)!;
				const restoredRank = generateNKeysBetween(undefined, undefined, 1)[0];

				fetchMock
					.mockResolvedValueOnce(getResponse({ rank: restoredRank }))
					.mockResolvedValueOnce(getResponse({ trashedTasksExist: true }));

				const page = render(Page);
				const moreOptionsButton = page.getByRole('button', { name: `More options for "${task.title}"` });
				await fireEvent.click(moreOptionsButton);
				const restoreButton = await page.findByRole('button', { name: 'Restore' });
				await userEvent.click(restoreButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
				expectRequestUrl(fetchMock, `/api/tasks/${task.id}/restore`);
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expectRequestUrl(fetchMock, '/api/tasks/trashed/exists', 1);
				expectRequestMethod(fetchMock, 'GET', 1);
				expectRequestCredentialsMode(fetchMock, 'same-origin', 1);
				expectRequestBody(fetchMock, undefined, 1);
				const completedTaskList = page.getByLabelText(`Completed tasks dated ${testData.pastCompletedPartitionLabel}`);
				const trashedTaskList = page.getByLabelText('Trashed tasks');
				expectTaskToBePresent(completedTaskList, task.title, task.description, category.title);
				expectTaskNotToBePresent(trashedTaskList, task.title);
				expect(trashedTaskStore.trashedTasksExist).toBe(true);
			});
		});

		describe('empty trash', () => {
			it('successfully empties trash without applied filters', async () => {
				fetchMock.mockResolvedValue(new Response(null, { status: 204 }));

				const page = render(Page);
				const emptyTrashButton = page.getByRole('button', { name: 'Empty Trash' });
				await userEvent.click(emptyTrashButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestUrl(fetchMock, '/api/tasks/trashed');
				expectRequestMethod(fetchMock, 'DELETE');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				for (const task of testData.trashedPartition.tasks) expectTaskNotToBePresent(page.container, task.title);
				expect(trashedTaskStore.trashedTasks).toHaveLength(0);
				expect(trashedTaskStore.trashedTasksExist).toBe(false);
				expect(page.queryByRole('button', { name: 'Empty Trash' })).not.toBeInTheDocument();
			});

			it('successfully empties trash with applied filters', async () => {
				fetchMock
					.mockResolvedValueOnce(getResponse({ tasks: [], hasMore: false }))
					.mockResolvedValueOnce(getResponse({ tasks: [], hasMore: false }))
					.mockResolvedValueOnce(new Response(null, { status: 204 }));

				const page = render(Page);
				const excludePersonalButtons = page.getAllByRole('button', { name: 'Exclude Personal' });
				await userEvent.click(excludePersonalButtons[0]!);
				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
				const emptyTrashButton = page.getByRole('button', { name: 'Empty Trash' });
				await userEvent.click(emptyTrashButton);

				const emptyTrashDialog = page.getByRole('dialog', { name: 'Empty Trash?' });
				expect(emptyTrashDialog).toHaveTextContent(
					'Emptying the trash will permanently delete all trashed tasks, including tasks hidden by these filters.'
				);
				const deleteAllButton = within(emptyTrashDialog).getByRole('button', { name: 'Delete All' });
				await userEvent.click(deleteAllButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(3));
				expectRequestUrl(fetchMock, '/api/tasks/trashed', 2);
				expectRequestMethod(fetchMock, 'DELETE', 2);
				expectRequestCredentialsMode(fetchMock, 'same-origin', 2);
				expectRequestBody(fetchMock, undefined, 2);
				expect(page.queryByRole('dialog', { name: 'Empty Trash?' })).not.toBeInTheDocument();
				expect(trashedTaskStore.trashedTasks).toHaveLength(0);
				expect(trashedTaskStore.trashedTasksExist).toBe(false);
				expect(page.queryByRole('button', { name: 'Empty Trash' })).not.toBeInTheDocument();
			});

			it('hides empty trash if there are no trashed tasks', () => {
				trashedTaskStore.reset();

				const page = render(Page);

				expect(page.queryByRole('button', { name: 'Empty Trash' })).not.toBeInTheDocument();
			});
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
	const pastDateLabel = `${String(pastDate.day).padStart(2, '0')}.${String(pastDate.month).padStart(2, '0')}.${pastDate.year}`;
	const futureDateLabel = `${String(futureDate.day).padStart(2, '0')}.${String(futureDate.month).padStart(2, '0')}.${futureDate.year}`;
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
		pastCompletedPartitionLabel: pastDateLabel,
		futureCompletedPartition,
		futureCompletedPartitionDate: futureDate,
		futureCompletedPartitionLabel: futureDateLabel,
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

function expectTaskNotToBePresent(taskList: HTMLElement, title: string): void {
	const task = within(taskList).queryByRole('listitem', { name: title });
	expect(task).not.toBeInTheDocument();
}

function getTaskColumn(page: RenderResult<typeof Page>, heading: string): HTMLElement {
	const taskColumn = page.getByRole('heading', { name: heading }).parentElement;
	if (!(taskColumn instanceof HTMLElement)) throw new Error(`Expected a task column headed "${heading}".`);
	return taskColumn;
}

async function openEditTaskDialog(page: RenderResult<typeof Page>, title: string): Promise<void> {
	const moreOptionsButton = page.getByRole('button', { name: `More options for "${title}"` });
	await fireEvent.click(moreOptionsButton);
	const editButton = await page.findByRole('button', { name: 'Edit' });
	await userEvent.click(editButton);
}

async function enterTaskTitle(page: RenderResult<typeof Page>, title: string): Promise<HTMLInputElement> {
	const titleInput = getInputById(page, 'task-title');
	await fireEvent.input(titleInput, { target: { value: title } });
	return titleInput;
}

async function enterTaskDescription(
	page: RenderResult<typeof Page>,
	description: string | undefined
): Promise<HTMLTextAreaElement> {
	const descriptionInput = getTextareaById(page, 'task-description');
	await fireEvent.input(descriptionInput, { target: { value: description ?? '' } });
	return descriptionInput;
}

function expectTaskToBePresent(
	taskList: HTMLElement,
	title: string,
	description: string | undefined,
	categoryName: string
): void {
	const taskListContent = within(taskList);
	const task = taskListContent.getByRole('listitem', { name: title });
	const taskContent = within(task);
	expect(task).toBeInTheDocument();
	if (description !== undefined) expect(taskContent.getByText(description)).toBeInTheDocument();
	expect(taskContent.getByText(categoryName)).toBeInTheDocument();
}
