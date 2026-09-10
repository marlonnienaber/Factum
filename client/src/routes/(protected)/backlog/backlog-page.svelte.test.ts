import {fireEvent, render, type RenderResult, waitFor, within} from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import {type CalendarDate, getLocalTimeZone, now, parseAbsolute, today} from '@internationalized/date';
import {generateKeyBetween, generateNKeysBetween} from 'fractional-indexing';
import {toast} from 'svelte-sonner';
import {type DndEvent, SOURCES, TRIGGERS} from 'svelte-dnd-action';
import {categoryStore} from '$lib/stores/category-store.svelte';
import {datedUntrashedDoneTaskStore} from '$lib/stores/dated-untrashed-done-task-store.svelte';
import {datedUntrashedOpenTaskStore} from '$lib/stores/dated-untrashed-open-task-store.svelte';
import {trashedTaskStore} from '$lib/stores/trashed-task-store.svelte';
import {undatedTaskStore} from '$lib/stores/undated-task-store.svelte';
import type {DatedUntrashedOpenTaskResponse} from '$lib/types/responses/dated-untrashed-open-task-response';
import type {UndatedTaskResponse} from '$lib/types/responses/undated-task-response';
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
	getTextareaById
} from '$lib/utils/testing';
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest';
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
	});

	afterEach(async () => {
		if (document.querySelector('[role="dialog"]')) await userEvent.keyboard('{Escape}');
		document.body.removeAttribute('style');
	});

	describe('category operations', () => {
		it('successfully creates and deletes a category', async () => {
			const createCategoryResponse = getResponse({ id: category.id }, 201);
			fetchMock.mockResolvedValue(createCategoryResponse);

			const page = render(Page);
			await openCreateTaskDialog(page);
			await startCategoryCreation(page);

			const categoryNameInput = page.getByRole('textbox', { name: 'Category name' });
			await fireEvent.input(categoryNameInput, { target: { value: category.title } });
			const categoryColorInput = page.getByLabelText('Choose category color');
			await fireEvent.input(categoryColorInput, { target: { value: category.color } });
			const createCategoryButton = page.getByRole('button', { name: 'Create category' });
			await userEvent.click(createCategoryButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const createRequestBody = JSON.stringify({ title: category.title, color: category.color });
			expectRequestUrl(fetchMock, '/api/categories');
			expectRequestMethod(fetchMock, 'POST');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock, createRequestBody);
			expectRequestContentType(fetchMock, 'application/json');
			const categorySelect = getSelectById(page, 'task-category');
			expect(categorySelect).toHaveTextContent(category.title);

			fetchMock.mockClear();
			const deleteCategoryResponse = new Response(null, { status: 204 });
			fetchMock.mockResolvedValue(deleteCategoryResponse);

			await openCategorySelect(page);
			const deleteCategoryButton = await page.findByRole('button', { name: `Delete ${category.title}` });
			await userEvent.click(deleteCategoryButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			expectRequestUrl(fetchMock, `/api/categories/${category.id}`);
			expectRequestMethod(fetchMock, 'DELETE');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock, undefined);
			await waitFor(() => expect(categorySelect).toHaveTextContent('None chosen'));
			expect(categoryStore.categories).not.toContainEqual(category);
		});

		it('shows an error toast when deleting a referenced category', async () => {
			categoryStore.setCategories([category]);
			undatedTaskStore.addTask(testData.noDatePartition.firstTask);

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
			categoryStore.setCategories([category]);

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
			categoryStore.setCategories([category]);

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
			categoryStore.setCategories([category]);

			const page = render(Page);
			await openCreateTaskDialog(page);
			await selectCategory(page, category.title);
			await enterTaskTitle(page, testData.noDatePartition.firstTask.title);
			const tooLongDescription = 'Compare the available trails, travel times, and weather forecast. '.repeat(9);
			const descriptionInput = await enterTaskDescription(page, tooLongDescription);
			await fireEvent.blur(descriptionInput);

			const createButton = page.getByRole('button', { name: 'Create' });
			expect(createButton).toBeDisabled();
			await expectTextareaToBeInvalid(page, 'task-description', 'Title must have at most 500 characters');
		});

		it('reacts correctly to missing task category', async () => {
			const page = render(Page);
			await openCreateTaskDialog(page);
			await enterTaskTitle(page, testData.noDatePartition.firstTask.title);
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
			categoryStore.setCategories([category]);

			const response = getResponse({ id: 42, rank: 'a0' }, 201);
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			await openCreateTaskDialog(page);
			await enterTaskTitle(page, testData.noDatePartition.firstTask.title);
			await enterTaskDescription(page, testData.noDatePartition.firstTask.description);
			await selectCategory(page, category.title);
			const createButton = page.getByRole('button', { name: 'Create' });
			await userEvent.click(createButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const createTaskDialog = page.queryByRole('dialog', { name: 'Create Task' });
			expect(createTaskDialog).not.toBeInTheDocument();
			const requestBody = JSON.stringify({
				title: testData.noDatePartition.firstTask.title,
				description: testData.noDatePartition.firstTask.description,
				categoryId: category.id
			});
			expectRequestUrl(fetchMock, '/api/tasks');
			expectRequestMethod(fetchMock, 'POST');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock, requestBody);
			expectRequestContentType(fetchMock, 'application/json');
			const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
			expectTaskToBePresent(
				tasksWithoutDateList,
				testData.noDatePartition.firstTask.title,
				testData.noDatePartition.firstTask.description,
				category.title
			);
		});
	});

	describe('task editing', () => {
		it('reacts correctly to missing task title', async () => {
			const task = testData.noDatePartition.firstTask;
			categoryStore.setCategories([category]);
			undatedTaskStore.addTask(task);

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
			categoryStore.setCategories([category]);
			undatedTaskStore.addTask(task);

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
			categoryStore.setCategories([category]);
			undatedTaskStore.addTask(task);

			const page = render(Page);
			await openEditTaskDialog(page, task.title);
			const tooLongDescription = 'Compare the available trails, travel times, and weather forecast. '.repeat(9);
			const descriptionInput = await enterTaskDescription(page, tooLongDescription);
			await fireEvent.blur(descriptionInput);

			const saveButton = page.getByRole('button', { name: 'Save' });
			expect(saveButton).toBeDisabled();
			await expectTextareaToBeInvalid(page, 'task-description', 'Title must have at most 500 characters');
		});

		it('cancels and closes the edit-task dialog', async () => {
			const task = testData.noDatePartition.firstTask;
			categoryStore.setCategories([category]);
			undatedTaskStore.addTask(task);

			const page = render(Page);
			await openEditTaskDialog(page, task.title);
			const cancelButton = page.getByRole('button', { name: 'Cancel' });
			await userEvent.click(cancelButton);

			const editTaskDialog = page.queryByRole('dialog', { name: 'Edit Task' });
			expect(editTaskDialog).not.toBeInTheDocument();
		});

		it('successfully edits a task without a date', async () => {
			const task = testData.noDatePartition.firstTask;
			categoryStore.setCategories([category]);
			undatedTaskStore.addTask(task);

			const newTitle = 'Book the hiking trail transfer';
			const newDescription = 'Reserve the earliest train and the return connection.';

			const response = getResponse({
				id: task.id,
				categoryId: category.id,
				title: newTitle,
				description: newDescription,
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
			categoryStore.setCategories([category]);
			datedUntrashedOpenTaskStore.addTask(task);

			const newTitle = 'Book the hiking trail transfer';
			const newDescription = 'Reserve the earliest train and the return connection.';

			const response = getResponse({
				id: task.id,
				categoryId: category.id,
				title: newTitle,
				description: newDescription,
				date: taskDate.toString(),
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
			categoryStore.setCategories([category]);
			datedUntrashedOpenTaskStore.addTask(task);

			const newTitle = 'Book the hiking trail transfer';
			const newDescription = 'Reserve the earliest train and the return connection.';

			const response = getResponse({
				id: task.id,
				categoryId: category.id,
				title: newTitle,
				description: newDescription,
				date: taskDate.toString(),
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
			categoryStore.setCategories([category]);
			undatedTaskStore.addTask(task);

			const response = new Response(null, { status: 204 });
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			const completeTaskButton = page.getByRole('checkbox', {
				name: `Mark "${task.title}" as completed`
			});
			await userEvent.click(completeTaskButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const expectedRank = generateKeyBetween(undefined, undefined);
			const requestUrl = `/api/tasks/${task.id}/toggle-done?rank=${expectedRank}`;
			expectRequestUrl(fetchMock, requestUrl);
			expectRequestMethod(fetchMock, 'PATCH');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock, undefined);
			expectTaskNotToBePresent(page, task.title);
		});

		it('successfully completes a task with a non-today date', async () => {
			const task = testData.nonTodayPastPartition.firstTask;
			categoryStore.setCategories([category]);
			datedUntrashedOpenTaskStore.addTask(task);

			const response = new Response(null, { status: 204 });
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			const completeTaskButton = page.getByRole('checkbox', {
				name: `Mark "${task.title}" as completed`
			});
			await userEvent.click(completeTaskButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const expectedRank = generateKeyBetween(undefined, undefined);
			const requestUrl = `/api/tasks/${task.id}/toggle-done?rank=${expectedRank}`;
			expectRequestUrl(fetchMock, requestUrl);
			expectRequestMethod(fetchMock, 'PATCH');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock, undefined);
			expectTaskNotToBePresent(page, task.title);
		});

		it('successfully completes a task dated today', async () => {
			const task = testData.todayPartition.firstTask;
			categoryStore.setCategories([category]);
			datedUntrashedOpenTaskStore.addTask(task);

			const response = new Response(null, { status: 204 });
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			const completeTaskButton = page.getByRole('checkbox', {
				name: `Mark "${task.title}" as completed`
			});
			await userEvent.click(completeTaskButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const expectedRank = generateKeyBetween(undefined, undefined);
			const requestUrl = `/api/tasks/${task.id}/toggle-done?rank=${expectedRank}`;
			expectRequestUrl(fetchMock, requestUrl);
			expectRequestMethod(fetchMock, 'PATCH');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock, undefined);
			expectTaskNotToBePresent(page, task.title);
		});
	});

	describe('task trashing', () => {
		it('successfully trashes a task without a date', async () => {
			const task = testData.noDatePartition.firstTask;
			categoryStore.setCategories([category]);
			undatedTaskStore.addTask(task);

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
			categoryStore.setCategories([category]);
			datedUntrashedOpenTaskStore.addTask(task);

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
			categoryStore.setCategories([category]);
			datedUntrashedOpenTaskStore.addTask(task);

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
			categoryStore.setCategories([category]);
			undatedTaskStore.addTask(tasks[0]);
			datedUntrashedOpenTaskStore.addTask(tasks[1]);
			datedUntrashedOpenTaskStore.addTask(tasks[2]);
			const taskIds = tasks.map((task) => task.id);
			const completedTaskRanks = generateNKeysBetween(undefined, undefined, tasks.length);
			const completedTasks = [
				{ ...tasks[0], date: testData.todayPartitionDate.toString(), rank: completedTaskRanks[0] },
				{ ...tasks[1], date: testData.nonTodayPastPartitionDate.toString(), rank: completedTaskRanks[1] },
				{ ...tasks[2], date: testData.todayPartitionDate.toString(), rank: completedTaskRanks[2] }
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
			categoryStore.setCategories([category]);
			undatedTaskStore.addTask(tasks[0]);
			datedUntrashedOpenTaskStore.addTask(tasks[1]);
			datedUntrashedOpenTaskStore.addTask(tasks[2]);
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
			categoryStore.setCategories([category]);
			undatedTaskStore.addTask(task);

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
			const task = testData.noDatePartition.firstTask;
			categoryStore.setCategories([category]);
			undatedTaskStore.addTask(task);

			const page = render(Page);
			await clickOnSelectTasksButton(page);

			const completeAllButton = page.getByRole('button', { name: 'Complete All' });
			const trashAllButton = page.getByRole('button', { name: 'Trash All' });
			expect(completeAllButton).toBeDisabled();
			expect(trashAllButton).toBeDisabled();
			expect(page.getByText('0 / 50')).toBeInTheDocument();
		});

		it('disables batch trash when more than 50 tasks are selected', async () => {
			categoryStore.setCategories([category]);
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
			categoryStore.setCategories([category]);
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
				categoryStore.setCategories([category]);
				undatedTaskStore.addTasks(testData.noDatePartition.tasks);

				const response = new Response(null, { status: 204 });
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
				const expectedRank = generateKeyBetween(undefined, testData.noDatePartition.firstTask.rank);
				const requestBody = JSON.stringify({ id: testData.noDatePartition.thirdTask.id, rank: expectedRank });
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(tasksWithoutDateList, reorderedTasks);
			});

			it('moves a task to the middle index', async () => {
				categoryStore.setCategories([category]);
				undatedTaskStore.addTasks(testData.noDatePartition.tasks);

				const response = new Response(null, { status: 204 });
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
				const expectedRank = generateKeyBetween(
					testData.noDatePartition.firstTask.rank,
					testData.noDatePartition.secondTask.rank
				);
				const requestBody = JSON.stringify({ id: testData.noDatePartition.thirdTask.id, rank: expectedRank });
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(tasksWithoutDateList, reorderedTasks);
			});

			it('moves a task to the last index', async () => {
				categoryStore.setCategories([category]);
				undatedTaskStore.addTasks(testData.noDatePartition.tasks);

				const response = new Response(null, { status: 204 });
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
				const expectedRank = generateKeyBetween(testData.noDatePartition.thirdTask.rank, undefined);
				const requestBody = JSON.stringify({ id: testData.noDatePartition.firstTask.id, rank: expectedRank });
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(tasksWithoutDateList, reorderedTasks);
			});
		});

		describe('within the today partition', () => {
			it('moves a task to the first index', async () => {
				categoryStore.setCategories([category]);
				datedUntrashedOpenTaskStore.addTasks(testData.todayPartition.tasks);

				const response = new Response(null, { status: 204 });
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
				const expectedRank = generateKeyBetween(undefined, testData.todayPartition.firstTask.rank);
				const requestBody = JSON.stringify({
					id: testData.todayPartition.thirdTask.id,
					date: testData.todayPartitionDate.toString(),
					rank: expectedRank
				});
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(todayTaskList, reorderedTasks);
			});

			it('moves a task to the middle index', async () => {
				categoryStore.setCategories([category]);
				datedUntrashedOpenTaskStore.addTasks(testData.todayPartition.tasks);

				const response = new Response(null, { status: 204 });
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
				const expectedRank = generateKeyBetween(
					testData.todayPartition.firstTask.rank,
					testData.todayPartition.secondTask.rank
				);
				const requestBody = JSON.stringify({
					id: testData.todayPartition.thirdTask.id,
					date: testData.todayPartitionDate.toString(),
					rank: expectedRank
				});
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(todayTaskList, reorderedTasks);
			});

			it('moves a task to the last index', async () => {
				categoryStore.setCategories([category]);
				datedUntrashedOpenTaskStore.addTasks(testData.todayPartition.tasks);

				const response = new Response(null, { status: 204 });
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
				const expectedRank = generateKeyBetween(testData.todayPartition.thirdTask.rank, undefined);
				const requestBody = JSON.stringify({
					id: testData.todayPartition.firstTask.id,
					date: testData.todayPartitionDate.toString(),
					rank: expectedRank
				});
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(todayTaskList, reorderedTasks);
			});
		});

		describe('within the non-today-date partition', () => {
			it('moves a task to the first index', async () => {
				categoryStore.setCategories([category]);
				datedUntrashedOpenTaskStore.addTasks(testData.nonTodayPastPartition.tasks);

				const response = new Response(null, { status: 204 });
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
				const expectedRank = generateKeyBetween(undefined, testData.nonTodayPastPartition.firstTask.rank);
				const requestBody = JSON.stringify({
					id: testData.nonTodayPastPartition.thirdTask.id,
					date: testData.nonTodayPastPartitionDate.toString(),
					rank: expectedRank
				});
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(nonTodayTaskList, reorderedTasks);
			});

			it('moves a task to the middle index', async () => {
				categoryStore.setCategories([category]);
				datedUntrashedOpenTaskStore.addTasks(testData.nonTodayPastPartition.tasks);

				const response = new Response(null, { status: 204 });
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
				const expectedRank = generateKeyBetween(
					testData.nonTodayPastPartition.firstTask.rank,
					testData.nonTodayPastPartition.secondTask.rank
				);
				const requestBody = JSON.stringify({
					id: testData.nonTodayPastPartition.thirdTask.id,
					date: testData.nonTodayPastPartitionDate.toString(),
					rank: expectedRank
				});
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(nonTodayTaskList, reorderedTasks);
			});

			it('moves a task to the last index', async () => {
				categoryStore.setCategories([category]);
				datedUntrashedOpenTaskStore.addTasks(testData.nonTodayPastPartition.tasks);

				const response = new Response(null, { status: 204 });
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
				const expectedRank = generateKeyBetween(testData.nonTodayPastPartition.thirdTask.rank, undefined);
				const requestBody = JSON.stringify({
					id: testData.nonTodayPastPartition.firstTask.id,
					date: testData.nonTodayPastPartitionDate.toString(),
					rank: expectedRank
				});
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(nonTodayTaskList, reorderedTasks);
			});
		});

		describe('between different non-today date groups', () => {
			it('moves a task to the first index', async () => {
				categoryStore.setCategories([category]);
				datedUntrashedOpenTaskStore.addTasks(testData.nonTodayPastPartition.tasks);
				datedUntrashedOpenTaskStore.addTasks(testData.nonTodayFuturePartition.tasks);

				const response = new Response(null, { status: 204 });
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
				const expectedRank = generateKeyBetween(undefined, testData.nonTodayFuturePartition.firstTask.rank);
				const requestBody = JSON.stringify({
					id: testData.nonTodayPastPartition.firstTask.id,
					date: testData.nonTodayFuturePartitionDate.toString(),
					rank: expectedRank
				});
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
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
				categoryStore.setCategories([category]);
				datedUntrashedOpenTaskStore.addTasks(testData.nonTodayPastPartition.tasks);
				datedUntrashedOpenTaskStore.addTasks(testData.nonTodayFuturePartition.tasks);

				const response = new Response(null, { status: 204 });
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
				const expectedRank = generateKeyBetween(
					testData.nonTodayFuturePartition.firstTask.rank,
					testData.nonTodayFuturePartition.secondTask.rank
				);
				const requestBody = JSON.stringify({
					id: testData.nonTodayPastPartition.firstTask.id,
					date: testData.nonTodayFuturePartitionDate.toString(),
					rank: expectedRank
				});
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
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
				categoryStore.setCategories([category]);
				datedUntrashedOpenTaskStore.addTasks(testData.nonTodayPastPartition.tasks);
				datedUntrashedOpenTaskStore.addTasks(testData.nonTodayFuturePartition.tasks);

				const response = new Response(null, { status: 204 });
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
				const expectedRank = generateKeyBetween(testData.nonTodayFuturePartition.thirdTask.rank, undefined);
				const requestBody = JSON.stringify({
					id: testData.nonTodayPastPartition.firstTask.id,
					date: testData.nonTodayFuturePartitionDate.toString(),
					rank: expectedRank
				});
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
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
					testData.nonTodayPastPartition.tasks,
					[testData.noDatePartition.thirdTask, ...testData.nonTodayPastPartition.tasks],
					undefined,
					testData.nonTodayPastPartition.firstTask.rank
				],
				[
					'middle',
					testData.nonTodayPastPartitionDate,
					testData.nonTodayPastPartitionLabel,
					testData.nonTodayPastPartition.tasks,
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
					testData.nonTodayPastPartition.tasks,
					[...testData.nonTodayPastPartition.tasks, testData.noDatePartition.thirdTask],
					testData.nonTodayPastPartition.thirdTask.rank,
					undefined
				]
			])(
				'moves a task to the %s index of the non-today-date partition',
				async (_, destinationDate, destinationDateLabel, destinationTasks, reorderedTasks, previousRank, nextRank) => {
					categoryStore.setCategories([category]);
					undatedTaskStore.addTask(testData.noDatePartition.thirdTask);
					datedUntrashedOpenTaskStore.addTasks(destinationTasks);

					const response = new Response(null, { status: 204 });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText('Tasks without a date');
					const destinationTaskList = page.getByLabelText(`Tasks dated ${destinationDateLabel}`);
					await dropTask(destinationTaskList, reorderedTasks, testData.noDatePartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const expectedRank = generateKeyBetween(previousRank, nextRank);
					const requestBody = JSON.stringify({
						id: testData.noDatePartition.thirdTask.id,
						date: destinationDate.toString(),
						rank: expectedRank
					});
					expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
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
					testData.todayPartition.tasks,
					[testData.noDatePartition.thirdTask, ...testData.todayPartition.tasks],
					undefined,
					testData.todayPartition.firstTask.rank
				],
				[
					'middle',
					testData.todayPartition.tasks,
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
					testData.todayPartition.tasks,
					[...testData.todayPartition.tasks, testData.noDatePartition.thirdTask],
					testData.todayPartition.thirdTask.rank,
					undefined
				]
			])(
				'moves a task to the %s index of the today partition',
				async (_, destinationTasks, reorderedTasks, previousRank, nextRank) => {
					categoryStore.setCategories([category]);
					undatedTaskStore.addTask(testData.noDatePartition.thirdTask);
					datedUntrashedOpenTaskStore.addTasks(destinationTasks);

					const response = new Response(null, { status: 204 });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText('Tasks without a date');
					const destinationTaskList = page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`);
					await dropTask(destinationTaskList, reorderedTasks, testData.noDatePartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const expectedRank = generateKeyBetween(previousRank, nextRank);
					const requestBody = JSON.stringify({
						id: testData.noDatePartition.thirdTask.id,
						date: testData.todayPartitionDate.toString(),
						rank: expectedRank
					});
					expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
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
		});

		describe('from the non-today-date partition', () => {
			it.each([
				[
					'first',
					testData.noDatePartition.tasks,
					[testData.nonTodayPastPartition.thirdTask, ...testData.noDatePartition.tasks],
					undefined,
					testData.noDatePartition.firstTask.rank
				],
				[
					'middle',
					testData.noDatePartition.tasks,
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
					testData.noDatePartition.tasks,
					[...testData.noDatePartition.tasks, testData.nonTodayPastPartition.thirdTask],
					testData.noDatePartition.thirdTask.rank,
					undefined
				]
			])(
				'moves a task to the %s index of the no-date partition',
				async (_, destinationTasks, reorderedTasks, previousRank, nextRank) => {
					categoryStore.setCategories([category]);
					datedUntrashedOpenTaskStore.addTask(testData.nonTodayPastPartition.thirdTask);
					undatedTaskStore.addTasks(destinationTasks);

					const response = new Response(null, { status: 204 });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayPastPartitionLabel}`);
					const destinationTaskList = page.getByLabelText('Tasks without a date');
					await dropTask(destinationTaskList, reorderedTasks, testData.nonTodayPastPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const expectedRank = generateKeyBetween(previousRank, nextRank);
					const requestBody = JSON.stringify({ id: testData.nonTodayPastPartition.thirdTask.id, rank: expectedRank });
					expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(
						within(sourceTaskList).queryByRole('listitem', { name: testData.nonTodayPastPartition.thirdTask.title })
					).not.toBeInTheDocument();
				}
			);

			it.each([
				[
					'first',
					testData.todayPartition.tasks,
					[testData.nonTodayPastPartition.thirdTask, ...testData.todayPartition.tasks],
					undefined,
					testData.todayPartition.firstTask.rank
				],
				[
					'middle',
					testData.todayPartition.tasks,
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
					testData.todayPartition.tasks,
					[...testData.todayPartition.tasks, testData.nonTodayPastPartition.thirdTask],
					testData.todayPartition.thirdTask.rank,
					undefined
				]
			])(
				'moves a task to the %s index of the today partition',
				async (_, destinationTasks, reorderedTasks, previousRank, nextRank) => {
					categoryStore.setCategories([category]);
					datedUntrashedOpenTaskStore.addTask(testData.nonTodayPastPartition.thirdTask);
					datedUntrashedOpenTaskStore.addTasks(destinationTasks);

					const response = new Response(null, { status: 204 });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayPastPartitionLabel}`);
					const destinationTaskList = page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`);
					await dropTask(destinationTaskList, reorderedTasks, testData.nonTodayPastPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const expectedRank = generateKeyBetween(previousRank, nextRank);
					const requestBody = JSON.stringify({
						id: testData.nonTodayPastPartition.thirdTask.id,
						date: testData.todayPartitionDate.toString(),
						rank: expectedRank
					});
					expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(
						within(sourceTaskList).queryByRole('listitem', { name: testData.nonTodayPastPartition.thirdTask.title })
					).not.toBeInTheDocument();
				}
			);
		});

		describe('from the today partition', () => {
			it.each([
				[
					'first',
					testData.noDatePartition.tasks,
					[testData.todayPartition.thirdTask, ...testData.noDatePartition.tasks],
					undefined,
					testData.noDatePartition.firstTask.rank
				],
				[
					'middle',
					testData.noDatePartition.tasks,
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
					testData.noDatePartition.tasks,
					[...testData.noDatePartition.tasks, testData.todayPartition.thirdTask],
					testData.noDatePartition.thirdTask.rank,
					undefined
				]
			])(
				'moves a task to the %s index of the no-date partition',
				async (_, destinationTasks, reorderedTasks, previousRank, nextRank) => {
					categoryStore.setCategories([category]);
					datedUntrashedOpenTaskStore.addTask(testData.todayPartition.thirdTask);
					undatedTaskStore.addTasks(destinationTasks);

					const response = new Response(null, { status: 204 });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`);
					const destinationTaskList = page.getByLabelText('Tasks without a date');
					await dropTask(destinationTaskList, reorderedTasks, testData.todayPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const expectedRank = generateKeyBetween(previousRank, nextRank);
					const requestBody = JSON.stringify({ id: testData.todayPartition.thirdTask.id, rank: expectedRank });
					expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(
						within(sourceTaskList).queryByRole('listitem', { name: testData.todayPartition.thirdTask.title })
					).not.toBeInTheDocument();
				}
			);

			it.each([
				[
					'first',
					testData.nonTodayPastPartition.tasks,
					[testData.todayPartition.thirdTask, ...testData.nonTodayPastPartition.tasks],
					undefined,
					testData.nonTodayPastPartition.firstTask.rank
				],
				[
					'middle',
					testData.nonTodayPastPartition.tasks,
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
					testData.nonTodayPastPartition.tasks,
					[...testData.nonTodayPastPartition.tasks, testData.todayPartition.thirdTask],
					testData.nonTodayPastPartition.thirdTask.rank,
					undefined
				]
			])(
				'moves a task to the %s index of the non-today-date partition',
				async (_, destinationTasks, reorderedTasks, previousRank, nextRank) => {
					categoryStore.setCategories([category]);
					datedUntrashedOpenTaskStore.addTask(testData.todayPartition.thirdTask);
					datedUntrashedOpenTaskStore.addTasks(destinationTasks);

					const response = new Response(null, { status: 204 });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText(`Tasks dated ${testData.todayPartitionLabel}`);
					const destinationTaskList = page.getByLabelText(`Tasks dated ${testData.nonTodayPastPartitionLabel}`);
					await dropTask(destinationTaskList, reorderedTasks, testData.todayPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const expectedRank = generateKeyBetween(previousRank, nextRank);
					const requestBody = JSON.stringify({
						id: testData.todayPartition.thirdTask.id,
						date: testData.nonTodayPastPartitionDate.toString(),
						rank: expectedRank
					});
					expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(
						within(sourceTaskList).queryByRole('listitem', { name: testData.todayPartition.thirdTask.title })
					).not.toBeInTheDocument();
				}
			);
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
		rank: noDateTaskRanks[index]
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
		rank: todayTaskRanks[index]
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
		rank: nonTodayPastTaskRanks[index]
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
		rank: nonTodayFutureTaskRanks[index]
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
	const ranks = generateNKeysBetween(undefined, undefined, taskCount);
	const tasks = ranks.map((rank, index) => ({
		id: index + 1,
		categoryId: category.id,
		title: `Review project note ${index + 1}`,
		description: `Review and organize project note ${index + 1}.`,
		rank
	}));
	undatedTaskStore.addTasks(tasks);
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
