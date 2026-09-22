import { fireEvent, render, type RenderResult, waitFor, within } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { getLocalTimeZone, now, parseAbsolute, today } from '@internationalized/date';
import { generateKeyBetween, generateNKeysBetween } from 'fractional-indexing';
import { toast } from 'svelte-sonner';
import { type DndEvent, SOURCES, TRIGGERS } from 'svelte-dnd-action';
import { categoryStore } from '$lib/stores/category-store.svelte';
import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
import { trashedTaskStore } from '$lib/stores/trashed-task-store.svelte';
import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
import type { DatedUntrashedOpenTaskResponse } from '$lib/types/responses/dated-untrashed-open-task-response';
import type { UndatedTaskResponse } from '$lib/types/responses/undated-task-response';
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

describe('backlog page tests', () => {
	beforeEach(() => {
		vi.unstubAllGlobals();
		vi.resetAllMocks();
		vi.stubGlobal('fetch', fetchMock);
		categoryStore.reset();
		undatedTaskStore.reset();
		datedUntrashedDoneTaskStore.reset();
		datedUntrashedOpenTaskStore.reset();
		trashedTaskStore.reset();
		categoryStore.setCategories([category]);
	});

	afterEach(async () => {
		if (document.querySelector('[role="dialog"]')) await userEvent.keyboard('{Escape}');
		document.body.removeAttribute('style');
	});

	describe('has no more tasks scenario', () => {
		beforeEach(() => {
			undatedTaskStore.addOrUpdateTasks(testData.noDatePartition.tasks);
			datedUntrashedOpenTaskStore.addOrUpdateTasks([
				...testData.todayPartition.tasks,
				...testData.nonTodayPastPartition.tasks,
				...testData.nonTodayFuturePartition.tasks
			]);
		});

		describe('category operations', () => {
			it('successfully creates and deletes a category', async () => {
				const newCategoryId = 8;
				const newCategoryTitle = 'Work';
				const newCategoryColor = '#3b82f6';

				const createCategoryResponse = getResponse({ id: newCategoryId }, 201);
				fetchMock.mockResolvedValue(createCategoryResponse);

				const page = render(Page);
				await openCreateTaskDialog(page);
				await startCategoryCreation(page);
				const categoryNameInput = page.getByRole('textbox', { name: 'Category name' });
				await fireEvent.input(categoryNameInput, { target: { value: newCategoryTitle } });
				const categoryColorInput = page.getByLabelText('Choose category color');
				await fireEvent.input(categoryColorInput, { target: { value: newCategoryColor } });
				const createCategoryButton = page.getByRole('button', { name: 'Create category' });
				await userEvent.click(createCategoryButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const createRequestBody = JSON.stringify({ title: newCategoryTitle, color: newCategoryColor });
				expectRequestUrl(fetchMock, '/api/categories');
				expectRequestMethod(fetchMock, 'POST');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, createRequestBody);
				expectRequestContentType(fetchMock, 'application/json');
				const categorySelect = getSelectById(page, 'task-category');
				expect(categorySelect).toHaveTextContent(newCategoryTitle);

				fetchMock.mockClear();
				const deleteCategoryResponse = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(deleteCategoryResponse);

				await openCategorySelect(page);
				const deleteCategoryButton = await page.findByRole('button', { name: `Delete ${newCategoryTitle}` });
				await userEvent.click(deleteCategoryButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestUrl(fetchMock, `/api/categories/${newCategoryId}`);
				expectRequestMethod(fetchMock, 'DELETE');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				await waitFor(() => expect(categorySelect).toHaveTextContent('None chosen'));
				expect(categoryStore.categories).not.toContainEqual({
					id: newCategoryId,
					title: newCategoryTitle,
					color: newCategoryColor
				});
			});

			it('shows an error toast when deleting a referenced category', async () => {
				const errorMessage = 'Category cannot be deleted while it is referenced by tasks.';
				const errorResponse = getResponse({ errorCode: 'category_still_referenced', message: errorMessage }, 409);

				fetchMock.mockResolvedValue(errorResponse);

				const page = render(Page);
				await openCreateTaskDialog(page);
				await openCategorySelect(page);
				const deleteCategoryButton = await page.findByRole('button', { name: `Delete ${category.title}` });
				await userEvent.click(deleteCategoryButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestUrl(fetchMock, `/api/categories/${category.id}`);
				expectRequestMethod(fetchMock, 'DELETE');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expect(toastErrorMock).toHaveBeenCalledOnce();
				expect(toastErrorMock).toHaveBeenCalledWith(errorMessage);
				expect(categoryStore.categories).toContainEqual(category);
				expect(deleteCategoryButton).toBeInTheDocument();
			});
		});

		describe('task creation', () => {
			it('reacts correctly to missing task title', async () => {
				const page = render(Page);
				await openCreateTaskDialog(page);
				await selectCategory(page, category.title);
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
				await selectCategory(page, category.title);
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
				await selectCategory(page, category.title);
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

			it('successfully adds a task', async () => {
				const title = 'Prepare the quarterly presentation';
				const description = 'Summarize project results and prepare the final slides.';
				const rank = generateKeyBetween(testData.noDatePartition.thirdTask.rank, undefined);

				const response = getResponse({ id: 13, rank }, 201);
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				await openCreateTaskDialog(page);
				await enterTaskTitle(page, title);
				await enterTaskDescription(page, description);
				await selectCategory(page, category.title);
				const createButton = page.getByRole('button', { name: 'Create' });
				await userEvent.click(createButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const createTaskDialog = page.queryByRole('dialog', { name: 'Create Task' });
				expect(createTaskDialog).not.toBeInTheDocument();
				const requestBody = JSON.stringify({
					title,
					description,
					categoryId: category.id
				});
				expectRequestUrl(fetchMock, '/api/tasks');
				expectRequestMethod(fetchMock, 'POST');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
				expectTaskToBePresent(tasksWithoutDateList, title, description, category.title);
			});
		});

		describe('task editing', () => {
			it('reacts correctly to missing task title', async () => {
				const task = testData.noDatePartition.firstTask;

				const page = render(Page);
				await openEditTaskDialog(page, task.title);
				const titleInput = await enterTaskTitle(page, '');
				await fireEvent.blur(titleInput);

				const saveButton = page.getByRole('button', { name: 'Save' });
				expect(saveButton).toBeDisabled();
				await expectInputToBeInvalid(page, 'task-title', 'Title must have at least 1 character');
			});

			it('reacts correctly to too long task title', async () => {
				const task = testData.noDatePartition.firstTask;

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
				const task = testData.noDatePartition.firstTask;

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
				const task = testData.noDatePartition.firstTask;

				const page = render(Page);
				await openEditTaskDialog(page, task.title);
				const cancelButton = page.getByRole('button', { name: 'Cancel' });
				await userEvent.click(cancelButton);

				const editTaskDialog = page.queryByRole('dialog', { name: 'Edit Task' });
				expect(editTaskDialog).not.toBeInTheDocument();
			});

			it('successfully edits a task without a date', async () => {
				const task = testData.noDatePartition.firstTask;
				const newTitle = 'Book the hiking trail transfer';
				const newDescription = 'Reserve the earliest train and the return connection.';

				const response = getResponse({
					id: task.id,
					categoryId: category.id,
					title: newTitle,
					description: newDescription,
					done: false,
					rank: task.rank
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
					categoryId: category.id,
					id: task.id
				});
				expectRequestUrl(fetchMock, '/api/tasks');
				expectRequestMethod(fetchMock, 'PUT');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
				expectTaskToBePresent(tasksWithoutDateList, newTitle, newDescription, category.title);
			});

			it('successfully edits a task with a non-today date', async () => {
				const task = testData.nonTodayPastPartition.firstTask;
				const taskDate = testData.nonTodayPastPartitionDate;
				const newTitle = 'Book the hiking trail transfer';
				const newDescription = 'Reserve the earliest train and the return connection.';

				const response = getResponse({
					id: task.id,
					categoryId: category.id,
					title: newTitle,
					description: newDescription,
					date: taskDate.toString(),
					done: false,
					rank: task.rank
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
				const serializedDate = taskDate.toString();
				const requestBody = JSON.stringify({
					title: newTitle,
					description: newDescription,
					categoryId: category.id,
					date: serializedDate,
					id: task.id
				});
				expectRequestUrl(fetchMock, '/api/tasks');
				expectRequestMethod(fetchMock, 'PUT');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				const datedTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayPastPartitionLabel}`);
				expectTaskToBePresent(datedTaskList, newTitle, newDescription, category.title);
			});

			it('successfully edits a task dated today', async () => {
				const task = testData.todayPartition.firstTask;
				const taskDate = testData.todayPartitionDate;
				const newTitle = 'Book the hiking trail transfer';
				const newDescription = 'Reserve the earliest train and the return connection.';

				const response = getResponse({
					id: task.id,
					categoryId: category.id,
					title: newTitle,
					description: newDescription,
					date: taskDate.toString(),
					done: false,
					rank: task.rank
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
				const serializedDate = taskDate.toString();
				const requestBody = JSON.stringify({
					title: newTitle,
					description: newDescription,
					categoryId: category.id,
					date: serializedDate,
					id: task.id
				});
				expectRequestUrl(fetchMock, '/api/tasks');
				expectRequestMethod(fetchMock, 'PUT');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				const todayTaskList = page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`);
				expectTaskToBePresent(todayTaskList, newTitle, newDescription, category.title);
			});
		});

		describe('task completion', () => {
			it('successfully completes a task without a date', async () => {
				const task = testData.noDatePartition.firstTask;
				const completedTaskDate = testData.todayPartitionDate;
				const completedTaskRank = generateKeyBetween(undefined, undefined);

				const response = getResponse({ date: completedTaskDate.toString(), rank: completedTaskRank });
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
				expectTaskNotToBePresent(page, task.title);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toContainEqual({
					...task,
					date: completedTaskDate,
					rank: completedTaskRank,
					done: true
				});
			});

			it('successfully completes a task with a non-today date', async () => {
				const task = testData.nonTodayPastPartition.firstTask;
				const completedTaskRank = generateKeyBetween(undefined, undefined);

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
				expectTaskNotToBePresent(page, task.title);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toContainEqual({
					...task,
					rank: completedTaskRank,
					done: true
				});
			});

			it('successfully completes a task dated today', async () => {
				const task = testData.todayPartition.firstTask;
				const completedTaskRank = generateKeyBetween(undefined, undefined);

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
				expectTaskNotToBePresent(page, task.title);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toContainEqual({
					...task,
					rank: completedTaskRank,
					done: true
				});
			});

			it('reports potentially inconsistent data when completing a dated task returns no rank', async () => {
				const task = testData.todayPartition.firstTask;

				const response = getResponse({ date: task.date.toString() });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const completeTaskButton = page.getByRole('checkbox', {
					name: `Mark "${task.title}" as completed`
				});
				await userEvent.click(completeTaskButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expect(toastErrorMock).toHaveBeenCalledWith(
					'An unexpected error occurred. Your data might be inconsistent. Please reload the page to resolve the issue.'
				);
				expectTaskToBePresent(
					page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`),
					task.title,
					task.description,
					category.title
				);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toHaveLength(0);
			});
		});

		describe('task trashing', () => {
			it('successfully trashes a task without a date', async () => {
				const task = testData.noDatePartition.firstTask;

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const moreOptionsButton = page.getByRole('button', {
					name: `More options for "${task.title}"`
				});
				await fireEvent.click(moreOptionsButton);
				const trashButton = await page.findByRole('button', { name: 'Trash' });
				await userEvent.click(trashButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, `/api/tasks/${task.id}/trash`);
				const trashedDate = getRequestSearchParameter(fetchMock, 'trashedDate');
				expect(trashedDate).not.toBeNull();
				if (trashedDate === null) throw new Error('Expected the trash request to contain a trashedDate.');
				const parsedTrashedDate = parseAbsolute(trashedDate, localTimeZone);
				const currentDate = now(localTimeZone);
				const dateDifference = Math.abs(parsedTrashedDate.compare(currentDate));
				expect(dateDifference).toBeLessThanOrEqual(maximumTrashTimestampDifferenceMilliseconds);
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expectTaskNotToBePresent(page, task.title);
			});

			it('successfully trashes a task with a non-today date', async () => {
				const task = testData.nonTodayPastPartition.firstTask;

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const moreOptionsButton = page.getByRole('button', {
					name: `More options for "${task.title}"`
				});
				await fireEvent.click(moreOptionsButton);
				const trashButton = await page.findByRole('button', { name: 'Trash' });
				await userEvent.click(trashButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, `/api/tasks/${task.id}/trash`);
				const trashedDate = getRequestSearchParameter(fetchMock, 'trashedDate');
				expect(trashedDate).not.toBeNull();
				if (trashedDate === null) throw new Error('Expected the trash request to contain a trashedDate.');
				const parsedTrashedDate = parseAbsolute(trashedDate, localTimeZone);
				const currentDate = now(localTimeZone);
				const dateDifference = Math.abs(parsedTrashedDate.compare(currentDate));
				expect(dateDifference).toBeLessThanOrEqual(maximumTrashTimestampDifferenceMilliseconds);
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expectTaskNotToBePresent(page, task.title);
			});

			it('successfully trashes a task dated today', async () => {
				const task = testData.todayPartition.firstTask;

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const moreOptionsButton = page.getByRole('button', {
					name: `More options for "${task.title}"`
				});
				await fireEvent.click(moreOptionsButton);
				const trashButton = await page.findByRole('button', { name: 'Trash' });
				await userEvent.click(trashButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, `/api/tasks/${task.id}/trash`);
				const trashedDate = getRequestSearchParameter(fetchMock, 'trashedDate');
				expect(trashedDate).not.toBeNull();
				if (trashedDate === null) throw new Error('Expected the trash request to contain a trashedDate.');
				const parsedTrashedDate = parseAbsolute(trashedDate, localTimeZone);
				const currentDate = now(localTimeZone);
				const dateDifference = Math.abs(parsedTrashedDate.compare(currentDate));
				expect(dateDifference).toBeLessThanOrEqual(maximumTrashTimestampDifferenceMilliseconds);
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expectTaskNotToBePresent(page, task.title);
			});
		});

		describe('task batch operations', () => {
			it('successfully completes all tasks', async () => {
				const tasks = [
					testData.noDatePartition.firstTask,
					testData.nonTodayPastPartition.secondTask,
					testData.todayPartition.thirdTask
				] as const;
				const taskIds = tasks.map((task) => task.id);
				const completedTaskRanks = generateNKeysBetween(undefined, undefined, tasks.length);
				const completedTasks = [
					{
						id: tasks[0].id,
						categoryId: tasks[0].categoryId,
						title: tasks[0].title,
						description: tasks[0].description,
						date: testData.todayPartitionDate.toString(),
						rank: completedTaskRanks[0]
					},
					{
						id: tasks[1].id,
						categoryId: tasks[1].categoryId,
						title: tasks[1].title,
						description: tasks[1].description,
						date: testData.nonTodayPastPartitionDate.toString(),
						rank: completedTaskRanks[1]
					},
					{
						id: tasks[2].id,
						categoryId: tasks[2].categoryId,
						title: tasks[2].title,
						description: tasks[2].description,
						date: testData.todayPartitionDate.toString(),
						rank: completedTaskRanks[2]
					}
				];

				const response = getResponse({ datedUntrashedDoneTasks: completedTasks, trashedTasks: [] });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const task of tasks) await clickOnTask(page, task.title);
				const completeAllButton = page.getByRole('button', { name: 'Complete All' });
				await userEvent.click(completeAllButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const requestBody = JSON.stringify({ taskIds });
				expectRequestUrl(fetchMock, '/api/tasks/batch-complete');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				for (const task of tasks) expectTaskNotToBePresent(page, task.title);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toHaveLength(3);
				const selectTasksButton = page.getByRole('button', { name: 'Select Tasks' });
				expect(selectTasksButton).toBeInTheDocument();
			});

			it('successfully trashes all tasks', async () => {
				const tasks = [
					testData.noDatePartition.firstTask,
					testData.nonTodayPastPartition.secondTask,
					testData.todayPartition.thirdTask
				] as const;
				const taskIds = tasks.map((task) => task.id);

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const task of tasks) await clickOnTask(page, task.title);
				const trashAllButton = page.getByRole('button', { name: 'Trash All' });
				await userEvent.click(trashAllButton);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, '/api/tasks/batch-trash');
				const trashedDate = getRequestSearchParameter(fetchMock, 'trashedDate');
				expect(trashedDate).not.toBeNull();
				if (trashedDate === null) throw new Error('Expected the batch trash request to contain a trashedDate.');
				const parsedTrashedDate = parseAbsolute(trashedDate, localTimeZone);
				const currentDate = now(localTimeZone);
				const dateDifference = Math.abs(parsedTrashedDate.compare(currentDate));
				expect(dateDifference).toBeLessThanOrEqual(maximumTrashTimestampDifferenceMilliseconds);
				const requestBody = JSON.stringify({ taskIds });
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				for (const task of tasks) expectTaskNotToBePresent(page, task.title);
				expect(trashedTaskStore.trashedTasks).toHaveLength(3);
				const selectTasksButton = page.getByRole('button', { name: 'Select Tasks' });
				expect(selectTasksButton).toBeInTheDocument();
			});

			it('successfully cancels a batch operation', async () => {
				const task = testData.noDatePartition.firstTask;

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				await clickOnTask(page, task.title);
				const cancelSelectionButton = page.getByRole('button', { name: 'Cancel Selection' });
				await userEvent.click(cancelSelectionButton);

				const selectTasksButton = page.getByRole('button', { name: 'Select Tasks' });
				expect(selectTasksButton).toBeInTheDocument();
				const completeTaskButton = page.getByRole('checkbox', {
					name: `Mark "${task.title}" as completed`
				});
				expect(completeTaskButton).toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});

			it('disables batch actions when no tasks are selected', async () => {
				const page = render(Page);
				await clickOnSelectTasksButton(page);

				const completeAllButton = page.getByRole('button', { name: 'Complete All' });
				const trashAllButton = page.getByRole('button', { name: 'Trash All' });
				expect(completeAllButton).toBeDisabled();
				expect(trashAllButton).toBeDisabled();
				expect(page.getByText('0 / 50')).toBeInTheDocument();
			});

			it('disables batch trash when more than 50 tasks are selected', async () => {
				const taskTitles = addTasksExceedingMaximumBatchSize();

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const title of taskTitles) await clickOnTask(page, title);

				const trashAllButton = page.getByRole('button', { name: 'Trash All' });
				expect(trashAllButton).toBeDisabled();
				expect(page.getByText('51 / 50')).toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});

			it('disables batch completion when more than 50 tasks are selected', async () => {
				const taskTitles = addTasksExceedingMaximumBatchSize();

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const title of taskTitles) await clickOnTask(page, title);

				const completeAllButton = page.getByRole('button', { name: 'Complete All' });
				expect(completeAllButton).toBeDisabled();
				expect(page.getByText('51 / 50')).toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});
		});

		describe('drag and drop', () => {
			describe('within the no-date partition', () => {
				it('moves a task to the first index', async () => {
					const expectedRank = generateKeyBetween(undefined, testData.noDatePartition.firstTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
					const reorderedTasks = [
						testData.noDatePartition.thirdTask,
						testData.noDatePartition.firstTask,
						testData.noDatePartition.secondTask
					];
					await dropTask(tasksWithoutDateList, reorderedTasks, testData.noDatePartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.noDatePartition.thirdTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(tasksWithoutDateList, reorderedTasks);
				});

				it('moves a task to the middle index', async () => {
					const expectedRank = generateKeyBetween(
						testData.noDatePartition.firstTask.rank,
						testData.noDatePartition.secondTask.rank
					);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
					const reorderedTasks = [
						testData.noDatePartition.firstTask,
						testData.noDatePartition.thirdTask,
						testData.noDatePartition.secondTask
					];
					await dropTask(tasksWithoutDateList, reorderedTasks, testData.noDatePartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({ previousTaskId: testData.noDatePartition.firstTask.id });
					expectRequestUrl(fetchMock, `/api/tasks/${testData.noDatePartition.thirdTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(tasksWithoutDateList, reorderedTasks);
				});

				it('moves a task to the last index', async () => {
					const expectedRank = generateKeyBetween(testData.noDatePartition.thirdTask.rank, undefined);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
					const reorderedTasks = [
						testData.noDatePartition.secondTask,
						testData.noDatePartition.thirdTask,
						testData.noDatePartition.firstTask
					];
					await dropTask(tasksWithoutDateList, reorderedTasks, testData.noDatePartition.firstTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({ previousTaskId: testData.noDatePartition.thirdTask.id });
					expectRequestUrl(fetchMock, `/api/tasks/${testData.noDatePartition.firstTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(tasksWithoutDateList, reorderedTasks);
				});
			});

			describe('within the today partition', () => {
				it('moves a task to the first index', async () => {
					const expectedRank = generateKeyBetween(undefined, testData.todayPartition.firstTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const todayTaskList = page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`);
					const reorderedTasks = [
						testData.todayPartition.thirdTask,
						testData.todayPartition.firstTask,
						testData.todayPartition.secondTask
					];
					await dropTask(todayTaskList, reorderedTasks, testData.todayPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.todayPartitionDate.toString()
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.todayPartition.thirdTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(todayTaskList, reorderedTasks);
				});

				it('moves a task to the middle index', async () => {
					const expectedRank = generateKeyBetween(
						testData.todayPartition.firstTask.rank,
						testData.todayPartition.secondTask.rank
					);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const todayTaskList = page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`);
					const reorderedTasks = [
						testData.todayPartition.firstTask,
						testData.todayPartition.thirdTask,
						testData.todayPartition.secondTask
					];
					await dropTask(todayTaskList, reorderedTasks, testData.todayPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.todayPartitionDate.toString(),
						previousTaskId: testData.todayPartition.firstTask.id
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.todayPartition.thirdTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(todayTaskList, reorderedTasks);
				});

				it('moves a task to the last index', async () => {
					const expectedRank = generateKeyBetween(testData.todayPartition.thirdTask.rank, undefined);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const todayTaskList = page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`);
					const reorderedTasks = [
						testData.todayPartition.secondTask,
						testData.todayPartition.thirdTask,
						testData.todayPartition.firstTask
					];
					await dropTask(todayTaskList, reorderedTasks, testData.todayPartition.firstTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.todayPartitionDate.toString(),
						previousTaskId: testData.todayPartition.thirdTask.id
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.todayPartition.firstTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(todayTaskList, reorderedTasks);
				});
			});

			describe('within the non-today-date partition', () => {
				it('moves a task to the first index', async () => {
					const expectedRank = generateKeyBetween(undefined, testData.nonTodayPastPartition.firstTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const nonTodayTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayPastPartitionLabel}`);
					const reorderedTasks = [
						testData.nonTodayPastPartition.thirdTask,
						testData.nonTodayPastPartition.firstTask,
						testData.nonTodayPastPartition.secondTask
					];
					await dropTask(nonTodayTaskList, reorderedTasks, testData.nonTodayPastPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.nonTodayPastPartitionDate.toString()
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.nonTodayPastPartition.thirdTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(nonTodayTaskList, reorderedTasks);
				});

				it('moves a task to the middle index', async () => {
					const expectedRank = generateKeyBetween(
						testData.nonTodayPastPartition.firstTask.rank,
						testData.nonTodayPastPartition.secondTask.rank
					);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const nonTodayTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayPastPartitionLabel}`);
					const reorderedTasks = [
						testData.nonTodayPastPartition.firstTask,
						testData.nonTodayPastPartition.thirdTask,
						testData.nonTodayPastPartition.secondTask
					];
					await dropTask(nonTodayTaskList, reorderedTasks, testData.nonTodayPastPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.nonTodayPastPartitionDate.toString(),
						previousTaskId: testData.nonTodayPastPartition.firstTask.id
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.nonTodayPastPartition.thirdTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(nonTodayTaskList, reorderedTasks);
				});

				it('moves a task to the last index', async () => {
					const expectedRank = generateKeyBetween(testData.nonTodayPastPartition.thirdTask.rank, undefined);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const nonTodayTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayPastPartitionLabel}`);
					const reorderedTasks = [
						testData.nonTodayPastPartition.secondTask,
						testData.nonTodayPastPartition.thirdTask,
						testData.nonTodayPastPartition.firstTask
					];
					await dropTask(nonTodayTaskList, reorderedTasks, testData.nonTodayPastPartition.firstTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.nonTodayPastPartitionDate.toString(),
						previousTaskId: testData.nonTodayPastPartition.thirdTask.id
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.nonTodayPastPartition.firstTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(nonTodayTaskList, reorderedTasks);
				});
			});

			describe('between different non-today-date partitions', () => {
				it('moves a task to the first index', async () => {
					const expectedRank = generateKeyBetween(undefined, testData.nonTodayFuturePartition.firstTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayPastPartitionLabel}`);
					const destinationTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayFuturePartitionLabel}`);
					const reorderedTasks = [
						testData.nonTodayPastPartition.firstTask,
						testData.nonTodayFuturePartition.firstTask,
						testData.nonTodayFuturePartition.secondTask,
						testData.nonTodayFuturePartition.thirdTask
					];
					await dropTask(destinationTaskList, reorderedTasks, testData.nonTodayPastPartition.firstTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.nonTodayFuturePartitionDate.toString()
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.nonTodayPastPartition.firstTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(
						within(sourceTaskList).queryByRole('listitem', { name: testData.nonTodayPastPartition.firstTask.title })
					).not.toBeInTheDocument();
				});

				it('moves a task to the middle index', async () => {
					const expectedRank = generateKeyBetween(
						testData.nonTodayFuturePartition.firstTask.rank,
						testData.nonTodayFuturePartition.secondTask.rank
					);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayPastPartitionLabel}`);
					const destinationTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayFuturePartitionLabel}`);
					const reorderedTasks = [
						testData.nonTodayFuturePartition.firstTask,
						testData.nonTodayPastPartition.firstTask,
						testData.nonTodayFuturePartition.secondTask,
						testData.nonTodayFuturePartition.thirdTask
					];
					await dropTask(destinationTaskList, reorderedTasks, testData.nonTodayPastPartition.firstTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.nonTodayFuturePartitionDate.toString(),
						previousTaskId: testData.nonTodayFuturePartition.firstTask.id
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.nonTodayPastPartition.firstTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(
						within(sourceTaskList).queryByRole('listitem', { name: testData.nonTodayPastPartition.firstTask.title })
					).not.toBeInTheDocument();
				});

				it('moves a task to the last index', async () => {
					const expectedRank = generateKeyBetween(testData.nonTodayFuturePartition.thirdTask.rank, undefined);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayPastPartitionLabel}`);
					const destinationTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayFuturePartitionLabel}`);
					const reorderedTasks = [
						testData.nonTodayFuturePartition.firstTask,
						testData.nonTodayFuturePartition.secondTask,
						testData.nonTodayFuturePartition.thirdTask,
						testData.nonTodayPastPartition.firstTask
					];
					await dropTask(destinationTaskList, reorderedTasks, testData.nonTodayPastPartition.firstTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.nonTodayFuturePartitionDate.toString(),
						previousTaskId: testData.nonTodayFuturePartition.thirdTask.id
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.nonTodayPastPartition.firstTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(
						within(sourceTaskList).queryByRole('listitem', { name: testData.nonTodayPastPartition.firstTask.title })
					).not.toBeInTheDocument();
				});
			});

			describe('from the no-date partition', () => {
				it.each([
					[
						'first',
						testData.nonTodayPastPartitionDate,
						testData.nonTodayPastPartitionLabel,
						[testData.noDatePartition.thirdTask, ...testData.nonTodayPastPartition.tasks],
						undefined,
						testData.nonTodayPastPartition.firstTask.rank
					],
					[
						'middle',
						testData.nonTodayPastPartitionDate,
						testData.nonTodayPastPartitionLabel,
						[
							testData.nonTodayPastPartition.firstTask,
							testData.noDatePartition.thirdTask,
							testData.nonTodayPastPartition.secondTask,
							testData.nonTodayPastPartition.thirdTask
						],
						testData.nonTodayPastPartition.firstTask.rank,
						testData.nonTodayPastPartition.secondTask.rank
					],
					[
						'last',
						testData.nonTodayPastPartitionDate,
						testData.nonTodayPastPartitionLabel,
						[...testData.nonTodayPastPartition.tasks, testData.noDatePartition.thirdTask],
						testData.nonTodayPastPartition.thirdTask.rank,
						undefined
					]
				])(
					'moves a task to the %s index of the non-today-date partition',
					async (_, destinationDate, destinationDateLabel, reorderedTasks, previousRank, nextRank) => {
						const response = getResponse({ rank: generateKeyBetween(previousRank, nextRank) });
						fetchMock.mockResolvedValue(response);

						const page = render(Page);
						const sourceTaskList = page.getByLabelText('Tasks without a date');
						const destinationTaskList = page.getByLabelText(`Tasks dated ${destinationDateLabel}`);
						await dropTask(destinationTaskList, reorderedTasks, testData.noDatePartition.thirdTask.id);

						await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
						const draggedTaskIndex = reorderedTasks.findIndex((task) => task.id === testData.noDatePartition.thirdTask.id);
						const requestBody = JSON.stringify({
							date: destinationDate.toString(),
							previousTaskId: reorderedTasks[draggedTaskIndex - 1]?.id
						});
						expectRequestUrl(fetchMock, `/api/tasks/${testData.noDatePartition.thirdTask.id}/drag-and-drop`);
						expectRequestMethod(fetchMock, 'PATCH');
						expectRequestCredentialsMode(fetchMock, 'same-origin');
						expectRequestBody(fetchMock, requestBody);
						expectRequestContentType(fetchMock, 'application/json');
						expectTaskOrder(destinationTaskList, reorderedTasks);
						expect(
							within(sourceTaskList).queryByRole('listitem', { name: testData.noDatePartition.thirdTask.title })
						).not.toBeInTheDocument();
					}
				);

				it.each([
					[
						'first',
						[testData.noDatePartition.thirdTask, ...testData.todayPartition.tasks],
						undefined,
						testData.todayPartition.firstTask.rank
					],
					[
						'middle',
						[
							testData.todayPartition.firstTask,
							testData.noDatePartition.thirdTask,
							testData.todayPartition.secondTask,
							testData.todayPartition.thirdTask
						],
						testData.todayPartition.firstTask.rank,
						testData.todayPartition.secondTask.rank
					],
					[
						'last',
						[...testData.todayPartition.tasks, testData.noDatePartition.thirdTask],
						testData.todayPartition.thirdTask.rank,
						undefined
					]
				])('moves a task to the %s index of the today partition', async (_, reorderedTasks, previousRank, nextRank) => {
					const response = getResponse({ rank: generateKeyBetween(previousRank, nextRank) });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText('Tasks without a date');
					const destinationTaskList = page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`);
					await dropTask(destinationTaskList, reorderedTasks, testData.noDatePartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const draggedTaskIndex = reorderedTasks.findIndex((task) => task.id === testData.noDatePartition.thirdTask.id);
					const requestBody = JSON.stringify({
						date: testData.todayPartitionDate.toString(),
						previousTaskId: reorderedTasks[draggedTaskIndex - 1]?.id
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.noDatePartition.thirdTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(
						within(sourceTaskList).queryByRole('listitem', { name: testData.noDatePartition.thirdTask.title })
					).not.toBeInTheDocument();
				});
			});

			describe('from the non-today-date partition', () => {
				it.each([
					[
						'first',
						[testData.nonTodayPastPartition.thirdTask, ...testData.noDatePartition.tasks],
						undefined,
						testData.noDatePartition.firstTask.rank
					],
					[
						'middle',
						[
							testData.noDatePartition.firstTask,
							testData.nonTodayPastPartition.thirdTask,
							testData.noDatePartition.secondTask,
							testData.noDatePartition.thirdTask
						],
						testData.noDatePartition.firstTask.rank,
						testData.noDatePartition.secondTask.rank
					],
					[
						'last',
						[...testData.noDatePartition.tasks, testData.nonTodayPastPartition.thirdTask],
						testData.noDatePartition.thirdTask.rank,
						undefined
					]
				])('moves a task to the %s index of the no-date partition', async (_, reorderedTasks, previousRank, nextRank) => {
					const response = getResponse({ rank: generateKeyBetween(previousRank, nextRank) });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayPastPartitionLabel}`);
					const destinationTaskList = page.getByLabelText('Tasks without a date');
					await dropTask(destinationTaskList, reorderedTasks, testData.nonTodayPastPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const draggedTaskIndex = reorderedTasks.findIndex((task) => task.id === testData.nonTodayPastPartition.thirdTask.id);
					const requestBody = JSON.stringify({ previousTaskId: reorderedTasks[draggedTaskIndex - 1]?.id });
					expectRequestUrl(fetchMock, `/api/tasks/${testData.nonTodayPastPartition.thirdTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(
						within(sourceTaskList).queryByRole('listitem', { name: testData.nonTodayPastPartition.thirdTask.title })
					).not.toBeInTheDocument();
				});

				it.each([
					[
						'first',
						[testData.nonTodayPastPartition.thirdTask, ...testData.todayPartition.tasks],
						undefined,
						testData.todayPartition.firstTask.rank
					],
					[
						'middle',
						[
							testData.todayPartition.firstTask,
							testData.nonTodayPastPartition.thirdTask,
							testData.todayPartition.secondTask,
							testData.todayPartition.thirdTask
						],
						testData.todayPartition.firstTask.rank,
						testData.todayPartition.secondTask.rank
					],
					[
						'last',
						[...testData.todayPartition.tasks, testData.nonTodayPastPartition.thirdTask],
						testData.todayPartition.thirdTask.rank,
						undefined
					]
				])('moves a task to the %s index of the today partition', async (_, reorderedTasks, previousRank, nextRank) => {
					const response = getResponse({ rank: generateKeyBetween(previousRank, nextRank) });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayPastPartitionLabel}`);
					const destinationTaskList = page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`);
					await dropTask(destinationTaskList, reorderedTasks, testData.nonTodayPastPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const draggedTaskIndex = reorderedTasks.findIndex((task) => task.id === testData.nonTodayPastPartition.thirdTask.id);
					const requestBody = JSON.stringify({
						date: testData.todayPartitionDate.toString(),
						previousTaskId: reorderedTasks[draggedTaskIndex - 1]?.id
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.nonTodayPastPartition.thirdTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(
						within(sourceTaskList).queryByRole('listitem', { name: testData.nonTodayPastPartition.thirdTask.title })
					).not.toBeInTheDocument();
				});
			});

			describe('from the today partition', () => {
				it.each([
					[
						'first',
						[testData.todayPartition.thirdTask, ...testData.noDatePartition.tasks],
						undefined,
						testData.noDatePartition.firstTask.rank
					],
					[
						'middle',
						[
							testData.noDatePartition.firstTask,
							testData.todayPartition.thirdTask,
							testData.noDatePartition.secondTask,
							testData.noDatePartition.thirdTask
						],
						testData.noDatePartition.firstTask.rank,
						testData.noDatePartition.secondTask.rank
					],
					[
						'last',
						[...testData.noDatePartition.tasks, testData.todayPartition.thirdTask],
						testData.noDatePartition.thirdTask.rank,
						undefined
					]
				])('moves a task to the %s index of the no-date partition', async (_, reorderedTasks, previousRank, nextRank) => {
					const response = getResponse({ rank: generateKeyBetween(previousRank, nextRank) });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`);
					const destinationTaskList = page.getByLabelText('Tasks without a date');
					await dropTask(destinationTaskList, reorderedTasks, testData.todayPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const draggedTaskIndex = reorderedTasks.findIndex((task) => task.id === testData.todayPartition.thirdTask.id);
					const requestBody = JSON.stringify({ previousTaskId: reorderedTasks[draggedTaskIndex - 1]?.id });
					expectRequestUrl(fetchMock, `/api/tasks/${testData.todayPartition.thirdTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(
						within(sourceTaskList).queryByRole('listitem', { name: testData.todayPartition.thirdTask.title })
					).not.toBeInTheDocument();
				});

				it.each([
					[
						'first',
						[testData.todayPartition.thirdTask, ...testData.nonTodayPastPartition.tasks],
						undefined,
						testData.nonTodayPastPartition.firstTask.rank
					],
					[
						'middle',
						[
							testData.nonTodayPastPartition.firstTask,
							testData.todayPartition.thirdTask,
							testData.nonTodayPastPartition.secondTask,
							testData.nonTodayPastPartition.thirdTask
						],
						testData.nonTodayPastPartition.firstTask.rank,
						testData.nonTodayPastPartition.secondTask.rank
					],
					[
						'last',
						[...testData.nonTodayPastPartition.tasks, testData.todayPartition.thirdTask],
						testData.nonTodayPastPartition.thirdTask.rank,
						undefined
					]
				])('moves a task to the %s index of the non-today-date partition', async (_, reorderedTasks, previousRank, nextRank) => {
					const response = getResponse({ rank: generateKeyBetween(previousRank, nextRank) });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`);
					const destinationTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayPastPartitionLabel}`);
					await dropTask(destinationTaskList, reorderedTasks, testData.todayPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const draggedTaskIndex = reorderedTasks.findIndex((task) => task.id === testData.todayPartition.thirdTask.id);
					const requestBody = JSON.stringify({
						date: testData.nonTodayPastPartitionDate.toString(),
						previousTaskId: reorderedTasks[draggedTaskIndex - 1]?.id
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.todayPartition.thirdTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(
						within(sourceTaskList).queryByRole('listitem', { name: testData.todayPartition.thirdTask.title })
					).not.toBeInTheDocument();
				});
			});
		});

		describe('infinite scroll', () => {
			it('does not enable infinite scrolling for exhausted partitions', () => {
				const page = render(Page);
				const noDateColumn = getTaskColumn(page, 'No Date Assigned');
				const todayColumn = getTaskColumn(page, 'Today');
				const nonTodayColumn = getTaskColumn(page, 'Date Assigned');

				expect(noDateColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(todayColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(nonTodayColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});
		});
	});

	describe('has more tasks scenario', () => {
		beforeEach(() => {
			undatedTaskStore.addPage(testData.noDatePartition.tasks, true);
			datedUntrashedOpenTaskStore.addTodayPage(testData.todayPartition.tasks, true);
			datedUntrashedOpenTaskStore.addNonTodayPage(
				[...testData.nonTodayPastPartition.tasks, ...testData.nonTodayFuturePartition.tasks],
				true
			);
		});

		describe('infinite scroll', () => {
			it('loads more tasks without a date', async () => {
				const nextTaskResponse = {
					id: 13,
					categoryId: category.id,
					title: 'Buy train tickets',
					rank: generateKeyBetween(testData.noDatePartition.thirdTask.rank, undefined)
				};
				const nextTask: UndatedTaskResponse = { ...nextTaskResponse, done: false };
				const existingDatedTasks = [...datedUntrashedOpenTaskStore.datedUntrashedOpenTasks];
				const resolveFetch = mockPendingFetch(fetchMock);

				const page = render(Page);
				const noDateColumn = getTaskColumn(page, 'No Date Assigned');
				const loadingSentinel = getLoadingSentinel(noDateColumn);
				intersectionObserverMock.enterNode(loadingSentinel);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, '/api/tasks/undated');
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expect(getRequestSearchParameter(fetchMock, 'rank')).toBe(testData.noDatePartition.thirdTask.rank);
				expect(within(noDateColumn).getByLabelText('Loading more items')).toBeInTheDocument();
				expect(within(getTaskColumn(page, 'Today')).queryByLabelText('Loading more items')).not.toBeInTheDocument();
				expect(within(getTaskColumn(page, 'Date Assigned')).queryByLabelText('Loading more items')).not.toBeInTheDocument();

				resolveFetch(getResponse({ tasks: [nextTaskResponse], hasMore: false }));

				await waitFor(() => expect(undatedTaskStore.undatedTasks).toContainEqual(nextTask));
				expect(undatedTaskStore.hasMore).toBe(false);
				expectTaskToBePresent(noDateColumn, nextTask.title, nextTask.description, category.title);
				expect(noDateColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(datedUntrashedOpenTaskStore.datedUntrashedOpenTasks).toEqual(existingDatedTasks);
			});

			it('loads more tasks dated today', async () => {
				const nextTaskResponse = {
					id: 13,
					categoryId: category.id,
					title: 'Pick up the parcel',
					date: testData.todayPartitionDate.toString(),
					rank: generateKeyBetween(testData.todayPartition.thirdTask.rank, undefined)
				};
				const nextTask: DatedUntrashedOpenTaskResponse = {
					...nextTaskResponse,
					date: testData.todayPartitionDate,
					done: false
				};
				const existingNonTodayTasks = [...datedUntrashedOpenTaskStore.datedUntrashedOpenNonTodayTasks];
				const resolveFetch = mockPendingFetch(fetchMock);

				const page = render(Page);
				const todayColumn = getTaskColumn(page, 'Today');
				const loadingSentinel = getLoadingSentinel(todayColumn);
				intersectionObserverMock.enterNode(loadingSentinel);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, '/api/tasks/dated-untrashed-open-today');
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expect(getRequestSearchParameter(fetchMock, 'rank')).toBe(testData.todayPartition.thirdTask.rank);
				expect(within(todayColumn).getByLabelText('Loading more items')).toBeInTheDocument();
				expect(within(getTaskColumn(page, 'Date Assigned')).queryByLabelText('Loading more items')).not.toBeInTheDocument();

				resolveFetch(getResponse({ tasks: [nextTaskResponse], hasMore: false }));

				await waitFor(() => expect(datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks).toContainEqual(nextTask));
				expect(datedUntrashedOpenTaskStore.todayTasksHasMore).toBe(false);
				expectTaskToBePresent(todayColumn, nextTask.title, nextTask.description, category.title);
				expect(todayColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(datedUntrashedOpenTaskStore.datedUntrashedOpenNonTodayTasks).toEqual(existingNonTodayTasks);
			});

			it('loads more tasks with a non-today date', async () => {
				const nextTaskResponse = {
					id: 13,
					categoryId: category.id,
					title: 'Check in for the train',
					date: testData.nonTodayFuturePartitionDate.toString(),
					rank: generateKeyBetween(testData.nonTodayFuturePartition.thirdTask.rank, undefined)
				};
				const nextTask: DatedUntrashedOpenTaskResponse = {
					...nextTaskResponse,
					date: testData.nonTodayFuturePartitionDate,
					done: false
				};
				const existingTodayTasks = [...datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks];
				const resolveFetch = mockPendingFetch(fetchMock);

				const page = render(Page);
				const nonTodayColumn = getTaskColumn(page, 'Date Assigned');
				const loadingSentinel = getLoadingSentinel(nonTodayColumn);
				intersectionObserverMock.enterNode(loadingSentinel);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, '/api/tasks/dated-untrashed-open-non-today');
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expect(getRequestSearchParameter(fetchMock, 'date')).toBe(testData.nonTodayFuturePartitionDate.toString());
				expect(getRequestSearchParameter(fetchMock, 'rank')).toBe(testData.nonTodayFuturePartition.thirdTask.rank);
				expect(within(nonTodayColumn).getByLabelText('Loading more items')).toBeInTheDocument();
				expect(within(getTaskColumn(page, 'Today')).queryByLabelText('Loading more items')).not.toBeInTheDocument();

				resolveFetch(getResponse({ tasks: [nextTaskResponse], hasMore: false }));

				await waitFor(() => expect(datedUntrashedOpenTaskStore.datedUntrashedOpenNonTodayTasks).toContainEqual(nextTask));
				expect(datedUntrashedOpenTaskStore.nonTodayTasksHasMore).toBe(false);
				const taskList = page.getByLabelText(`Tasks dated ${testData.nonTodayFuturePartitionLabel}`);
				expectTaskToBePresent(taskList, nextTask.title, nextTask.description, category.title);
				expect(nonTodayColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(datedUntrashedOpenTaskStore.datedUntrashedOpenTodayTasks).toEqual(existingTodayTasks);
			});

			it('disables infinite scrolling while a task is being dragged', async () => {
				const page = render(Page);
				const noDateColumn = getTaskColumn(page, 'No Date Assigned');
				const todayColumn = getTaskColumn(page, 'Today');
				const nonTodayColumn = getTaskColumn(page, 'Date Assigned');
				const noDateLoadingSentinel = getLoadingSentinel(noDateColumn);
				const todayLoadingSentinel = getLoadingSentinel(todayColumn);
				const nonTodayLoadingSentinel = getLoadingSentinel(nonTodayColumn);

				const todayTaskList = page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`);
				const event = new CustomEvent<DndEvent<DatedUntrashedOpenTaskResponse>>('consider', {
					detail: {
						items: testData.todayPartition.tasks,
						info: {
							trigger: TRIGGERS.DRAG_STARTED,
							id: String(testData.todayPartition.firstTask.id),
							source: SOURCES.POINTER
						}
					}
				});
				await fireEvent(todayTaskList, event);

				intersectionObserverMock.enterNodes([noDateLoadingSentinel, todayLoadingSentinel, nonTodayLoadingSentinel]);
				expect(fetchMock).not.toHaveBeenCalled();
			});
		});
	});
});

const fetchMock = vi.fn();
const toastErrorMock = vi.mocked(toast.error);
const category = { id: 7, title: 'Personal', color: '#22c55e' };
const localTimeZone = getLocalTimeZone();
const maximumTrashTimestampDifferenceMilliseconds = 5_000;
const todayDate = today(localTimeZone);
const testData = createTestData();

type TaskPartition<Task> = {
	firstTask: Task;
	secondTask: Task;
	thirdTask: Task;
	tasks: Task[];
};

function createTestData() {
	const nonTodayPastDate = todayDate.subtract({ days: 1 });
	const nonTodayFutureDate = todayDate.add({ days: 1 });
	let taskId = 1;

	const noDateTaskDetails = [
		{ title: 'Plan weekend hiking route', description: 'Compare trails near Munich and choose one suitable for Saturday.' },
		{ title: 'Organize the photo archive' },
		{ title: 'Call the bicycle repair shop', description: 'Ask when the replacement brake pads will arrive.' }
	] as const;
	const noDateTaskRanks = generateNKeysBetween(undefined, undefined, noDateTaskDetails.length);
	const noDateTasks: UndatedTaskResponse[] = noDateTaskDetails.map((details, index) => ({
		...details,
		id: taskId++,
		categoryId: category.id,
		rank: noDateTaskRanks[index],
		done: false
	}));

	const todayTaskDetails = [
		{ title: 'Buy groceries' },
		{ title: 'Submit the expense report', description: 'Attach the train and hotel receipts before submitting.' },
		{ title: 'Water the balcony plants', description: 'Check the herbs and tomatoes before watering.' }
	] as const;
	const todayTaskRanks = generateNKeysBetween(undefined, undefined, todayTaskDetails.length);
	const todayTasks: DatedUntrashedOpenTaskResponse[] = todayTaskDetails.map((details, index) => ({
		...details,
		id: taskId++,
		categoryId: category.id,
		date: todayDate,
		rank: todayTaskRanks[index],
		done: false
	}));

	const nonTodayPastTaskDetails = [
		{ title: 'Renew the library books', description: 'Extend the loans that are due this week.' },
		{ title: 'Reply to the dentist', description: 'Confirm the proposed appointment for next month.' },
		{ title: 'Back up the laptop' }
	] as const;
	const nonTodayPastTaskRanks = generateNKeysBetween(undefined, undefined, nonTodayPastTaskDetails.length);
	const nonTodayPastTasks: DatedUntrashedOpenTaskResponse[] = nonTodayPastTaskDetails.map((details, index) => ({
		...details,
		id: taskId++,
		categoryId: category.id,
		date: nonTodayPastDate,
		rank: nonTodayPastTaskRanks[index],
		done: false
	}));

	const nonTodayFutureTaskDetails = [
		{ title: 'Pack the conference bag', description: 'Include the charger, notebook, and name badge.' },
		{ title: 'Confirm the train seat', description: 'Check the carriage and seat number in the booking.' },
		{ title: 'Download the city map', description: 'Save an offline map before leaving home.' }
	] as const;
	const nonTodayFutureTaskRanks = generateNKeysBetween(undefined, undefined, nonTodayFutureTaskDetails.length);
	const nonTodayFutureTasks: DatedUntrashedOpenTaskResponse[] = nonTodayFutureTaskDetails.map((details, index) => ({
		...details,
		id: taskId++,
		categoryId: category.id,
		date: nonTodayFutureDate,
		rank: nonTodayFutureTaskRanks[index],
		done: false
	}));

	const noDatePartition: TaskPartition<UndatedTaskResponse> = {
		firstTask: noDateTasks[0]!,
		secondTask: noDateTasks[1]!,
		thirdTask: noDateTasks[2]!,
		tasks: noDateTasks
	};
	const todayPartition: TaskPartition<DatedUntrashedOpenTaskResponse> = {
		firstTask: todayTasks[0]!,
		secondTask: todayTasks[1]!,
		thirdTask: todayTasks[2]!,
		tasks: todayTasks
	};
	const nonTodayPastPartition: TaskPartition<DatedUntrashedOpenTaskResponse> = {
		firstTask: nonTodayPastTasks[0]!,
		secondTask: nonTodayPastTasks[1]!,
		thirdTask: nonTodayPastTasks[2]!,
		tasks: nonTodayPastTasks
	};
	const nonTodayFuturePartition: TaskPartition<DatedUntrashedOpenTaskResponse> = {
		firstTask: nonTodayFutureTasks[0]!,
		secondTask: nonTodayFutureTasks[1]!,
		thirdTask: nonTodayFutureTasks[2]!,
		tasks: nonTodayFutureTasks
	};

	const nonTodayPastPartitionLabel = `${String(nonTodayPastDate.day).padStart(2, '0')}.${String(nonTodayPastDate.month).padStart(2, '0')}.${nonTodayPastDate.year}`;
	const nonTodayFuturePartitionLabel = `${String(nonTodayFutureDate.day).padStart(2, '0')}.${String(nonTodayFutureDate.month).padStart(2, '0')}.${nonTodayFutureDate.year}`;

	return {
		noDatePartition,
		todayPartition,
		todayPartitionDate: todayDate,
		todayPartitionLabel: 'Today',
		nonTodayPastPartition,
		nonTodayPastPartitionDate: nonTodayPastDate,
		nonTodayPastPartitionLabel,
		nonTodayFuturePartition,
		nonTodayFuturePartitionDate: nonTodayFutureDate,
		nonTodayFuturePartitionLabel
	};
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

async function startCategoryCreation(page: RenderResult<typeof Page>): Promise<void> {
	await openCategorySelect(page);
	const addCategoryOption = await page.findByRole('option', { name: 'Add category' });
	await fireEvent.pointerUp(addCategoryOption);
}

async function openCategorySelect(page: RenderResult<typeof Page>): Promise<void> {
	const categorySelect = getSelectById(page, 'task-category');
	categorySelect.focus();
	await fireEvent.keyDown(categorySelect, { key: 'ArrowDown' });
}

function addTasksExceedingMaximumBatchSize(): string[] {
	const taskCount = 51;
	const firstTaskId = 13;
	const ranks = generateNKeysBetween(undefined, undefined, taskCount);
	const tasks = ranks.map((rank, index) => ({
		id: firstTaskId + index,
		categoryId: category.id,
		title: `Review project note ${index + 1}`,
		description: `Review and organize project note ${index + 1}.`,
		rank,
		done: false as const
	}));
	undatedTaskStore.addOrUpdateTasks(tasks);
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

async function dropTask(
	taskList: HTMLElement,
	newOrderedTasks: (UndatedTaskResponse | DatedUntrashedOpenTaskResponse)[],
	draggedTaskId: number
): Promise<void> {
	const event = new CustomEvent<DndEvent<UndatedTaskResponse | DatedUntrashedOpenTaskResponse>>('finalize', {
		detail: {
			items: newOrderedTasks,
			info: { trigger: TRIGGERS.DROPPED_INTO_ZONE, id: String(draggedTaskId), source: SOURCES.POINTER }
		}
	});
	await fireEvent(taskList, event);
}

function expectTaskOrder(taskList: HTMLElement, tasks: (UndatedTaskResponse | DatedUntrashedOpenTaskResponse)[]): void {
	const taskListContent = within(taskList);
	const renderedTasks = taskListContent.getAllByRole('listitem');
	const renderedTaskTitles = renderedTasks.map((task) => task.getAttribute('aria-label'));
	const expectedTaskTitles = tasks.map((task) => task.title);
	expect(renderedTaskTitles).toEqual(expectedTaskTitles);
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
	const taskCategory = taskContent.getByText(categoryName);
	expect(task).toBeInTheDocument();
	if (description !== undefined) expect(taskContent.getByText(description)).toBeInTheDocument();
	expect(taskCategory).toBeInTheDocument();
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

function getLoadingSentinel(taskColumn: HTMLElement): HTMLElement {
	const loadingSentinel = taskColumn.querySelector('.loading-sentinel');
	if (!(loadingSentinel instanceof HTMLElement)) throw new Error('Expected the task column to have a loading sentinel.');
	return loadingSentinel;
}
