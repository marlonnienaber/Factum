import { fireEvent, render, waitFor, within, type RenderResult } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
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
	categoryStore.setCategories([personalCategory]);
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
			const requestBody = JSON.stringify({
				title: taskTitle,
				description: taskDescription,
				categoryId: personalCategory.id
			});
			const requestUrl = fetchMock.mock.calls[0]?.[0];
			const requestOptions = fetchMock.mock.calls[0]?.[1] as RequestInit;
			expect(requestUrl).toBe('/api/tasks');
			expect(requestOptions.method).toBe('POST');
			expect(requestOptions.credentials).toBe('same-origin');
			expect(requestOptions.body).toBe(requestBody);
			const requestHeaders = requestOptions.headers as Headers;
			expect(requestHeaders).toBeInstanceOf(Headers);
			const contentType = requestHeaders.get('Content-Type');
			expect(contentType).toBe('application/json');
			const createTaskDialog = page.queryByRole('dialog', { name: 'Create Task' });
			expect(createTaskDialog).not.toBeInTheDocument();

			const tasksWithoutDateList = page.getByLabelText('Tasks without a date');
			expectTaskToBePresent(tasksWithoutDateList, taskTitle, taskDescription, personalCategory.title);
		});
	});

	describe('task editing', () => {
		it.todo('reacts correctly to missing task title');
		it.todo('reacts correctly to too long task title');
		it.todo('reacts correctly to too long task description');
		it.todo('reacts correctly to missing task category');
		it.todo('cancels and closes the edit-task dialog');
		it.todo('successfully edits a task without a date');
		it.todo('successfully edits a task with a non-today date');
		it.todo('successfully edits a task dated today');
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
const taskTitle = 'Plan weekend hiking route';
const taskDescription = 'Compare trails near Munich and choose one suitable for Saturday.';

async function openCreateTaskDialogAndAssertOpen(page: RenderResult<typeof Page>): Promise<void> {
	const addTaskButton = page.getByRole('button', { name: 'Add Task' });
	await userEvent.click(addTaskButton);
	const createTaskDialog = page.getByRole('dialog', { name: 'Create Task' });
	expect(createTaskDialog).toBeInTheDocument();
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
