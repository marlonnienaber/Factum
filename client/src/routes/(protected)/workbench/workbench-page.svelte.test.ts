import { fireEvent, render, type RenderResult, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { getLocalTimeZone, today } from '@internationalized/date';
import { generateNKeysBetween } from 'fractional-indexing';
import { categoryStore } from '$lib/stores/category-store.svelte';
import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
import { todayProgressStore } from '$lib/stores/today-progress-store.svelte';
import { trashedTaskStore } from '$lib/stores/trashed-task-store.svelte';
import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
import type { DatedUntrashedDoneTaskResponse } from '$lib/types/responses/dated-untrashed-done-task-response';
import type { DatedUntrashedOpenTaskResponse } from '$lib/types/responses/dated-untrashed-open-task-response';
import {
	expectRequestBody,
	expectRequestContentType,
	expectRequestCredentialsMode,
	expectRequestMethod,
	expectRequestUrl,
	getResponse
} from '$lib/utils/testing';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Page from './+page.svelte';

describe('workbench page tests', () => {
	beforeEach(() => {
		vi.unstubAllGlobals();
		vi.resetAllMocks();
		vi.stubGlobal('fetch', fetchMock);
		categoryStore.reset();
		undatedTaskStore.reset();
		datedUntrashedDoneTaskStore.reset();
		datedUntrashedOpenTaskStore.reset();
		trashedTaskStore.reset();
		todayProgressStore.reset();
		categoryStore.setCategories(categories);
		todayProgressStore.setProgress({
			openTaskCount: testData.openPartition.tasks.length,
			doneTaskCount: testData.donePartition.tasks.length
		});
	});

	describe('has no more tasks scenario', () => {
		beforeEach(() => {
			datedUntrashedOpenTaskStore.addTodayPage(testData.openPartition.tasks, false);
			datedUntrashedDoneTaskStore.addPage(testData.donePartition.tasks, false);
		});

		describe('task batch operations', () => {
			it('successfully trashes tasks', async () => {
				const tasks = [
					testData.openPartition.firstTask,
					testData.openPartition.secondTask,
					testData.donePartition.thirdTask
				] as const;
				const taskIds = tasks.map((task) => task.id);
				const trashedTaskRanks = generateNKeysBetween(undefined, undefined, tasks.length);
				const trashedTasks = tasks.map((task, index) => ({ id: task.id, rank: trashedTaskRanks[index] }));
				const response = getResponse(trashedTasks);
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const task of tasks) await clickOnTask(page, task.title);
				const trashAllButton = page.getByRole('button', { name: 'Trash' });
				await userEvent.click(trashAllButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const requestBody = JSON.stringify({ taskIds });
				expectRequestUrl(fetchMock, '/api/tasks/batch-trash');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				for (const task of tasks) expectTaskNotToBePresent(page, task.title);
				expect(datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks).toHaveLength(1);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toHaveLength(2);
				expect(trashedTaskStore.trashedTasks).toHaveLength(3);
				const selectTasksButton = page.getByRole('button', { name: 'Select Tasks' });
				expect(selectTasksButton).toBeInTheDocument();
			});

			it('successfully completes all open tasks', async () => {
				const tasks = testData.openPartition.tasks;
				const taskIds = tasks.map((task) => task.id);
				const completedTaskRanks = generateNKeysBetween(undefined, testData.donePartition.firstTask.rank, tasks.length);
				const completedTasks = tasks.map((task, index) => ({
					id: task.id,
					date: todayDate.toString(),
					rank: completedTaskRanks[index]
				}));
				const response = getResponse(completedTasks);
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const task of tasks) await clickOnTask(page, task.title);
				const completeAllButton = page.getByRole('button', { name: 'Complete' });
				await userEvent.click(completeAllButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const requestBody = JSON.stringify({ taskIds });
				expectRequestUrl(fetchMock, '/api/tasks/batch-complete');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expect(datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks).toHaveLength(0);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toHaveLength(6);
				const selectTasksButton = page.getByRole('button', { name: 'Select Tasks' });
				expect(selectTasksButton).toBeInTheDocument();
			});

			it('successfully opens all done tasks', async () => {
				const tasks = testData.donePartition.tasks;
				const taskIds = tasks.map((task) => task.id);
				const openedTaskRanks = generateNKeysBetween(undefined, testData.openPartition.firstTask.rank, tasks.length);
				const openedTasks = tasks.map((task, index) => ({ id: task.id, rank: openedTaskRanks[index] }));
				const response = getResponse(openedTasks);
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const task of tasks) await clickOnTask(page, task.title);
				const openAllButton = page.getByRole('button', { name: 'Open' });
				await userEvent.click(openAllButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const requestBody = JSON.stringify({ taskIds });
				expectRequestUrl(fetchMock, '/api/tasks/batch-open');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expect(datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks).toHaveLength(6);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toHaveLength(0);
				const selectTasksButton = page.getByRole('button', { name: 'Select Tasks' });
				expect(selectTasksButton).toBeInTheDocument();
			});

			it('disables batch opening if an open task is selected', async () => {
				const task = testData.openPartition.firstTask;

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				await clickOnTask(page, task.title);

				const openAllButton = page.getByRole('button', { name: 'Open' });
				const completeAllButton = page.getByRole('button', { name: 'Complete' });
				const trashAllButton = page.getByRole('button', { name: 'Trash' });
				expect(openAllButton).toBeDisabled();
				expect(completeAllButton).toBeEnabled();
				expect(trashAllButton).toBeEnabled();
				expect(fetchMock).not.toHaveBeenCalled();
			});

			it('disables batch completion if a done task is selected', async () => {
				const task = testData.donePartition.firstTask;

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				await clickOnTask(page, task.title);

				const completeAllButton = page.getByRole('button', { name: 'Complete' });
				const openAllButton = page.getByRole('button', { name: 'Open' });
				const trashAllButton = page.getByRole('button', { name: 'Trash' });
				expect(completeAllButton).toBeDisabled();
				expect(openAllButton).toBeEnabled();
				expect(trashAllButton).toBeEnabled();
				expect(fetchMock).not.toHaveBeenCalled();
			});

			it('disables batch operations when more than 50 tasks are selected', async () => {
				const taskTitles = addTasksExceedingMaximumBatchSize();

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const title of taskTitles) await clickOnTask(page, title);

				const completeAllButton = page.getByRole('button', { name: 'Complete' });
				const openAllButton = page.getByRole('button', { name: 'Open' });
				const trashAllButton = page.getByRole('button', { name: 'Trash' });
				expect(completeAllButton).toBeDisabled();
				expect(openAllButton).toBeDisabled();
				expect(trashAllButton).toBeDisabled();
				expect(page.getByText('51 / 50')).toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});

			it('disables batch operations when no task is selected', async () => {
				const page = render(Page);
				await clickOnSelectTasksButton(page);

				const completeAllButton = page.getByRole('button', { name: 'Complete' });
				const openAllButton = page.getByRole('button', { name: 'Open' });
				const trashAllButton = page.getByRole('button', { name: 'Trash' });
				expect(completeAllButton).toBeDisabled();
				expect(openAllButton).toBeDisabled();
				expect(trashAllButton).toBeDisabled();
				expect(page.getByText('0 / 50')).toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});

			it('successfully cancels a batch operation', async () => {
				const task = testData.openPartition.firstTask;

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				await clickOnTask(page, task.title);
				const cancelSelectionButton = page.getByRole('button', { name: 'Cancel' });
				await userEvent.click(cancelSelectionButton);

				const selectTasksButton = page.getByRole('button', { name: 'Select Tasks' });
				expect(selectTasksButton).toBeInTheDocument();
				const completeTaskButton = page.getByRole('checkbox', {
					name: `Mark "${task.title}" as completed`
				});
				expect(completeTaskButton).toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});
		});

		describe('task operations', () => {
			describe('task editing', () => {
				it.todo('reacts correctly to missing task title');

				it.todo('reacts correctly to too long task title');

				it.todo('reacts correctly to too long task description');

				it.todo('cancels and closes the edit-task dialog');

				it.todo('edits open task');

				it.todo('edits done task');
			});

			describe('task creation', () => {
				it.todo('reacts correctly to missing task title');

				it.todo('reacts correctly to too long task title');

				it.todo('reacts correctly to too long task description');

				it.todo('cancels and closes the edit-task dialog');

				it.todo('adds a task'); // check especially that today date is set as default
			});

			describe('task trashing', () => {
				it.todo('trashes open task');

				it.todo('trashed done task');
			});
		});

		describe('filtering', () => {
			it.todo('applies category filter');
		});

		describe('infinite scroll', () => {
			it.todo('does not enable infinite scrolling for exhausted partitions');
		});

		describe('drag and drop', () => {
			describe('within open partition', () => {
				it.todo('moves a task to the first index');

				it.todo('moves a task to the middle index');

				it.todo('moves a task to the last index');
			});

			describe('within done partition', () => {
				it.todo('moves a task to the first index');

				it.todo('moves a task to the middle index');

				it.todo('moves a task to the last index');
			});

			describe('from open to done partition', () => {
				it.todo('moves a task to the first index');

				it.todo('moves a task to the middle index');

				it.todo('moves a task to the last index');
			});

			describe('from done to open partition', () => {
				it.todo('moves a task to the first index');

				it.todo('moves a task to the middle index');

				it.todo('moves a task to the last index');
			});
		});

		describe('progress', () => {
			it.todo('displays progress information correctly');

			it.todo('reacts correctly to completion by checkbox');

			it.todo('reacts correctly to completion by drag and drop');

			it.todo('reacts correctly to opening by checkbox');

			it.todo('reacts correctly to opening by drag and drop');
		});
	});

	describe('has more tasks scenario', () => {
		beforeEach(() => {
			datedUntrashedOpenTaskStore.addTodayPage(testData.openPartition.tasks, true);
			datedUntrashedDoneTaskStore.addPage(testData.donePartition.tasks, true);
		});

		describe('infinite scroll', () => {
			it.todo('loads more open tasks without applied filters');

			it.todo('loads more open tasks with applied filters');

			it.todo('loads more done tasks without applied filters');

			it.todo('loads more done tasks with applied filters');

			it.todo('disables infinite scrolling while a task is being dragged');
		});
	});
});

const fetchMock = vi.fn();
const personalCategory = { id: 7, title: 'Personal', color: '#22c55e' };
const workCategory = { id: 8, title: 'Work', color: '#3b82f6' };
const categories = [personalCategory, workCategory];
const todayDate = today(getLocalTimeZone());
const testData = createTestData();

type TaskPartition<Task> = {
	firstTask: Task;
	secondTask: Task;
	thirdTask: Task;
	tasks: Task[];
};

function createTestData() {
	let taskId = 1;

	const openTaskDetails = [
		{ categoryId: personalCategory.id, title: 'Buy groceries' },
		{
			categoryId: workCategory.id,
			title: 'Submit the expense report',
			description: 'Attach the train and hotel receipts before submitting.'
		},
		{
			categoryId: workCategory.id,
			title: 'Water the balcony plants',
			description: 'Check the herbs and tomatoes before watering.'
		}
	] as const;
	const openTaskRanks = generateNKeysBetween(undefined, undefined, openTaskDetails.length);
	const openTasks: DatedUntrashedOpenTaskResponse[] = openTaskDetails.map((details, index) => ({
		...details,
		id: taskId++,
		date: todayDate,
		rank: openTaskRanks[index],
		done: false
	}));

	const doneTaskDetails = [
		{
			categoryId: personalCategory.id,
			title: 'Review the project proposal',
			description: 'Leave comments on the final draft.'
		},
		{ categoryId: personalCategory.id, title: 'Book the dentist appointment' },
		{
			categoryId: workCategory.id,
			title: 'Send the weekly status update',
			description: 'Summarize completed work and current blockers.'
		}
	] as const;
	const doneTaskRanks = generateNKeysBetween(undefined, undefined, doneTaskDetails.length);
	const doneTasks: DatedUntrashedDoneTaskResponse[] = doneTaskDetails.map((details, index) => ({
		...details,
		id: taskId++,
		date: todayDate,
		rank: doneTaskRanks[index],
		done: true
	}));

	const openPartition: TaskPartition<DatedUntrashedOpenTaskResponse> = {
		firstTask: openTasks[0]!,
		secondTask: openTasks[1]!,
		thirdTask: openTasks[2]!,
		tasks: openTasks
	};
	const donePartition: TaskPartition<DatedUntrashedDoneTaskResponse> = {
		firstTask: doneTasks[0]!,
		secondTask: doneTasks[1]!,
		thirdTask: doneTasks[2]!,
		tasks: doneTasks
	};

	return { openPartition, donePartition };
}

function addTasksExceedingMaximumBatchSize(): string[] {
	const taskCount = 51;
	const firstTaskId = 10;
	const ranks = generateNKeysBetween(testData.openPartition.thirdTask.rank, undefined, taskCount);
	const tasks: DatedUntrashedOpenTaskResponse[] = ranks.map((rank, index) => ({
		id: firstTaskId + index,
		categoryId: personalCategory.id,
		title: `Review open project note ${index + 1}`,
		description: `Review and organize open project note ${index + 1}.`,
		date: todayDate,
		rank,
		done: false
	}));
	datedUntrashedOpenTaskStore.addOrUpdateTasks(tasks);
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
