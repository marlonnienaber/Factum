import { fireEvent, render, waitFor, within, type RenderResult } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { getLocalTimeZone, parseDate, today, type CalendarDate } from '@internationalized/date';
import { categoryStore } from '$lib/stores/category-store.svelte';
import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
import {
	expectInputToBeInvalid,
	expectRequestBody,
	expectRequestContentType,
	expectRequestCredentialsMode,
	expectRequestMethod,
	expectRequestUrl,
	expectSelectToBeInvalid,
	expectTextareaToBeInvalid,
	getInputById,
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
});

afterEach(async () => {
	if (document.querySelector('[role="dialog"]')) await userEvent.keyboard('{Escape}');
	document.body.removeAttribute('style');
});

describe('backlog page tests', () => {
	// TODO: Cover navbar actions.

	describe('category operations', () => {
		it.todo('successfully creates and deletes a category');
		it.todo('shows an error toast when deleting a referenced category');
	});

	describe('task creation', () => {
		it('reacts correctly to missing task title', async () => {
			addCategory();
			const page = render(Page);
			await openCreateTaskDialogAndAssertOpen(page);
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
			await openCreateTaskDialogAndAssertOpen(page);
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
			await openCreateTaskDialogAndAssertOpen(page);
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
			await openCreateTaskDialogAndAssertOpen(page);
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
			await openCreateTaskDialogAndAssertOpen(page);

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
			await openCreateTaskDialogAndAssertOpen(page);

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
			const datedTaskList = page.getByLabelText('Tasks dated 15.06.2030');
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
			const requestUrl = `/api/tasks/${taskId}/toggle-done?rank=a0`;
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
			const requestUrl = `/api/tasks/${taskId}/toggle-done?rank=a0`;
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
			const requestUrl = `/api/tasks/${taskId}/toggle-done?rank=a0`;
			expectRequestUrl(fetchMock, requestUrl);
			expectRequestMethod(fetchMock, 'PATCH');
			expectRequestCredentialsMode(fetchMock, 'same-origin');
			expectRequestBody(fetchMock);
			expectTaskNotToBePresent(page, taskTitle);
		});
	});

	describe('task trashing', () => {
		it.todo('successfully trashes a task without a date');
		it.todo('successfully trashes a task with a non-today date');
		it.todo('successfully trashes a task dated today');
	});

	describe('drag and drop', () => {
		describe('within the no-date partition', () => {
			it.todo('moves a task to the first index');
			it.todo('moves a task to the middle index');
			it.todo('moves a task to the last index');
		});

		describe('within the non-today-date partition', () => {
			it.todo('moves a task to the first index');
			it.todo('moves a task to the middle index');
			it.todo('moves a task to the last index');
		});

		describe('between different non-today date groups', () => {
			it.todo('moves a task to the first index');
			it.todo('moves a task to the middle index');
			it.todo('moves a task to the last index');
		});

		describe('within the today partition', () => {
			it.todo('moves a task to the first index');
			it.todo('moves a task to the middle index');
			it.todo('moves a task to the last index');
		});

		describe('from the no-date partition', () => {
			it.todo('moves a task to the first index of the non-today-date partition');
			it.todo('moves a task to the middle index of the non-today-date partition');
			it.todo('moves a task to the last index of the non-today-date partition');
			it.todo('moves a task to the first index of the today partition');
			it.todo('moves a task to the middle index of the today partition');
			it.todo('moves a task to the last index of the today partition');
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

	describe('task batch operations', () => {
		it.todo('successfully completes all tasks');
		it.todo('successfully trashes all tasks');
		it.todo('successfully cancels a batch operation');
		it.todo('disables batch actions when no tasks are selected');
		it.todo('disables batch trash when more than 50 tasks are selected');
		it.todo('disables batch completion when more than 50 tasks are selected');
	});
});

const fetchMock = vi.fn();
const category = { id: 7, title: 'Personal', color: '#22c55e' };
const taskId = 42;
const taskRank = 'a0';
const taskTitle = 'Plan weekend hiking route';
const taskDescription = 'Compare trails near Munich and choose one suitable for Saturday.';
const editedTaskTitle = 'Book the hiking trail transfer';
const editedTaskDescription = 'Reserve the earliest train and the return connection.';
const nonTodayDate = parseDate('2030-06-15');
const todayDate = today(getLocalTimeZone());

async function openCreateTaskDialogAndAssertOpen(page: RenderResult<typeof Page>): Promise<void> {
	const addTaskButton = page.getByRole('button', { name: 'Add Task' });
	await userEvent.click(addTaskButton);
	const createTaskDialog = page.getByRole('dialog', { name: 'Create Task' });
	expect(createTaskDialog).toBeInTheDocument();
}

async function openEditTaskDialog(page: RenderResult<typeof Page>, title: string): Promise<void> {
	const moreOptionsButton = page.getByRole('button', { name: `More options for "${title}"` });
	await fireEvent.click(moreOptionsButton);
	const editButton = await page.findByRole('button', { name: 'Edit' });
	await userEvent.click(editButton);
	const editTaskDialog = page.getByRole('dialog', { name: 'Edit Task' });
	expect(editTaskDialog).toBeInTheDocument();
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

function addCategory(): void {
	categoryStore.setCategories([category]);
}

function addTaskWithoutDate(): void {
	undatedTaskStore.addTask({
		id: taskId,
		categoryId: category.id,
		title: taskTitle,
		description: taskDescription,
		rank: taskRank
	});
}

function addTaskWithDate(date: CalendarDate): void {
	datedUntrashedOpenTaskStore.addTask({
		id: taskId,
		categoryId: category.id,
		title: taskTitle,
		description: taskDescription,
		date,
		rank: taskRank
	});
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
