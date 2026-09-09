import { fireEvent, render, waitFor, within, type RenderResult } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { getLocalTimeZone, now, parseAbsolute, today, type CalendarDate } from '@internationalized/date';
import { generateKeyBetween, generateNKeysBetween } from 'fractional-indexing';
import { toast } from 'svelte-sonner';
import { SOURCES, TRIGGERS, type DndEvent } from 'svelte-dnd-action';
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
	getTextareaById
} from '$lib/utils/testing';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import Page from './+page.svelte';

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

describe('backlog page tests', () => {
	// TODO: Cover navbar actions.

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
			expectRequestBody(fetchMock);
			await waitFor(() => expect(categorySelect).toHaveTextContent('None chosen'));
			expect(categoryStore.categories).not.toContainEqual(category);
		});

		it('shows an error toast when deleting a referenced category', async () => {
			addCategory();
			addTaskWithoutDate();

			const errorMessage = 'Category cannot be deleted while it is referenced by tasks.';
			const errorResponse = getResponse(
				{ errorCode: 'category_still_referenced', message: errorMessage },
				409
			);
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
			expectRequestBody(fetchMock);
			expect(toastErrorMock).toHaveBeenCalledOnce();
			expect(toastErrorMock).toHaveBeenCalledWith(errorMessage);
			expect(categoryStore.categories).toContainEqual(category);
			expect(deleteCategoryButton).toBeInTheDocument();
		});
	});

	describe('task creation', () => {
		it('reacts correctly to missing task title', async () => {
			addCategory();
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
			addCategory();
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
			addCategory();
			const page = render(Page);
			await openCreateTaskDialog(page);
			await selectCategory(page, category.title);
			await enterTaskTitle(page, taskTitle);

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
			await enterTaskTitle(page, taskTitle);

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
			addCategory();
			const response = getResponse({ id: 42, rank: 'a0' }, 201);
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			await openCreateTaskDialog(page);

			await enterTaskTitle(page, taskTitle);
			await enterTaskDescription(page, taskDescription);
			await selectCategory(page, category.title);

			const createButton = page.getByRole('button', { name: 'Create' });
			await userEvent.click(createButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const createTaskDialog = page.queryByRole('dialog', { name: 'Create Task' });
			expect(createTaskDialog).not.toBeInTheDocument();
			const requestBody = JSON.stringify({ title: taskTitle, description: taskDescription, categoryId: category.id });
			expectRequestUrl(fetchMock, '/api/tasks');
			expectRequestMethod(fetchMock, 'POST');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock, requestBody);
			expectRequestContentType(fetchMock, 'application/json');
			const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
			expectTaskToBePresent(tasksWithoutDateList, taskTitle, taskDescription, category.title);
		});
	});

	describe('task editing', () => {
		it('reacts correctly to missing task title', async () => {
			addCategory();
			addTaskWithoutDate();

			const page = render(Page);
			await openEditTaskDialog(page, taskTitle);

			const titleInput = await enterTaskTitle(page, '');
			await fireEvent.blur(titleInput);

			const saveButton = page.getByRole('button', { name: 'Save' });
			expect(saveButton).toBeDisabled();
			await expectInputToBeInvalid(page, 'task-title', 'Title must have at least 1 character');
		});

		it('reacts correctly to too long task title', async () => {
			addCategory();
			addTaskWithoutDate();

			const page = render(Page);
			await openEditTaskDialog(page, taskTitle);

			const tooLongTitle = 'Plan the weekend hiking route '.repeat(4);
			const titleInput = await enterTaskTitle(page, tooLongTitle);
			await fireEvent.blur(titleInput);

			const saveButton = page.getByRole('button', { name: 'Save' });
			expect(saveButton).toBeDisabled();
			await expectInputToBeInvalid(page, 'task-title', 'Title must have at most 100 characters');
		});

		it('reacts correctly to too long task description', async () => {
			addCategory();
			addTaskWithoutDate();

			const page = render(Page);
			await openEditTaskDialog(page, taskTitle);

			const tooLongDescription = 'Compare the available trails, travel times, and weather forecast. '.repeat(9);
			const descriptionInput = await enterTaskDescription(page, tooLongDescription);
			await fireEvent.blur(descriptionInput);

			const saveButton = page.getByRole('button', { name: 'Save' });
			expect(saveButton).toBeDisabled();
			await expectTextareaToBeInvalid(page, 'task-description', 'Title must have at most 500 characters');
		});

		it('cancels and closes the edit-task dialog', async () => {
			addCategory();
			addTaskWithoutDate();

			const page = render(Page);
			await openEditTaskDialog(page, taskTitle);

			const cancelButton = page.getByRole('button', { name: 'Cancel' });
			await userEvent.click(cancelButton);

			const editTaskDialog = page.queryByRole('dialog', { name: 'Edit Task' });
			expect(editTaskDialog).not.toBeInTheDocument();
		});

		it('successfully edits a task without a date', async () => {
			addCategory();
			addTaskWithoutDate();

			const response = getResponse({
				id: taskId,
				categoryId: category.id,
				title: editedTaskTitle,
				description: editedTaskDescription,
				rank: taskRank
			});
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			await openEditTaskDialog(page, taskTitle);

			await enterTaskTitle(page, editedTaskTitle);
			await enterTaskDescription(page, editedTaskDescription);

			const saveButton = page.getByRole('button', { name: 'Save' });
			await userEvent.click(saveButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const editTaskDialog = page.queryByRole('dialog', { name: 'Edit Task' });
			expect(editTaskDialog).not.toBeInTheDocument();
			const requestBody = JSON.stringify({
				title: editedTaskTitle,
				description: editedTaskDescription,
				categoryId: category.id,
				id: taskId
			});
			expectRequestUrl(fetchMock, '/api/tasks');
			expectRequestMethod(fetchMock, 'PUT');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock, requestBody);
			expectRequestContentType(fetchMock, 'application/json');
			const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
			expectTaskToBePresent(tasksWithoutDateList, editedTaskTitle, editedTaskDescription, category.title);
		});

		it('successfully edits a task with a non-today date', async () => {
			addCategory();
			addTaskWithDate(nonTodayDate);

			const response = getResponse({
				id: taskId,
				categoryId: category.id,
				title: editedTaskTitle,
				description: editedTaskDescription,
				date: nonTodayDate.toString(),
				rank: taskRank
			});
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			await openEditTaskDialog(page, taskTitle);

			await enterTaskTitle(page, editedTaskTitle);
			await enterTaskDescription(page, editedTaskDescription);

			const saveButton = page.getByRole('button', { name: 'Save' });
			await userEvent.click(saveButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const editTaskDialog = page.queryByRole('dialog', { name: 'Edit Task' });
			expect(editTaskDialog).not.toBeInTheDocument();
			const serializedDate = nonTodayDate.toString();
			const requestBody = JSON.stringify({
				title: editedTaskTitle,
				description: editedTaskDescription,
				categoryId: category.id,
				date: serializedDate,
				id: taskId
			});
			expectRequestUrl(fetchMock, '/api/tasks');
			expectRequestMethod(fetchMock, 'PUT');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock, requestBody);
			expectRequestContentType(fetchMock, 'application/json');
			const datedTaskList = page.getByLabelText(`Tasks dated ${nonTodayDateLabel}`);
			expectTaskToBePresent(datedTaskList, editedTaskTitle, editedTaskDescription, category.title);
		});

		it('successfully edits a task dated today', async () => {
			addCategory();
			addTaskWithDate(todayDate);

			const response = getResponse({
				id: taskId,
				categoryId: category.id,
				title: editedTaskTitle,
				description: editedTaskDescription,
				date: todayDate.toString(),
				rank: taskRank
			});
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			await openEditTaskDialog(page, taskTitle);

			await enterTaskTitle(page, editedTaskTitle);
			await enterTaskDescription(page, editedTaskDescription);

			const saveButton = page.getByRole('button', { name: 'Save' });
			await userEvent.click(saveButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const editTaskDialog = page.queryByRole('dialog', { name: 'Edit Task' });
			expect(editTaskDialog).not.toBeInTheDocument();
			const serializedDate = todayDate.toString();
			const requestBody = JSON.stringify({
				title: editedTaskTitle,
				description: editedTaskDescription,
				categoryId: category.id,
				date: serializedDate,
				id: taskId
			});
			expectRequestUrl(fetchMock, '/api/tasks');
			expectRequestMethod(fetchMock, 'PUT');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock, requestBody);
			expectRequestContentType(fetchMock, 'application/json');
			const todayTaskList = page.getByLabelText('Tasks dated Today');
			expectTaskToBePresent(todayTaskList, editedTaskTitle, editedTaskDescription, category.title);
		});
	});

	describe('task completion', () => {
		it('successfully completes a task without a date', async () => {
			addCategory();
			addTaskWithoutDate();

			const response = new Response(null, { status: 204 });
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			const completeTaskButton = page.getByRole('checkbox', { name: `Mark "${taskTitle}" as completed` });
			await userEvent.click(completeTaskButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const expectedRank = generateKeyBetween(undefined, undefined);
			const requestUrl = `/api/tasks/${taskId}/toggle-done?rank=${expectedRank}`;
			expectRequestUrl(fetchMock, requestUrl);
			expectRequestMethod(fetchMock, 'PATCH');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock);
			expectTaskNotToBePresent(page, taskTitle);
		});

		it('successfully completes a task with a non-today date', async () => {
			addCategory();
			addTaskWithDate(nonTodayDate);

			const response = new Response(null, { status: 204 });
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			const completeTaskButton = page.getByRole('checkbox', { name: `Mark "${taskTitle}" as completed` });
			await userEvent.click(completeTaskButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const expectedRank = generateKeyBetween(undefined, undefined);
			const requestUrl = `/api/tasks/${taskId}/toggle-done?rank=${expectedRank}`;
			expectRequestUrl(fetchMock, requestUrl);
			expectRequestMethod(fetchMock, 'PATCH');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock);
			expectTaskNotToBePresent(page, taskTitle);
		});

		it('successfully completes a task dated today', async () => {
			addCategory();
			addTaskWithDate(todayDate);

			const response = new Response(null, { status: 204 });
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			const completeTaskButton = page.getByRole('checkbox', { name: `Mark "${taskTitle}" as completed` });
			await userEvent.click(completeTaskButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const expectedRank = generateKeyBetween(undefined, undefined);
			const requestUrl = `/api/tasks/${taskId}/toggle-done?rank=${expectedRank}`;
			expectRequestUrl(fetchMock, requestUrl);
			expectRequestMethod(fetchMock, 'PATCH');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock);
			expectTaskNotToBePresent(page, taskTitle);
		});
	});

	describe('task trashing', () => {
		it('successfully trashes a task without a date', async () => {
			addCategory();
			addTaskWithoutDate();

			const response = new Response(null, { status: 204 });
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			const moreOptionsButton = page.getByRole('button', { name: `More options for "${taskTitle}"` });
			await fireEvent.click(moreOptionsButton);
			const trashButton = await page.findByRole('button', { name: 'Trash' });
			await userEvent.click(trashButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			expectRequestPath(fetchMock, `/api/tasks/${taskId}/trash`);
			const trashedDate = getRequestSearchParameter(fetchMock, 'trashedDate');
			expect(trashedDate).not.toBeNull();
			if (trashedDate === null) throw new Error('Expected the trash request to contain a trashedDate.');
			const parsedTrashedDate = parseAbsolute(trashedDate, localTimeZone);
			const currentDate = now(localTimeZone);
			const dateDifference = Math.abs(parsedTrashedDate.compare(currentDate));
			expect(dateDifference).toBeLessThanOrEqual(maximumDateDifference);
			expectRequestMethod(fetchMock, 'PATCH');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock);
			expectTaskNotToBePresent(page, taskTitle);
		});

		it('successfully trashes a task with a non-today date', async () => {
			addCategory();
			addTaskWithDate(nonTodayDate);

			const response = new Response(null, { status: 204 });
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			const moreOptionsButton = page.getByRole('button', { name: `More options for "${taskTitle}"` });
			await fireEvent.click(moreOptionsButton);
			const trashButton = await page.findByRole('button', { name: 'Trash' });
			await userEvent.click(trashButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			expectRequestPath(fetchMock, `/api/tasks/${taskId}/trash`);
			const trashedDate = getRequestSearchParameter(fetchMock, 'trashedDate');
			expect(trashedDate).not.toBeNull();
			if (trashedDate === null) throw new Error('Expected the trash request to contain a trashedDate.');
			const parsedTrashedDate = parseAbsolute(trashedDate, localTimeZone);
			const currentDate = now(localTimeZone);
			const dateDifference = Math.abs(parsedTrashedDate.compare(currentDate));
			expect(dateDifference).toBeLessThanOrEqual(maximumDateDifference);
			expectRequestMethod(fetchMock, 'PATCH');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock);
			expectTaskNotToBePresent(page, taskTitle);
		});

		it('successfully trashes a task dated today', async () => {
			addCategory();
			addTaskWithDate(todayDate);

			const response = new Response(null, { status: 204 });
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			const moreOptionsButton = page.getByRole('button', { name: `More options for "${taskTitle}"` });
			await fireEvent.click(moreOptionsButton);
			const trashButton = await page.findByRole('button', { name: 'Trash' });
			await userEvent.click(trashButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			expectRequestPath(fetchMock, `/api/tasks/${taskId}/trash`);
			const trashedDate = getRequestSearchParameter(fetchMock, 'trashedDate');
			expect(trashedDate).not.toBeNull();
			if (trashedDate === null) throw new Error('Expected the trash request to contain a trashedDate.');
			const parsedTrashedDate = parseAbsolute(trashedDate, localTimeZone);
			const currentDate = now(localTimeZone);
			const dateDifference = Math.abs(parsedTrashedDate.compare(currentDate));
			expect(dateDifference).toBeLessThanOrEqual(maximumDateDifference);
			expectRequestMethod(fetchMock, 'PATCH');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock);
			expectTaskNotToBePresent(page, taskTitle);
		});
	});

	describe('task batch operations', () => {
		it('successfully completes all tasks', async () => {
			addCategory();
			addTasksForBatchOperation();

			const response = getResponse({ datedUntrashedDoneTasks: completedBatchTasks, trashedTasks: [] });
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			await clickOnSelectTasksButton(page);
			await clickOnTask(page, taskTitle);
			await clickOnTask(page, datedTaskTitle);
			await clickOnTask(page, todayTaskTitle);

			const completeAllButton = page.getByRole('button', { name: 'Complete All' });
			await userEvent.click(completeAllButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const requestBody = JSON.stringify({ taskIds: batchTaskIds });
			expectRequestUrl(fetchMock, '/api/tasks/batch-complete');
			expectRequestMethod(fetchMock, 'PATCH');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock, requestBody);
			expectRequestContentType(fetchMock, 'application/json');
			expectTaskNotToBePresent(page, taskTitle);
			expectTaskNotToBePresent(page, datedTaskTitle);
			expectTaskNotToBePresent(page, todayTaskTitle);
			expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toHaveLength(3);
			const selectTasksButton = page.getByRole('button', { name: 'Select Tasks' });
			expect(selectTasksButton).toBeInTheDocument();
		});

		it('successfully trashes all tasks', async () => {
			addCategory();
			addTasksForBatchOperation();

			const response = new Response(null, { status: 204 });
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			await clickOnSelectTasksButton(page);
			await clickOnTask(page, taskTitle);
			await clickOnTask(page, datedTaskTitle);
			await clickOnTask(page, todayTaskTitle);

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
			expect(dateDifference).toBeLessThanOrEqual(maximumDateDifference);
			const requestBody = JSON.stringify({ taskIds: batchTaskIds });
			expectRequestMethod(fetchMock, 'PATCH');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock, requestBody);
			expectRequestContentType(fetchMock, 'application/json');
			expectTaskNotToBePresent(page, taskTitle);
			expectTaskNotToBePresent(page, datedTaskTitle);
			expectTaskNotToBePresent(page, todayTaskTitle);
			expect(trashedTaskStore.trashedTasks).toHaveLength(3);
			const selectTasksButton = page.getByRole('button', { name: 'Select Tasks' });
			expect(selectTasksButton).toBeInTheDocument();
		});

		it('successfully cancels a batch operation', async () => {
			addCategory();
			addTaskWithoutDate();

			const page = render(Page);
			await clickOnSelectTasksButton(page);
			await clickOnTask(page, taskTitle);

			const cancelSelectionButton = page.getByRole('button', { name: 'Cancel Selection' });
			await userEvent.click(cancelSelectionButton);

			const selectTasksButton = page.getByRole('button', { name: 'Select Tasks' });
			expect(selectTasksButton).toBeInTheDocument();
			const completeTaskButton = page.getByRole('checkbox', { name: `Mark "${taskTitle}" as completed` });
			expect(completeTaskButton).toBeInTheDocument();
			expect(fetchMock).not.toHaveBeenCalled();
		});

		it('disables batch actions when no tasks are selected', async () => {
			addCategory();
			addTaskWithoutDate();

			const page = render(Page);
			await clickOnSelectTasksButton(page);

			const completeAllButton = page.getByRole('button', { name: 'Complete All' });
			const trashAllButton = page.getByRole('button', { name: 'Trash All' });
			expect(completeAllButton).toBeDisabled();
			expect(trashAllButton).toBeDisabled();
			expect(page.getByText('0 / 50')).toBeInTheDocument();
		});

		it('disables batch trash when more than 50 tasks are selected', async () => {
			addCategory();
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
			addCategory();
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
				addCategory();
				undatedTaskStore.addTasks(noDateTasks);

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
				const reorderedTasks = [fourthNoDateTask, firstNoDateTask, secondNoDateTask, thirdNoDateTask];
				await dropTask(tasksWithoutDateList, reorderedTasks, fourthNoDateTask.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const expectedRank = generateKeyBetween(undefined, firstNoDateTask.rank);
				const requestBody = JSON.stringify({ id: fourthNoDateTask.id, rank: expectedRank });
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(tasksWithoutDateList, reorderedTasks);
			});

			it('moves a task to the middle index', async () => {
				addCategory();
				undatedTaskStore.addTasks(noDateTasks);

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
				const reorderedTasks = [firstNoDateTask, secondNoDateTask, fourthNoDateTask, thirdNoDateTask];
				await dropTask(tasksWithoutDateList, reorderedTasks, fourthNoDateTask.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const expectedRank = generateKeyBetween(secondNoDateTask.rank, thirdNoDateTask.rank);
				const requestBody = JSON.stringify({ id: fourthNoDateTask.id, rank: expectedRank });
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(tasksWithoutDateList, reorderedTasks);
			});

			it('moves a task to the last index', async () => {
				addCategory();
				undatedTaskStore.addTasks(noDateTasks);

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
				const reorderedTasks = [secondNoDateTask, thirdNoDateTask, fourthNoDateTask, firstNoDateTask];
				await dropTask(tasksWithoutDateList, reorderedTasks, firstNoDateTask.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const expectedRank = generateKeyBetween(fourthNoDateTask.rank, undefined);
				const requestBody = JSON.stringify({ id: firstNoDateTask.id, rank: expectedRank });
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
				addCategory();
				datedUntrashedOpenTaskStore.addTasks(todayTasks);

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const todayTaskList = page.getByLabelText('Tasks dated Today');
				const reorderedTasks = [fourthTodayTask, firstTodayTask, secondTodayTask, thirdTodayTask];
				await dropTask(todayTaskList, reorderedTasks, fourthTodayTask.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const expectedRank = generateKeyBetween(undefined, firstTodayTask.rank);
				const requestBody = JSON.stringify({ id: fourthTodayTask.id, date: todayDate.toString(), rank: expectedRank });
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(todayTaskList, reorderedTasks);
			});

			it('moves a task to the middle index', async () => {
				addCategory();
				datedUntrashedOpenTaskStore.addTasks(todayTasks);

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const todayTaskList = page.getByLabelText('Tasks dated Today');
				const reorderedTasks = [firstTodayTask, secondTodayTask, fourthTodayTask, thirdTodayTask];
				await dropTask(todayTaskList, reorderedTasks, fourthTodayTask.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const expectedRank = generateKeyBetween(secondTodayTask.rank, thirdTodayTask.rank);
				const requestBody = JSON.stringify({ id: fourthTodayTask.id, date: todayDate.toString(), rank: expectedRank });
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(todayTaskList, reorderedTasks);
			});

			it('moves a task to the last index', async () => {
				addCategory();
				datedUntrashedOpenTaskStore.addTasks(todayTasks);

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const todayTaskList = page.getByLabelText('Tasks dated Today');
				const reorderedTasks = [secondTodayTask, thirdTodayTask, fourthTodayTask, firstTodayTask];
				await dropTask(todayTaskList, reorderedTasks, firstTodayTask.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const expectedRank = generateKeyBetween(fourthTodayTask.rank, undefined);
				const requestBody = JSON.stringify({ id: firstTodayTask.id, date: todayDate.toString(), rank: expectedRank });
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
				addCategory();
				datedUntrashedOpenTaskStore.addTasks(nonTodayTasks);

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const nonTodayTaskList = page.getByLabelText(`Tasks dated ${nonTodayDateLabel}`);
				const reorderedTasks = [fourthNonTodayTask, firstNonTodayTask, secondNonTodayTask, thirdNonTodayTask];
				await dropTask(nonTodayTaskList, reorderedTasks, fourthNonTodayTask.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const expectedRank = generateKeyBetween(undefined, firstNonTodayTask.rank);
				const requestBody = JSON.stringify({ id: fourthNonTodayTask.id, date: nonTodayDate.toString(), rank: expectedRank });
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(nonTodayTaskList, reorderedTasks);
			});

			it('moves a task to the middle index', async () => {
				addCategory();
				datedUntrashedOpenTaskStore.addTasks(nonTodayTasks);

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const nonTodayTaskList = page.getByLabelText(`Tasks dated ${nonTodayDateLabel}`);
				const reorderedTasks = [firstNonTodayTask, secondNonTodayTask, fourthNonTodayTask, thirdNonTodayTask];
				await dropTask(nonTodayTaskList, reorderedTasks, fourthNonTodayTask.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const expectedRank = generateKeyBetween(secondNonTodayTask.rank, thirdNonTodayTask.rank);
				const requestBody = JSON.stringify({ id: fourthNonTodayTask.id, date: nonTodayDate.toString(), rank: expectedRank });
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(nonTodayTaskList, reorderedTasks);
			});

			it('moves a task to the last index', async () => {
				addCategory();
				datedUntrashedOpenTaskStore.addTasks(nonTodayTasks);

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const nonTodayTaskList = page.getByLabelText(`Tasks dated ${nonTodayDateLabel}`);
				const reorderedTasks = [secondNonTodayTask, thirdNonTodayTask, fourthNonTodayTask, firstNonTodayTask];
				await dropTask(nonTodayTaskList, reorderedTasks, firstNonTodayTask.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const expectedRank = generateKeyBetween(fourthNonTodayTask.rank, undefined);
				const requestBody = JSON.stringify({ id: firstNonTodayTask.id, date: nonTodayDate.toString(), rank: expectedRank });
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
				addCategory();
				datedUntrashedOpenTaskStore.addTasks(nonTodayTasks);
				datedUntrashedOpenTaskStore.addTasks(secondNonTodayTasks);

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const sourceTaskList = page.getByLabelText(`Tasks dated ${nonTodayDateLabel}`);
				const destinationTaskList = page.getByLabelText(`Tasks dated ${secondNonTodayDateLabel}`);
				const reorderedTasks = [firstNonTodayTask, firstSecondNonTodayTask, secondSecondNonTodayTask, thirdSecondNonTodayTask];
				await dropTask(destinationTaskList, reorderedTasks, firstNonTodayTask.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const expectedRank = generateKeyBetween(undefined, firstSecondNonTodayTask.rank);
				const requestBody = JSON.stringify({ id: firstNonTodayTask.id, date: secondNonTodayDate.toString(), rank: expectedRank });
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(destinationTaskList, reorderedTasks);
				expect(
					within(sourceTaskList).queryByRole('listitem', { name: firstNonTodayTask.title })
				).not.toBeInTheDocument();
			});

			it('moves a task to the middle index', async () => {
				addCategory();
				datedUntrashedOpenTaskStore.addTasks(nonTodayTasks);
				datedUntrashedOpenTaskStore.addTasks(secondNonTodayTasks);

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const sourceTaskList = page.getByLabelText(`Tasks dated ${nonTodayDateLabel}`);
				const destinationTaskList = page.getByLabelText(`Tasks dated ${secondNonTodayDateLabel}`);
				const reorderedTasks = [firstSecondNonTodayTask, firstNonTodayTask, secondSecondNonTodayTask, thirdSecondNonTodayTask];
				await dropTask(destinationTaskList, reorderedTasks, firstNonTodayTask.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const expectedRank = generateKeyBetween(firstSecondNonTodayTask.rank, secondSecondNonTodayTask.rank);
				const requestBody = JSON.stringify({ id: firstNonTodayTask.id, date: secondNonTodayDate.toString(), rank: expectedRank });
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(destinationTaskList, reorderedTasks);
				expect(
					within(sourceTaskList).queryByRole('listitem', { name: firstNonTodayTask.title })
				).not.toBeInTheDocument();
			});

			it('moves a task to the last index', async () => {
				addCategory();
				datedUntrashedOpenTaskStore.addTasks(nonTodayTasks);
				datedUntrashedOpenTaskStore.addTasks(secondNonTodayTasks);

				const response = new Response(null, { status: 204 });
				fetchMock.mockResolvedValue(response);

				const page = render(Page);
				const sourceTaskList = page.getByLabelText(`Tasks dated ${nonTodayDateLabel}`);
				const destinationTaskList = page.getByLabelText(`Tasks dated ${secondNonTodayDateLabel}`);
				const reorderedTasks = [firstSecondNonTodayTask, secondSecondNonTodayTask, thirdSecondNonTodayTask, firstNonTodayTask];
				await dropTask(destinationTaskList, reorderedTasks, firstNonTodayTask.id);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				const expectedRank = generateKeyBetween(thirdSecondNonTodayTask.rank, undefined);
				const requestBody = JSON.stringify({ id: firstNonTodayTask.id, date: secondNonTodayDate.toString(), rank: expectedRank });
				expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
				expectRequestMethod(fetchMock, 'PATCH');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, requestBody);
				expectRequestContentType(fetchMock, 'application/json');
				expectTaskOrder(destinationTaskList, reorderedTasks);
				expect(
					within(sourceTaskList).queryByRole('listitem', { name: firstNonTodayTask.title })
				).not.toBeInTheDocument();
			});
		});

		describe('from the no-date partition', () => {
			it.each([
				[
					'first',
					nonTodayDate,
					nonTodayDateLabel,
					nonTodayTasks,
					[sourceNoDateTask, ...nonTodayTasks],
					undefined,
					firstNonTodayTask.rank
				],
				[
					'middle',
					nonTodayDate,
					nonTodayDateLabel,
					nonTodayTasks,
					[firstNonTodayTask, sourceNoDateTask, secondNonTodayTask, thirdNonTodayTask, fourthNonTodayTask],
					firstNonTodayTask.rank,
					secondNonTodayTask.rank
				],
				[
					'last',
					nonTodayDate,
					nonTodayDateLabel,
					nonTodayTasks,
					[...nonTodayTasks, sourceNoDateTask],
					fourthNonTodayTask.rank,
					undefined
				]
			])(
				'moves a task to the %s index of the non-today-date partition',
				async (_, destinationDate, destinationDateLabel, destinationTasks, reorderedTasks, previousRank, nextRank) => {
					addCategory();
					undatedTaskStore.addTask(sourceNoDateTask);
					datedUntrashedOpenTaskStore.addTasks(destinationTasks);

					const response = new Response(null, { status: 204 });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText('Tasks without a date');
					const destinationTaskList = page.getByLabelText(`Tasks dated ${destinationDateLabel}`);
					await dropTask(destinationTaskList, reorderedTasks, sourceNoDateTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const expectedRank = generateKeyBetween(previousRank, nextRank);
					const requestBody = JSON.stringify({ id: sourceNoDateTask.id, date: destinationDate.toString(), rank: expectedRank });
					expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(within(sourceTaskList).queryByRole('listitem', { name: sourceNoDateTask.title })).not.toBeInTheDocument();
				}
			);

			it.each([
				['first', todayTasks, [sourceNoDateTask, ...todayTasks], undefined, firstTodayTask.rank],
				[
					'middle',
					todayTasks,
					[firstTodayTask, sourceNoDateTask, secondTodayTask, thirdTodayTask, fourthTodayTask],
					firstTodayTask.rank,
					secondTodayTask.rank
				],
				['last', todayTasks, [...todayTasks, sourceNoDateTask], fourthTodayTask.rank, undefined]
			])(
				'moves a task to the %s index of the today partition',
				async (_, destinationTasks, reorderedTasks, previousRank, nextRank) => {
					addCategory();
					undatedTaskStore.addTask(sourceNoDateTask);
					datedUntrashedOpenTaskStore.addTasks(destinationTasks);

					const response = new Response(null, { status: 204 });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const sourceTaskList = page.getByLabelText('Tasks without a date');
					const destinationTaskList = page.getByLabelText('Tasks dated Today');
					await dropTask(destinationTaskList, reorderedTasks, sourceNoDateTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const expectedRank = generateKeyBetween(previousRank, nextRank);
					const requestBody = JSON.stringify({ id: sourceNoDateTask.id, date: todayDate.toString(), rank: expectedRank });
					expectRequestUrl(fetchMock, '/api/tasks/change-date-and-or-rank');
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(destinationTaskList, reorderedTasks);
					expect(within(sourceTaskList).queryByRole('listitem', { name: sourceNoDateTask.title })).not.toBeInTheDocument();
				}
			);
		});

		describe('from the non-today-date partition', () => {
			it.todo('moves a task to the first index of the no-date partition');
			it.todo('moves a task to the middle index of the no-date partition');
			it.todo('moves a task to the last index of the no-date partition');
			it.todo('moves a task to the first index of the today partition');
			it.todo('moves a task to the middle index of the today partition');
			it.todo('moves a task to the last index of the today partition');
		});

		describe('from the today partition', () => {
			it.todo('moves a task to the first index of the no-date partition');
			it.todo('moves a task to the middle index of the no-date partition');
			it.todo('moves a task to the last index of the no-date partition');
			it.todo('moves a task to the first index of the non-today-date partition');
			it.todo('moves a task to the middle index of the non-today-date partition');
			it.todo('moves a task to the last index of the non-today-date partition');
		});
	});
});

const fetchMock = vi.fn();
const toastErrorMock = vi.mocked(toast.error);
const category = { id: 7, title: 'Personal', color: '#22c55e' };
const taskId = 42;
const taskRank = 'a0';
const taskTitle = 'Plan weekend hiking route';
const taskDescription = 'Compare trails near Munich and choose one suitable for Saturday.';
const datedTaskId = 43;
const datedTaskTitle = 'Prepare the trip packing list';
const datedTaskDescription = 'List the equipment needed for the overnight stay.';
const todayTaskId = 44;
const todayTaskTitle = 'Buy food for the hike';
const todayTaskDescription = 'Pick up water, fruit, and trail snacks before departure.';
const editedTaskTitle = 'Book the hiking trail transfer';
const editedTaskDescription = 'Reserve the earliest train and the return connection.';
const localTimeZone = getLocalTimeZone();
const todayDate = today(localTimeZone);
const nonTodayDate = todayDate.add({ days: 1 });
const nonTodayDateLabel = `${String(nonTodayDate.day).padStart(2, '0')}.${String(nonTodayDate.month).padStart(2, '0')}.${nonTodayDate.year}`;
const maximumDateDifference = 5_000;
const batchTaskIds = [taskId, datedTaskId, todayTaskId];
const batchTaskRanks = generateNKeysBetween(undefined, undefined, batchTaskIds.length);
const completedBatchTaskRanks = generateNKeysBetween(undefined, undefined, batchTaskIds.length);
const completedBatchTasks = [
	{
		id: taskId,
		categoryId: category.id,
		title: taskTitle,
		description: taskDescription,
		date: todayDate.toString(),
		rank: completedBatchTaskRanks[0]
	},
	{
		id: datedTaskId,
		categoryId: category.id,
		title: datedTaskTitle,
		description: datedTaskDescription,
		date: nonTodayDate.toString(),
		rank: completedBatchTaskRanks[1]
	},
	{
		id: todayTaskId,
		categoryId: category.id,
		title: todayTaskTitle,
		description: todayTaskDescription,
		date: todayDate.toString(),
		rank: completedBatchTaskRanks[2]
	}
];
const noDateTaskRanks = generateNKeysBetween(undefined, undefined, 4);
const firstNoDateTask = {
	id: 101,
	categoryId: category.id,
	title: 'Compare hiking routes',
	description: 'Compare the distance and elevation of nearby routes.',
	rank: noDateTaskRanks[0]
};
const secondNoDateTask = {
	id: 102,
	categoryId: category.id,
	title: 'Check the weather forecast',
	description: 'Review the hourly forecast for the planned hiking day.',
	rank: noDateTaskRanks[1]
};
const thirdNoDateTask = {
	id: 103,
	categoryId: category.id,
	title: 'Prepare hiking equipment',
	description: 'Inspect boots, rainwear, and the first-aid kit.',
	rank: noDateTaskRanks[2]
};
const fourthNoDateTask = {
	id: 104,
	categoryId: category.id,
	title: 'Buy trail provisions',
	description: 'Buy water and food for the day on the trail.',
	rank: noDateTaskRanks[3]
};
const noDateTasks = [firstNoDateTask, secondNoDateTask, thirdNoDateTask, fourthNoDateTask];
const sourceNoDateTask = {
	id: 105,
	categoryId: category.id,
	title: 'Reserve the mountain hut',
	description: 'Confirm availability for the planned overnight stop.',
	rank: generateKeyBetween(fourthNoDateTask.rank, undefined)
};
const todayTaskRanks = generateNKeysBetween(undefined, undefined, 4);
const firstTodayTask = { ...firstNoDateTask, id: 201, date: todayDate, rank: todayTaskRanks[0] };
const secondTodayTask = { ...secondNoDateTask, id: 202, date: todayDate, rank: todayTaskRanks[1] };
const thirdTodayTask = { ...thirdNoDateTask, id: 203, date: todayDate, rank: todayTaskRanks[2] };
const fourthTodayTask = { ...fourthNoDateTask, id: 204, date: todayDate, rank: todayTaskRanks[3] };
const todayTasks = [firstTodayTask, secondTodayTask, thirdTodayTask, fourthTodayTask];
const nonTodayTaskRanks = generateNKeysBetween(undefined, undefined, 4);
const firstNonTodayTask = { ...firstNoDateTask, id: 301, date: nonTodayDate, rank: nonTodayTaskRanks[0] };
const secondNonTodayTask = { ...secondNoDateTask, id: 302, date: nonTodayDate, rank: nonTodayTaskRanks[1] };
const thirdNonTodayTask = { ...thirdNoDateTask, id: 303, date: nonTodayDate, rank: nonTodayTaskRanks[2] };
const fourthNonTodayTask = { ...fourthNoDateTask, id: 304, date: nonTodayDate, rank: nonTodayTaskRanks[3] };
const nonTodayTasks = [firstNonTodayTask, secondNonTodayTask, thirdNonTodayTask, fourthNonTodayTask];
const secondNonTodayDate = todayDate.add({ days: 2 });
const secondNonTodayDateLabel = `${String(secondNonTodayDate.day).padStart(2, '0')}.${String(secondNonTodayDate.month).padStart(2, '0')}.${secondNonTodayDate.year}`;
const secondNonTodayTaskRanks = generateNKeysBetween(undefined, undefined, 3);
const firstSecondNonTodayTask = {
	...firstNoDateTask,
	id: 401,
	title: 'Confirm trail transport',
	date: secondNonTodayDate,
	rank: secondNonTodayTaskRanks[0]
};
const secondSecondNonTodayTask = {
	...secondNoDateTask,
	id: 402,
	title: 'Charge navigation devices',
	date: secondNonTodayDate,
	rank: secondNonTodayTaskRanks[1]
};
const thirdSecondNonTodayTask = {
	...thirdNoDateTask,
	id: 403,
	title: 'Share the route plan',
	date: secondNonTodayDate,
	rank: secondNonTodayTaskRanks[2]
};
const secondNonTodayTasks = [firstSecondNonTodayTask, secondSecondNonTodayTask, thirdSecondNonTodayTask];

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

async function enterTaskDescription(page: RenderResult<typeof Page>, description: string): Promise<HTMLTextAreaElement> {
	const descriptionInput = getTextareaById(page, 'task-description');
	await fireEvent.input(descriptionInput, { target: { value: description } });
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

function addCategory(): void {
	categoryStore.setCategories([category]);
}

function addTaskWithoutDate(
	id = taskId,
	title = taskTitle,
	description = taskDescription,
	rank = taskRank
): void {
	undatedTaskStore.addTask({
		id,
		categoryId: category.id,
		title,
		description,
		rank
	});
}

function addTaskWithDate(
	date: CalendarDate,
	id = taskId,
	title = taskTitle,
	description = taskDescription,
	rank = taskRank
): void {
	datedUntrashedOpenTaskStore.addTask({
		id,
		categoryId: category.id,
		title,
		description,
		date,
		rank
	});
}

function addTasksForBatchOperation(): void {
	addTaskWithoutDate(taskId, taskTitle, taskDescription, batchTaskRanks[0]);
	addTaskWithDate(nonTodayDate, datedTaskId, datedTaskTitle, datedTaskDescription, batchTaskRanks[1]);
	addTaskWithDate(todayDate, todayTaskId, todayTaskTitle, todayTaskDescription, batchTaskRanks[2]);
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

async function dropTask(taskList: HTMLElement, newOrderedTasks: (UndatedTaskResponse | DatedUntrashedOpenTaskResponse)[], taskId: number): Promise<void> {
	const event = new CustomEvent<DndEvent<UndatedTaskResponse | DatedUntrashedOpenTaskResponse>>('finalize', {
		detail: {
			items: newOrderedTasks,
			info: { trigger: TRIGGERS.DROPPED_INTO_ZONE, id: String(taskId), source: SOURCES.POINTER }
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

function expectTaskToBePresent(taskList: HTMLElement, title: string, description: string, categoryName: string): void {
	const taskListContent = within(taskList);
	const task = taskListContent.getByRole('listitem', { name: title });
	const taskContent = within(task);
	const taskDescription = taskContent.getByText(description);
	const taskCategory = taskContent.getByText(categoryName);
	expect(task).toBeInTheDocument();
	expect(taskDescription).toBeInTheDocument();
	expect(taskCategory).toBeInTheDocument();
}

function expectTaskNotToBePresent(page: RenderResult<typeof Page>, title: string): void {
	const task = page.queryByRole('listitem', { name: title });
	expect(task).not.toBeInTheDocument();
}
