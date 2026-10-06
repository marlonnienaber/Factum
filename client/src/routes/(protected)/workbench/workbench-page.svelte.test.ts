import { fireEvent, render, type RenderResult, waitFor, within } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { getLocalTimeZone, today } from '@internationalized/date';
import { generateKeyBetween, generateNKeysBetween } from 'fractional-indexing';
import { type DndEvent, SOURCES, TRIGGERS } from 'svelte-dnd-action';
import { categoryStore } from '$lib/stores/category-store.svelte';
import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
import { todayProgressStore } from '$lib/stores/today-progress-store.svelte';
import { trashedTaskStore } from '$lib/stores/trashed-task-store.svelte';
import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
import type { DatedUntrashedDoneTaskResponse } from '$lib/types/responses/dated-untrashed-done-task-response';
import type { DatedUntrashedOpenTaskResponse } from '$lib/types/responses/dated-untrashed-open-task-response';
import {
	expectInputToBeInvalid,
	expectRequestBody,
	expectRequestContentType,
	expectRequestCredentialsMode,
	expectRequestMethod,
	expectRequestPath,
	expectRequestUrl,
	expectSelectToBeInvalid,
	expectTextareaToBeInvalid,
	getInputById,
	getRequestSearchParameter,
	getResponse,
	getSelectById,
	getTextareaById,
	intersectionObserverMock,
	mockPendingFetch
} from '$lib/utils/testing';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
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

	afterEach(async () => {
		if (document.querySelector('[role="dialog"]')) await userEvent.keyboard('{Escape}');
		document.body.removeAttribute('style');
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
				it('reacts correctly to missing task title', async () => {
					const task = testData.openPartition.firstTask;

					const page = render(Page);
					await openEditTaskDialog(page, task.title);
					const titleInput = await enterTaskTitle(page, '');
					await fireEvent.blur(titleInput);

					const saveButton = page.getByRole('button', { name: 'Save' });
					expect(saveButton).toBeDisabled();
					await expectInputToBeInvalid(page, 'task-title', 'Title must have at least 1 character');
				});

				it('reacts correctly to too long task title', async () => {
					const task = testData.openPartition.firstTask;

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
					const task = testData.openPartition.firstTask;

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
					const task = testData.openPartition.firstTask;

					const page = render(Page);
					await openEditTaskDialog(page, task.title);
					const cancelButton = page.getByRole('button', { name: 'Cancel' });
					await userEvent.click(cancelButton);

					const editTaskDialog = page.queryByRole('dialog', { name: 'Edit Task' });
					expect(editTaskDialog).not.toBeInTheDocument();
				});

				it('edits open task', async () => {
					const task = testData.openPartition.firstTask;
					const category = categories.find((category) => category.id === task.categoryId)!;
					const newTitle = 'Buy groceries for the weekend';
					const newDescription = 'Pick up vegetables, bread, and coffee.';

					const response = getResponse({
						id: task.id,
						categoryId: task.categoryId,
						title: newTitle,
						description: newDescription,
						date: task.date.toString(),
						rank: task.rank,
						done: false,
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
					const serializedDate = task.date.toString();
					const requestBody = JSON.stringify({
						title: newTitle,
						description: newDescription,
						categoryId: task.categoryId,
						date: serializedDate,
						id: task.id
					});
					expectRequestUrl(fetchMock, '/api/tasks');
					expectRequestMethod(fetchMock, 'PUT');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					const openTaskList = page.getByLabelText('Open tasks');
					expectTaskToBePresent(openTaskList, newTitle, newDescription, category.title);
				});

				it('edits done task', async () => {
					const task = testData.donePartition.firstTask;
					const category = categories.find((category) => category.id === task.categoryId)!;
					const newTitle = 'Review the final project proposal';
					const newDescription = 'Resolve the remaining comments on the final draft.';

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
					const serializedDate = task.date.toString();
					const requestBody = JSON.stringify({
						title: newTitle,
						description: newDescription,
						categoryId: task.categoryId,
						date: serializedDate,
						id: task.id
					});
					expectRequestUrl(fetchMock, '/api/tasks');
					expectRequestMethod(fetchMock, 'PUT');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					const doneTaskList = page.getByLabelText('Done tasks');
					expectTaskToBePresent(doneTaskList, newTitle, newDescription, category.title);
				});
			});

			describe('task creation', () => {
				it('reacts correctly to missing task title', async () => {
					const page = render(Page);
					await openCreateTaskDialog(page);
					await selectCategory(page, personalCategory.title);
					const titleInput = getInputById(page, 'task-title');
					await userEvent.click(titleInput);
					await userEvent.tab();

					const createButton = page.getByRole('button', { name: 'Create' });
					expect(createButton).toBeDisabled();
					await expectInputToBeInvalid(page, 'task-title', 'Title must have at least 1 character');
				});

				it('reacts correctly to too long task title', async () => {
					const page = render(Page);
					await openCreateTaskDialog(page);
					await selectCategory(page, personalCategory.title);
					const tooLongTitle = 'Plan the weekend hiking route '.repeat(4);
					const titleInput = await enterTaskTitle(page, tooLongTitle);
					await fireEvent.blur(titleInput);

					const createButton = page.getByRole('button', { name: 'Create' });
					expect(createButton).toBeDisabled();
					await expectInputToBeInvalid(page, 'task-title', 'Title must have at most 100 characters');
				});

				it('reacts correctly to too long task description', async () => {
					const page = render(Page);
					await openCreateTaskDialog(page);
					await selectCategory(page, personalCategory.title);
					await enterTaskTitle(page, 'Plan a team workshop');
					const tooLongDescription = 'Compare the available trails, travel times, and weather forecast. '.repeat(9);
					const descriptionInput = await enterTaskDescription(page, tooLongDescription);
					await fireEvent.blur(descriptionInput);

					const createButton = page.getByRole('button', { name: 'Create' });
					expect(createButton).toBeDisabled();
					await expectTextareaToBeInvalid(page, 'task-description', 'Description must have at most 500 characters');
				});

				it('reacts correctly to missing task category', async () => {
					const page = render(Page);
					await openCreateTaskDialog(page);
					await enterTaskTitle(page, 'Schedule an annual checkup');
					const categorySelect = getSelectById(page, 'task-category');
					await userEvent.click(categorySelect);
					await userEvent.keyboard('{Escape}');
					await userEvent.tab();

					const createButton = page.getByRole('button', { name: 'Create' });
					expect(createButton).toBeDisabled();
					await expectSelectToBeInvalid(page, 'task-category', 'Category is required');
				});

				it('cancels and closes the create-task dialog', async () => {
					const page = render(Page);
					await openCreateTaskDialog(page);
					const cancelButton = page.getByRole('button', { name: 'Cancel' });
					await userEvent.click(cancelButton);

					const createTaskDialog = page.queryByRole('dialog', { name: 'Create Task' });
					expect(createTaskDialog).not.toBeInTheDocument();
				});

				it('adds a task', async () => {
					const title = 'Prepare the quarterly presentation';
					const description = 'Summarize project results and prepare the final slides.';
					const rank = generateKeyBetween(testData.openPartition.thirdTask.rank, undefined);
					const taskId = 7;

					const response = getResponse({ id: taskId, rank }, 201);
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					await openCreateTaskDialog(page);
					await enterTaskTitle(page, title);
					await enterTaskDescription(page, description);
					await selectCategory(page, personalCategory.title);
					const createButton = page.getByRole('button', { name: 'Create' });
					await userEvent.click(createButton);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const createTaskDialog = page.queryByRole('dialog', { name: 'Create Task' });
					expect(createTaskDialog).not.toBeInTheDocument();
					const requestBody = JSON.stringify({
						title,
						description,
						categoryId: personalCategory.id,
						date: todayDate.toString()
					});
					expectRequestUrl(fetchMock, '/api/tasks');
					expectRequestMethod(fetchMock, 'POST');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					const openTaskList = page.getByLabelText('Open tasks');
					expectTaskToBePresent(openTaskList, title, description, personalCategory.title);
				});
			});

			it('postpones an open task to tomorrow', async () => {
				const task = testData.openPartition.firstTask;
				const tomorrowDate = todayDate.add({ days: 1 });
				const postponedTaskRank = generateKeyBetween(undefined, undefined);
				fetchMock.mockResolvedValue(getResponse({ rank: postponedTaskRank }));

				const page = render(Page);
				const moreOptionsButton = page.getByRole('button', { name: `More options for "${task.title}"` });
				await fireEvent.click(moreOptionsButton);
				const tomorrowButton = await page.findByRole('button', { name: 'Tomorrow' });
				await userEvent.click(tomorrowButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
				const requestBody = JSON.stringify({
					date: tomorrowDate.toString(),
					done: false,
					trashed: false
				});
				expectRequestUrl(fetchMock, requestUrl);
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expect(datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks).not.toContainEqual(task);
				expect(datedUntrashedOpenTaskStore.datedUntrashedOpenNonTodayTasks).toContainEqual({
					...task,
					date: tomorrowDate,
					rank: postponedTaskRank
				});
			});

			it('clones a done task', async () => {
				const task = testData.donePartition.firstTask;
				const tomorrowDate = todayDate.add({ days: 1 });
				const clonedTaskId = 7;
				const clonedTaskRank = generateKeyBetween(undefined, undefined);
				fetchMock.mockResolvedValue(getResponse({ id: clonedTaskId, rank: clonedTaskRank }, 201));

				const page = render(Page);
				const moreOptionsButton = page.getByRole('button', { name: `More options for "${task.title}"` });
				await fireEvent.click(moreOptionsButton);
				const cloneButton = await page.findByRole('button', { name: 'Clone' });
				await userEvent.click(cloneButton);
				expect(page.getByRole('dialog', { name: 'Create Task' })).toBeInTheDocument();
				const createButton = page.getByRole('button', { name: 'Create' });
				await userEvent.click(createButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expect(page.queryByRole('dialog', { name: 'Create Task' })).not.toBeInTheDocument();
				const requestBody = JSON.stringify({
					title: task.title,
					description: task.description,
					categoryId: task.categoryId,
					date: tomorrowDate.toString()
				});
				expectRequestUrl(fetchMock, '/api/tasks');
				expectRequestMethod(fetchMock, 'POST');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expect(datedUntrashedOpenTaskStore.datedUntrashedOpenNonTodayTasks).toContainEqual({
					id: clonedTaskId,
					categoryId: task.categoryId,
					title: task.title,
					description: task.description,
					date: tomorrowDate,
					rank: clonedTaskRank,
					done: false
				});
			});

			describe('task trashing', () => {
				it('trashes open task', async () => {
					const task = testData.openPartition.firstTask;
					const trashedTaskRank = generateKeyBetween(undefined, undefined);
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
					expectTaskNotToBePresent(page, task.title);
					expect(trashedTaskStore.trashedTasks).toContainEqual({ ...task, rank: trashedTaskRank });
				});

				it('trashes done task', async () => {
					const task = testData.donePartition.firstTask;
					const trashedTaskRank = generateKeyBetween(undefined, undefined);
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
					expectTaskNotToBePresent(page, task.title);
					expect(trashedTaskStore.trashedTasks).toContainEqual({ ...task, rank: trashedTaskRank });
				});
			});
		});

		describe('filtering', () => {
			it('applies category filter', async () => {
				const openTasks = [testData.openPartition.secondTask, testData.openPartition.thirdTask];
				const doneTasks = [testData.donePartition.thirdTask];
				mockFilterResponses(openTasks, doneTasks);

				const page = render(Page);
				const excludePersonalButtons = page.getAllByRole('button', { name: 'Exclude Personal' });
				await userEvent.click(excludePersonalButtons[0]!);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
				expectRequestUrl(fetchMock, '/api/tasks/dated-untrashed-open-today?categoryId=8');
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expectRequestUrl(fetchMock, '/api/tasks/dated-untrashed-done-today?categoryId=8', 1);
				expectRequestMethod(fetchMock, 'GET', 1);
				expectRequestCredentialsMode(fetchMock, 'same-origin', 1);
				expectRequestBody(fetchMock, undefined, 1);
				expect(datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks).toEqual(openTasks);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toEqual(doneTasks);
			});
		});

		describe('infinite scroll', () => {
			it('does not enable infinite scrolling for exhausted partitions', () => {
				const page = render(Page);
				const openTaskColumn = getTaskColumn(page, 'Open');
				const doneTaskColumn = getTaskColumn(page, 'Done');

				expect(openTaskColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(doneTaskColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});
		});

		describe('drag and drop', () => {
			describe('within open partition', () => {
				it('moves a task to the first index', async () => {
					const task = testData.openPartition.thirdTask;
					const expectedRank = generateKeyBetween(undefined, testData.openPartition.firstTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const openTaskList = page.getByLabelText('Open tasks');
					const reorderedTasks = [task, testData.openPartition.firstTask, testData.openPartition.secondTask];
					await dropTask(openTaskList, reorderedTasks, task.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: todayDate.toString(),
						done: false,
						trashed: false
					});
					const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
					expectRequestUrl(fetchMock, requestUrl);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(openTaskList, reorderedTasks);
				});

				it('moves a task to the middle index', async () => {
					const task = testData.openPartition.thirdTask;
					const expectedRank = generateKeyBetween(testData.openPartition.firstTask.rank, testData.openPartition.secondTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const openTaskList = page.getByLabelText('Open tasks');
					const reorderedTasks = [testData.openPartition.firstTask, task, testData.openPartition.secondTask];
					await dropTask(openTaskList, reorderedTasks, task.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: todayDate.toString(),
						done: false,
						trashed: false,
						previousTaskId: testData.openPartition.firstTask.id
					});
					const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
					expectRequestUrl(fetchMock, requestUrl);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(openTaskList, reorderedTasks);
				});

				it('moves a task to the last index', async () => {
					const task = testData.openPartition.firstTask;
					const expectedRank = generateKeyBetween(testData.openPartition.thirdTask.rank, undefined);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const openTaskList = page.getByLabelText('Open tasks');
					const reorderedTasks = [testData.openPartition.secondTask, testData.openPartition.thirdTask, task];
					await dropTask(openTaskList, reorderedTasks, task.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: todayDate.toString(),
						done: false,
						trashed: false,
						previousTaskId: testData.openPartition.thirdTask.id
					});
					const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
					expectRequestUrl(fetchMock, requestUrl);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(openTaskList, reorderedTasks);
				});
			});

			describe('within done partition', () => {
				it('moves a task to the first index', async () => {
					const task = testData.donePartition.thirdTask;
					const expectedRank = generateKeyBetween(undefined, testData.donePartition.firstTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const doneTaskList = page.getByLabelText('Done tasks');
					const reorderedTasks = [task, testData.donePartition.firstTask, testData.donePartition.secondTask];
					await dropTask(doneTaskList, reorderedTasks, task.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: todayDate.toString(),
						done: true,
						trashed: false
					});
					const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
					expectRequestUrl(fetchMock, requestUrl);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(doneTaskList, reorderedTasks);
				});

				it('moves a task to the middle index', async () => {
					const task = testData.donePartition.thirdTask;
					const expectedRank = generateKeyBetween(testData.donePartition.firstTask.rank, testData.donePartition.secondTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const doneTaskList = page.getByLabelText('Done tasks');
					const reorderedTasks = [testData.donePartition.firstTask, task, testData.donePartition.secondTask];
					await dropTask(doneTaskList, reorderedTasks, task.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: todayDate.toString(),
						done: true,
						trashed: false,
						previousTaskId: testData.donePartition.firstTask.id
					});
					const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
					expectRequestUrl(fetchMock, requestUrl);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(doneTaskList, reorderedTasks);
				});

				it('moves a task to the last index', async () => {
					const task = testData.donePartition.firstTask;
					const expectedRank = generateKeyBetween(testData.donePartition.thirdTask.rank, undefined);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const doneTaskList = page.getByLabelText('Done tasks');
					const reorderedTasks = [testData.donePartition.secondTask, testData.donePartition.thirdTask, task];
					await dropTask(doneTaskList, reorderedTasks, task.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: todayDate.toString(),
						done: true,
						trashed: false,
						previousTaskId: testData.donePartition.thirdTask.id
					});
					const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
					expectRequestUrl(fetchMock, requestUrl);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(doneTaskList, reorderedTasks);
				});
			});

			describe('from open to done partition', () => {
				it('moves a task to the first index', async () => {
					const task = testData.openPartition.thirdTask;
					const expectedRank = generateKeyBetween(undefined, testData.donePartition.firstTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText('Open tasks');
					const destinationTaskList = page.getByLabelText('Done tasks');
					const reorderedTasks = [task, ...testData.donePartition.tasks];
					await dropTask(destinationTaskList, reorderedTasks, task.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: todayDate.toString(),
						done: true,
						trashed: false
					});
					const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
					expectRequestUrl(fetchMock, requestUrl);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					await waitFor(() => expectTaskNotToBePresentInTaskList(sourceTaskList, task.title));
				});

				it('moves a task to the middle index', async () => {
					const task = testData.openPartition.thirdTask;
					const expectedRank = generateKeyBetween(testData.donePartition.firstTask.rank, testData.donePartition.secondTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText('Open tasks');
					const destinationTaskList = page.getByLabelText('Done tasks');
					const reorderedTasks = [
						testData.donePartition.firstTask,
						task,
						testData.donePartition.secondTask,
						testData.donePartition.thirdTask
					];
					await dropTask(destinationTaskList, reorderedTasks, task.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: todayDate.toString(),
						done: true,
						trashed: false,
						previousTaskId: testData.donePartition.firstTask.id
					});
					const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
					expectRequestUrl(fetchMock, requestUrl);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					await waitFor(() => expectTaskNotToBePresentInTaskList(sourceTaskList, task.title));
				});

				it('moves a task to the last index', async () => {
					const task = testData.openPartition.thirdTask;
					const expectedRank = generateKeyBetween(testData.donePartition.thirdTask.rank, undefined);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText('Open tasks');
					const destinationTaskList = page.getByLabelText('Done tasks');
					const reorderedTasks = [...testData.donePartition.tasks, task];
					await dropTask(destinationTaskList, reorderedTasks, task.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: todayDate.toString(),
						done: true,
						trashed: false,
						previousTaskId: testData.donePartition.thirdTask.id
					});
					const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
					expectRequestUrl(fetchMock, requestUrl);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					await waitFor(() => expectTaskNotToBePresentInTaskList(sourceTaskList, task.title));
				});
			});

			describe('from done to open partition', () => {
				it('moves a task to the first index', async () => {
					const task = testData.donePartition.thirdTask;
					const expectedRank = generateKeyBetween(undefined, testData.openPartition.firstTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText('Done tasks');
					const destinationTaskList = page.getByLabelText('Open tasks');
					const reorderedTasks = [task, ...testData.openPartition.tasks];
					await dropTask(destinationTaskList, reorderedTasks, task.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: todayDate.toString(),
						done: false,
						trashed: false
					});
					const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
					expectRequestUrl(fetchMock, requestUrl);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					await waitFor(() => expectTaskNotToBePresentInTaskList(sourceTaskList, task.title));
				});

				it('moves a task to the middle index', async () => {
					const task = testData.donePartition.thirdTask;
					const expectedRank = generateKeyBetween(testData.openPartition.firstTask.rank, testData.openPartition.secondTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText('Done tasks');
					const destinationTaskList = page.getByLabelText('Open tasks');
					const reorderedTasks = [
						testData.openPartition.firstTask,
						task,
						testData.openPartition.secondTask,
						testData.openPartition.thirdTask
					];
					await dropTask(destinationTaskList, reorderedTasks, task.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: todayDate.toString(),
						done: false,
						trashed: false,
						previousTaskId: testData.openPartition.firstTask.id
					});
					const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
					expectRequestUrl(fetchMock, requestUrl);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					await waitFor(() => expectTaskNotToBePresentInTaskList(sourceTaskList, task.title));
				});

				it('moves a task to the last index', async () => {
					const task = testData.donePartition.thirdTask;
					const expectedRank = generateKeyBetween(testData.openPartition.thirdTask.rank, undefined);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText('Done tasks');
					const destinationTaskList = page.getByLabelText('Open tasks');
					const reorderedTasks = [...testData.openPartition.tasks, task];
					await dropTask(destinationTaskList, reorderedTasks, task.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: todayDate.toString(),
						done: false,
						trashed: false,
						previousTaskId: testData.openPartition.thirdTask.id
					});
					const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
					expectRequestUrl(fetchMock, requestUrl);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					await waitFor(() => expectTaskNotToBePresentInTaskList(sourceTaskList, task.title));
				});
			});
		});

		describe('progress', () => {
			it('displays progress information correctly', () => {
				const expectedOpenTaskCount = testData.openPartition.tasks.length;
				const expectedDoneTaskCount = testData.donePartition.tasks.length;
				const expectedCompletionPercentage = 50;

				const page = render(Page);

				expectProgressInformation(page, expectedOpenTaskCount, expectedDoneTaskCount, expectedCompletionPercentage);
			});

			it('reacts correctly to completion by checkbox', async () => {
				const task = testData.openPartition.firstTask;
				const completedTaskRank = generateKeyBetween(undefined, testData.donePartition.firstTask.rank);
				const response = getResponse({ date: task.date.toString(), rank: completedTaskRank });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const completeTaskButton = page.getByRole('checkbox', {
					name: `Mark "${task.title}" as completed`
				});
				await userEvent.click(completeTaskButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const requestUrl = `/api/tasks/${task.id}/toggle-done`;
				expectRequestUrl(fetchMock, requestUrl);
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				const expectedOpenTaskCount = testData.openPartition.tasks.length - 1;
				const expectedDoneTaskCount = testData.donePartition.tasks.length + 1;
				const expectedCompletionPercentage = 66;
				await waitFor(() =>
					expectProgressInformation(page, expectedOpenTaskCount, expectedDoneTaskCount, expectedCompletionPercentage)
				);
			});

			it('reacts correctly to completion by drag and drop', async () => {
				const task = testData.openPartition.thirdTask;
				const completedTaskRank = generateKeyBetween(testData.donePartition.thirdTask.rank, undefined);
				const response = getResponse({ rank: completedTaskRank });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const doneTaskList = page.getByLabelText('Done tasks');
				const reorderedTasks = [...testData.donePartition.tasks, task];
				await dropTask(doneTaskList, reorderedTasks, task.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const requestBody = JSON.stringify({
					date: todayDate.toString(),
					done: true,
					trashed: false,
					previousTaskId: testData.donePartition.thirdTask.id
				});
				const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
				expectRequestUrl(fetchMock, requestUrl);
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				const expectedOpenTaskCount = testData.openPartition.tasks.length - 1;
				const expectedDoneTaskCount = testData.donePartition.tasks.length + 1;
				const expectedCompletionPercentage = 66;
				await waitFor(() =>
					expectProgressInformation(page, expectedOpenTaskCount, expectedDoneTaskCount, expectedCompletionPercentage)
				);
			});

			it('reacts correctly to opening by checkbox', async () => {
				const task = testData.donePartition.firstTask;
				const openedTaskRank = generateKeyBetween(undefined, testData.openPartition.firstTask.rank);
				const response = getResponse({ date: task.date.toString(), rank: openedTaskRank });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const openTaskButton = page.getByRole('checkbox', {
					name: `Mark "${task.title}" as completed`
				});
				await userEvent.click(openTaskButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const requestUrl = `/api/tasks/${task.id}/toggle-done`;
				expectRequestUrl(fetchMock, requestUrl);
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				const expectedOpenTaskCount = testData.openPartition.tasks.length + 1;
				const expectedDoneTaskCount = testData.donePartition.tasks.length - 1;
				const expectedCompletionPercentage = 33;
				await waitFor(() =>
					expectProgressInformation(page, expectedOpenTaskCount, expectedDoneTaskCount, expectedCompletionPercentage)
				);
			});

			it('reacts correctly to opening by drag and drop', async () => {
				const task = testData.donePartition.thirdTask;
				const openedTaskRank = generateKeyBetween(testData.openPartition.thirdTask.rank, undefined);
				const response = getResponse({ rank: openedTaskRank });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const openTaskList = page.getByLabelText('Open tasks');
				const reorderedTasks = [...testData.openPartition.tasks, task];
				await dropTask(openTaskList, reorderedTasks, task.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const requestBody = JSON.stringify({
					date: todayDate.toString(),
					done: false,
					trashed: false,
					previousTaskId: testData.openPartition.thirdTask.id
				});
				const requestUrl = `/api/tasks/${task.id}/drag-and-drop`;
				expectRequestUrl(fetchMock, requestUrl);
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				const expectedOpenTaskCount = testData.openPartition.tasks.length + 1;
				const expectedDoneTaskCount = testData.donePartition.tasks.length - 1;
				const expectedCompletionPercentage = 33;
				await waitFor(() =>
					expectProgressInformation(page, expectedOpenTaskCount, expectedDoneTaskCount, expectedCompletionPercentage)
				);
			});
		});
	});

	describe('has more tasks scenario', () => {
		beforeEach(() => {
			datedUntrashedOpenTaskStore.addTodayPage(testData.openPartition.tasks, true);
			datedUntrashedDoneTaskStore.addPage(testData.donePartition.tasks, true);
		});

		describe('infinite scroll', () => {
			it('loads more open tasks without applied filters', async () => {
				const nextTaskResponse = {
					id: 7,
					categoryId: personalCategory.id,
					title: 'Plan the weekend groceries',
					date: todayDate.toString(),
					rank: generateKeyBetween(testData.openPartition.thirdTask.rank, undefined)
				};
				const nextTask: DatedUntrashedOpenTaskResponse = {
					...nextTaskResponse,
					date: todayDate,
					done: false
				};
				const existingDoneTasks = [...datedUntrashedDoneTaskStore.datedUntrashedDoneTasks];
				const resolveFetch = mockPendingFetch(fetchMock);

				const page = render(Page);
				const openTaskColumn = getTaskColumn(page, 'Open');
				const loadingSentinel = getLoadingSentinel(openTaskColumn);
				intersectionObserverMock.enterNode(loadingSentinel);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, '/api/tasks/dated-untrashed-open-today');
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expect(getRequestSearchParameter(fetchMock, 'rank')).toBe(testData.openPartition.thirdTask.rank);
				expect(within(openTaskColumn).getByLabelText('Loading more items')).toBeInTheDocument();
				expect(within(getTaskColumn(page, 'Done')).queryByLabelText('Loading more items')).not.toBeInTheDocument();

				resolveFetch(getResponse({ tasks: [nextTaskResponse], hasMore: false }));

				await waitFor(() => expect(datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks).toContainEqual(nextTask));
				expect(datedUntrashedOpenTaskStore.todayTasksHasMore).toBe(false);
				expectTaskToBePresent(openTaskColumn, nextTask.title, nextTask.description, personalCategory.title);
				expect(openTaskColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toEqual(existingDoneTasks);
			});

			it('loads more open tasks with applied filters', async () => {
				const filteredOpenTasks = [testData.openPartition.secondTask, testData.openPartition.thirdTask];
				const filteredDoneTasks = [testData.donePartition.thirdTask];
				mockFilterResponses(filteredOpenTasks, filteredDoneTasks, true);

				const page = render(Page);
				const excludePersonalButtons = page.getAllByRole('button', { name: 'Exclude Personal' });
				await userEvent.click(excludePersonalButtons[0]!);
				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
				fetchMock.mockClear();

				const nextTaskResponse = {
					id: 7,
					categoryId: workCategory.id,
					title: "Prepare tomorrow's stand-up notes",
					date: todayDate.toString(),
					rank: generateKeyBetween(testData.openPartition.thirdTask.rank, undefined)
				};
				const nextTask: DatedUntrashedOpenTaskResponse = {
					...nextTaskResponse,
					date: todayDate,
					done: false
				};
				const resolveFetch = mockPendingFetch(fetchMock);
				const openTaskColumn = getTaskColumn(page, 'Open');
				intersectionObserverMock.enterNode(getLoadingSentinel(openTaskColumn));

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, '/api/tasks/dated-untrashed-open-today');
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expect(getRequestSearchParameter(fetchMock, 'categoryId')).toBe(String(workCategory.id));
				expect(getRequestSearchParameter(fetchMock, 'rank')).toBe(testData.openPartition.thirdTask.rank);

				resolveFetch(getResponse({ tasks: [nextTaskResponse], hasMore: false }));

				await waitFor(() => expect(datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks).toContainEqual(nextTask));
				expect(datedUntrashedOpenTaskStore.todayTasksHasMore).toBe(false);
				expectTaskToBePresent(openTaskColumn, nextTask.title, nextTask.description, workCategory.title);
				expect(openTaskColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
			});

			it('loads more done tasks without applied filters', async () => {
				const nextTaskResponse = {
					id: 7,
					categoryId: personalCategory.id,
					title: 'Confirm the dentist appointment',
					date: todayDate.toString(),
					rank: generateKeyBetween(testData.donePartition.thirdTask.rank, undefined)
				};
				const nextTask: DatedUntrashedDoneTaskResponse = {
					...nextTaskResponse,
					date: todayDate,
					done: true
				};
				const existingOpenTasks = [...datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks];
				const resolveFetch = mockPendingFetch(fetchMock);

				const page = render(Page);
				const doneTaskColumn = getTaskColumn(page, 'Done');
				const loadingSentinel = getLoadingSentinel(doneTaskColumn);
				intersectionObserverMock.enterNode(loadingSentinel);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, '/api/tasks/dated-untrashed-done-today');
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expect(getRequestSearchParameter(fetchMock, 'rank')).toBe(testData.donePartition.thirdTask.rank);
				expect(within(doneTaskColumn).getByLabelText('Loading more items')).toBeInTheDocument();
				expect(within(getTaskColumn(page, 'Open')).queryByLabelText('Loading more items')).not.toBeInTheDocument();

				resolveFetch(getResponse({ tasks: [nextTaskResponse], hasMore: false }));

				await waitFor(() => expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toContainEqual(nextTask));
				expect(datedUntrashedDoneTaskStore.hasMore).toBe(false);
				expectTaskToBePresent(doneTaskColumn, nextTask.title, nextTask.description, personalCategory.title);
				expect(doneTaskColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks).toEqual(existingOpenTasks);
			});

			it('loads more done tasks with applied filters', async () => {
				const filteredOpenTasks = [testData.openPartition.secondTask, testData.openPartition.thirdTask];
				const filteredDoneTasks = [testData.donePartition.thirdTask];
				mockFilterResponses(filteredOpenTasks, filteredDoneTasks, true);

				const page = render(Page);
				const excludePersonalButtons = page.getAllByRole('button', { name: 'Exclude Personal' });
				await userEvent.click(excludePersonalButtons[0]!);
				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
				fetchMock.mockClear();

				const nextTaskResponse = {
					id: 7,
					categoryId: workCategory.id,
					title: 'Publish the weekly status update',
					date: todayDate.toString(),
					rank: generateKeyBetween(testData.donePartition.thirdTask.rank, undefined)
				};
				const nextTask: DatedUntrashedDoneTaskResponse = {
					...nextTaskResponse,
					date: todayDate,
					done: true
				};
				const resolveFetch = mockPendingFetch(fetchMock);
				const doneTaskColumn = getTaskColumn(page, 'Done');
				intersectionObserverMock.enterNode(getLoadingSentinel(doneTaskColumn));

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, '/api/tasks/dated-untrashed-done-today');
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expect(getRequestSearchParameter(fetchMock, 'categoryId')).toBe(String(workCategory.id));
				expect(getRequestSearchParameter(fetchMock, 'rank')).toBe(testData.donePartition.thirdTask.rank);

				resolveFetch(getResponse({ tasks: [nextTaskResponse], hasMore: false }));

				await waitFor(() => expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toContainEqual(nextTask));
				expect(datedUntrashedDoneTaskStore.hasMore).toBe(false);
				expectTaskToBePresent(doneTaskColumn, nextTask.title, nextTask.description, workCategory.title);
				expect(doneTaskColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
			});

			it('disables infinite scrolling while a task is being dragged', async () => {
				const page = render(Page);
				const openTaskColumn = getTaskColumn(page, 'Open');
				const doneTaskColumn = getTaskColumn(page, 'Done');
				const openLoadingSentinel = getLoadingSentinel(openTaskColumn);
				const doneLoadingSentinel = getLoadingSentinel(doneTaskColumn);

				const openTaskList = page.getByLabelText('Open tasks');
				const event = new CustomEvent<DndEvent<DatedUntrashedOpenTaskResponse>>('consider', {
					detail: {
						items: testData.openPartition.tasks,
						info: {
							trigger: TRIGGERS.DRAG_STARTED,
							id: String(testData.openPartition.firstTask.id),
							source: SOURCES.POINTER
						}
					}
				});
				await fireEvent(openTaskList, event);

				intersectionObserverMock.enterNodes([openLoadingSentinel, doneLoadingSentinel]);
				expect(fetchMock).not.toHaveBeenCalled();
			});
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

async function openCreateTaskDialog(page: RenderResult<typeof Page>): Promise<void> {
	const addTaskButton = page.getByRole('button', { name: 'Add Task' });
	await userEvent.click(addTaskButton);
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

async function selectCategory(page: RenderResult<typeof Page>, categoryName: string): Promise<void> {
	const categorySelect = getSelectById(page, 'task-category');
	categorySelect.focus();
	const firstCategoryNameCharacter = categoryName[0];
	const categoryTypeaheadKey = firstCategoryNameCharacter?.toLowerCase();
	await fireEvent.keyDown(categorySelect, { key: categoryTypeaheadKey });
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

function mockFilterResponses(
	openTasks: DatedUntrashedOpenTaskResponse[],
	doneTasks: DatedUntrashedDoneTaskResponse[],
	hasMore = false
): void {
	const serializeTasks = (tasks: (DatedUntrashedOpenTaskResponse | DatedUntrashedDoneTaskResponse)[]) =>
		tasks.map((task) => ({
			id: task.id,
			categoryId: task.categoryId,
			title: task.title,
			description: task.description,
			date: task.date.toString(),
			rank: task.rank
		}));
	const openTasksResponse = getResponse({ tasks: serializeTasks(openTasks), hasMore });
	const doneTasksResponse = getResponse({ tasks: serializeTasks(doneTasks), hasMore });
	fetchMock.mockResolvedValueOnce(openTasksResponse).mockResolvedValueOnce(doneTasksResponse);
}

function getTaskColumn(page: RenderResult<typeof Page>, heading: string): HTMLElement {
	const taskColumn = page.getByRole('heading', { name: heading }).closest('section');
	if (!(taskColumn instanceof HTMLElement)) throw new Error(`Expected a task column headed "${heading}".`);
	return taskColumn;
}

function getLoadingSentinel(taskColumn: HTMLElement): HTMLElement {
	const loadingSentinel = taskColumn.querySelector('.loading-sentinel');
	if (!(loadingSentinel instanceof HTMLElement)) throw new Error('Expected the task column to have a loading sentinel.');
	return loadingSentinel;
}

async function dropTask(
	taskList: HTMLElement,
	reorderedTasks: (DatedUntrashedOpenTaskResponse | DatedUntrashedDoneTaskResponse)[],
	draggedTaskId: number
): Promise<void> {
	const event = new CustomEvent<DndEvent<DatedUntrashedOpenTaskResponse | DatedUntrashedDoneTaskResponse>>('finalize', {
		detail: {
			items: reorderedTasks,
			info: { trigger: TRIGGERS.DROPPED_INTO_ZONE, id: String(draggedTaskId), source: SOURCES.POINTER }
		}
	});
	await fireEvent(taskList, event);
}

function expectTaskOrder(
	taskList: HTMLElement,
	tasks: (DatedUntrashedOpenTaskResponse | DatedUntrashedDoneTaskResponse)[]
): void {
	const taskListContent = within(taskList);
	const renderedTasks = taskListContent.getAllByRole('listitem');
	const renderedTaskTitles = renderedTasks.map((task) => task.getAttribute('aria-label'));
	const expectedTaskTitles = tasks.map((task) => task.title);
	expect(renderedTaskTitles).toEqual(expectedTaskTitles);
}

function expectTaskNotToBePresentInTaskList(taskList: HTMLElement, title: string): void {
	const task = within(taskList).queryByRole('listitem', { name: title });
	expect(task).not.toBeInTheDocument();
}

function expectProgressInformation(
	page: RenderResult<typeof Page>,
	openTaskCount: number,
	doneTaskCount: number,
	completionPercentage: number
): void {
	const progressSection = getTaskColumn(page, 'Progress');
	const progressSectionContent = within(progressSection);
	const openTaskCountContainer = progressSectionContent.getByText('open').parentElement;
	const doneTaskCountContainer = progressSectionContent.getByText('done').parentElement;
	const progressBar = progressSectionContent.getByRole('progressbar', { name: "Today's task completion" });
	const totalTaskCount = openTaskCount + doneTaskCount;
	const expectedValueText = `${completionPercentage}% (${doneTaskCount} of ${totalTaskCount} tasks done)`;

	expect(openTaskCountContainer).toHaveTextContent(`${openTaskCount} open`);
	expect(doneTaskCountContainer).toHaveTextContent(`${doneTaskCount} done`);
	expect(progressBar).toHaveAttribute('aria-valuenow', String(completionPercentage));
	expect(progressBar).toHaveAttribute('aria-valuetext', expectedValueText);
	expect(progressSectionContent.getByText(`${completionPercentage}%`)).toBeInTheDocument();
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
