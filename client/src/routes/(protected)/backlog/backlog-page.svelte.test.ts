import { fireEvent, render, waitFor, within, type RenderResult } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { getLocalTimeZone, parseDate, today, type CalendarDate } from '@internationalized/date';
import { categoryStore } from '$lib/stores/category-store.svelte';
import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
import {
	expectInputToBeInvalid,
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
			addPersonalCategory();
			const page = render(Page);
			await openCreateTaskDialogAndAssertOpen(page);
			await selectCategory(page, personalCategory.title);

			const titleInput = getInputById(page, 'task-title');
			await userEvent.click(titleInput);
			await userEvent.tab();

			const createButton = page.getByRole('button', { name: 'Create' });
			expect(createButton).toBeDisabled();
			await expectInputToBeInvalid(page, 'task-title', 'Title must have at least 1 character');
		});

		it('reacts correctly to too long task title', async () => {
			addPersonalCategory();
			const page = render(Page);
			await openCreateTaskDialogAndAssertOpen(page);
			await selectCategory(page, personalCategory.title);

			const tooLongTitle = 'Plan the weekend hiking route '.repeat(4);
			const titleInput = await enterTaskTitle(page, tooLongTitle);
			await fireEvent.blur(titleInput);

			const createButton = page.getByRole('button', { name: 'Create' });
			expect(createButton).toBeDisabled();
			await expectInputToBeInvalid(page, 'task-title', 'Title must have at most 100 characters');
		});

		it('reacts correctly to too long task description', async () => {
			addPersonalCategory();
			const page = render(Page);
			await openCreateTaskDialogAndAssertOpen(page);
			await selectCategory(page, personalCategory.title);
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
			addPersonalCategory();
			const response = getResponse({ id: 42, rank: 'a0' }, 201);
			fetchMock.mockResolvedValue(response);

			const page = render(Page);
			await openCreateTaskDialogAndAssertOpen(page);

			await enterTaskTitle(page, taskTitle);
			await enterTaskDescription(page, taskDescription);
			await selectCategory(page, personalCategory.title);

			const createButton = page.getByRole('button', { name: 'Create' });
			await userEvent.click(createButton);

			await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
			const createTaskDialog = page.queryByRole('dialog', { name: 'Create Task' });
			expect(createTaskDialog).not.toBeInTheDocument();
			expectTaskCreationRequest(taskTitle, taskDescription);

			const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
			expectTaskToBePresent(tasksWithoutDateList, taskTitle, taskDescription, personalCategory.title);
		});
	});

	describe('task editing', () => {
		it('reacts correctly to missing task title', async () => {
			addPersonalCategory();
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
			addPersonalCategory();
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
			addPersonalCategory();
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

		it('reacts correctly to missing task category', async () => {
			addTaskWithoutCategory();

			const page = render(Page);
			await openEditTaskDialog(page, taskTitle);

			const categorySelect = getSelectById(page, 'task-category');
			await userEvent.click(categorySelect);
			await userEvent.keyboard('{Escape}');
			await userEvent.tab();

			const saveButton = page.getByRole('button', { name: 'Save' });
			expect(saveButton).toBeDisabled();
			await expectSelectToBeInvalid(page, 'task-category', 'Category is required');
		});

		it('cancels and closes the edit-task dialog', async () => {
			addPersonalCategory();
			addTaskWithoutDate();

			const page = render(Page);
			await openEditTaskDialog(page, taskTitle);

			const cancelButton = page.getByRole('button', { name: 'Cancel' });
			await userEvent.click(cancelButton);

			const editTaskDialog = page.queryByRole('dialog', { name: 'Edit Task' });
			expect(editTaskDialog).not.toBeInTheDocument();
		});

		it('successfully edits a task without a date', async () => {
			addPersonalCategory();
			addTaskWithoutDate();

			const response = getResponse({
				id: taskId,
				categoryId: personalCategory.id,
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
			expectTaskUpdateRequest(editedTaskTitle, editedTaskDescription);
			const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
			expectTaskToBePresent(tasksWithoutDateList, editedTaskTitle, editedTaskDescription, personalCategory.title);
		});

		it('successfully edits a task with a non-today date', async () => {
			addPersonalCategory();
			addTaskWithDate(nonTodayDate);

			const response = getResponse({
				id: taskId,
				categoryId: personalCategory.id,
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
			expectTaskUpdateRequest(editedTaskTitle, editedTaskDescription, nonTodayDate);
			const datedTaskList = page.getByLabelText('Tasks dated 15.06.2030');
			expectTaskToBePresent(datedTaskList, editedTaskTitle, editedTaskDescription, personalCategory.title);
		});

		it('successfully edits a task dated today', async () => {
			addPersonalCategory();
			addTaskWithDate(todayDate);

			const response = getResponse({
				id: taskId,
				categoryId: personalCategory.id,
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
			expectTaskUpdateRequest(editedTaskTitle, editedTaskDescription, todayDate);
			const todayTaskList = page.getByLabelText('Tasks dated Today');
			expectTaskToBePresent(todayTaskList, editedTaskTitle, editedTaskDescription, personalCategory.title);
		});
	});

	describe('task completion', () => {
		it.todo('successfully completes a task without a date');
		it.todo('successfully completes a task with a non-today date');
		it.todo('successfully completes a task dated today');
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
const personalCategory = { id: 7, title: 'Personal', color: '#22c55e' };
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

function addPersonalCategory(): void {
	categoryStore.setCategories([personalCategory]);
}

function addTaskWithoutDate(): void {
	undatedTaskStore.addTask({
		id: taskId,
		categoryId: personalCategory.id,
		title: taskTitle,
		description: taskDescription,
		rank: taskRank
	});
}

function addTaskWithoutCategory(): void {
	undatedTaskStore.addTask({
		id: taskId,
		categoryId: undefined as unknown as number,
		title: taskTitle,
		description: taskDescription,
		rank: taskRank
	});
}

function addTaskWithDate(date: CalendarDate): void {
	datedUntrashedOpenTaskStore.addTask({
		id: taskId,
		categoryId: personalCategory.id,
		title: taskTitle,
		description: taskDescription,
		date,
		rank: taskRank
	});
}

function expectTaskCreationRequest(title: string, description: string): void {
	const requestUrl = fetchMock.mock.calls[0]?.[0];
	const requestOptions = fetchMock.mock.calls[0]?.[1] as RequestInit;
	const requestBody = JSON.stringify({ title, description, categoryId: personalCategory.id });
	expect(requestUrl).toBe('/api/tasks');
	expect(requestOptions.method).toBe('POST');
	expect(requestOptions.credentials).toBe('same-origin');
	expect(requestOptions.body).toBe(requestBody);
	const requestHeaders = requestOptions.headers as Headers;
	expect(requestHeaders).toBeInstanceOf(Headers);
	const contentType = requestHeaders.get('Content-Type');
	expect(contentType).toBe('application/json');
}

function expectTaskUpdateRequest(title: string, description: string, date?: CalendarDate): void {
	const requestUrl = fetchMock.mock.calls[0]?.[0];
	const requestOptions = fetchMock.mock.calls[0]?.[1] as RequestInit;
	const requestBody = date
		? JSON.stringify({ title, description, categoryId: personalCategory.id, date: date.toString(), id: taskId })
		: JSON.stringify({ title, description, categoryId: personalCategory.id, id: taskId });
	expect(requestUrl).toBe('/api/tasks');
	expect(requestOptions.method).toBe('PUT');
	expect(requestOptions.credentials).toBe('same-origin');
	expect(requestOptions.body).toBe(requestBody);
	const requestHeaders = requestOptions.headers as Headers;
	expect(requestHeaders).toBeInstanceOf(Headers);
	const contentType = requestHeaders.get('Content-Type');
	expect(contentType).toBe('application/json');
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
