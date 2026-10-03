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
	expectRequestPath,
	expectRequestUrl,
	expectTextareaToBeInvalid,
	getInputById,
	getRequestSearchParameter,
	getResponse,
	getTextareaById,
	intersectionObserverMock,
	mockPendingFetch
} from '$lib/utils/testing';
import { getLocalTimeZone, today, type DateValue } from '@internationalized/date';
import { generateKeyBetween, generateNKeysBetween } from 'fractional-indexing';
import { type DndEvent, SOURCES, TRIGGERS } from 'svelte-dnd-action';
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
				const trashAllButton = page.getByRole('button', { name: 'Trash' });
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
				const openAllButton = page.getByRole('button', { name: 'Open' });
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
				const cancelSelectionButton = page.getByRole('button', { name: 'Cancel' });
				await userEvent.click(cancelSelectionButton);

				expect(page.getByRole('button', { name: 'Select Tasks' })).toBeInTheDocument();
				expect(page.getByRole('checkbox', { name: `Mark "${task.title}" as completed` })).toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});

			it('disables batch actions when no tasks are selected', async () => {
				const page = render(Page);
				await clickOnSelectTasksButton(page);

				expect(page.getByRole('button', { name: 'Open' })).toBeDisabled();
				expect(page.getByRole('button', { name: 'Trash' })).toBeDisabled();
				expect(page.getByText('0 / 50')).toBeInTheDocument();
			});

			it('disables batch open when more than 50 tasks are selected', async () => {
				const taskTitles = addTasksExceedingMaximumBatchSize();

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const title of taskTitles) await clickOnTask(page, title);

				expect(page.getByRole('button', { name: 'Open' })).toBeDisabled();
				expect(page.getByText('51 / 50')).toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});

			it('disables batch trash when more than 50 tasks are selected', async () => {
				const taskTitles = addTasksExceedingMaximumBatchSize();

				const page = render(Page);
				await clickOnSelectTasksButton(page);
				for (const title of taskTitles) await clickOnTask(page, title);

				expect(page.getByRole('button', { name: 'Trash' })).toBeDisabled();
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

		it('shows trash info popover on info icon hover', async () => {
			const page = render(Page);
			const infoButton = page.getByRole('button', { name: 'About automatic task deletion' });
			await userEvent.hover(infoButton);

			expect(
				await page.findByText('Trashed tasks are automatically deleted every day at midnight German time.')
			).toBeInTheDocument();
		});

		describe('filtering', () => {
			it('applies category filter', async () => {
				const completedTasks = [
					testData.pastCompletedPartition.secondTask,
					testData.futureCompletedPartition.firstTask,
					testData.futureCompletedPartition.thirdTask
				];
				const trashedTasks = [testData.trashedPartition.datedOpenTask];
				mockFilterResponses(completedTasks, trashedTasks);

				const page = render(Page);
				const excludePersonalButtons = page.getAllByRole('button', { name: 'Exclude Personal' });
				await userEvent.click(excludePersonalButtons[0]!);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
				expectRequestUrl(fetchMock, '/api/tasks/dated-untrashed-done?categoryId=8');
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expectRequestUrl(fetchMock, '/api/tasks/trashed?categoryId=8', 1);
				expectRequestMethod(fetchMock, 'GET', 1);
				expectRequestCredentialsMode(fetchMock, 'same-origin', 1);
				expectRequestBody(fetchMock, undefined, 1);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toEqual(completedTasks);
				expect(trashedTaskStore.trashedTasks).toEqual(trashedTasks);
			});

			it('applies start date filter', async () => {
				const completedTasks = testData.futureCompletedPartition.tasks;
				const trashedTasks = [testData.trashedPartition.datedOpenTask];
				mockFilterResponses(completedTasks, trashedTasks);

				const page = render(Page);
				await enterFilterDate(page, 'Start', todayDate);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
				const queryParameters = `categoryId=7&categoryId=8&startDate=${todayDate.toString()}`;
				expectRequestUrl(fetchMock, `/api/tasks/dated-untrashed-done?${queryParameters}`);
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expectRequestUrl(fetchMock, `/api/tasks/trashed?${queryParameters}`, 1);
				expectRequestMethod(fetchMock, 'GET', 1);
				expectRequestCredentialsMode(fetchMock, 'same-origin', 1);
				expectRequestBody(fetchMock, undefined, 1);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toEqual(completedTasks);
				expect(trashedTaskStore.trashedTasks).toEqual(trashedTasks);
			});

			it('applies end date filter', async () => {
				const completedTasks = testData.pastCompletedPartition.tasks;
				const trashedTasks = [testData.trashedPartition.datedDoneTask];
				mockFilterResponses(completedTasks, trashedTasks);

				const page = render(Page);
				await enterFilterDate(page, 'End', todayDate);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
				const queryParameters = `categoryId=7&categoryId=8&endDate=${todayDate.toString()}`;
				expectRequestUrl(fetchMock, `/api/tasks/dated-untrashed-done?${queryParameters}`);
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expectRequestUrl(fetchMock, `/api/tasks/trashed?${queryParameters}`, 1);
				expectRequestMethod(fetchMock, 'GET', 1);
				expectRequestCredentialsMode(fetchMock, 'same-origin', 1);
				expectRequestBody(fetchMock, undefined, 1);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toEqual(completedTasks);
				expect(trashedTaskStore.trashedTasks).toEqual(trashedTasks);
			});

			it('applies date range filter', async () => {
				const completedTasks = testData.pastCompletedPartition.tasks;
				const trashedTasks = [testData.trashedPartition.datedDoneTask];
				mockFilterResponses(completedTasks, trashedTasks);

				const page = render(Page);
				await enterFilterDate(page, 'Start', testData.pastCompletedPartitionDate);
				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
				fetchMock.mockClear();
				mockFilterResponses(completedTasks, trashedTasks);
				await enterFilterDate(page, 'End', todayDate);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
				const queryParameters = `categoryId=7&categoryId=8&startDate=${testData.pastCompletedPartitionDate.toString()}&endDate=${todayDate.toString()}`;
				expectRequestUrl(fetchMock, `/api/tasks/dated-untrashed-done?${queryParameters}`);
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expectRequestUrl(fetchMock, `/api/tasks/trashed?${queryParameters}`, 1);
				expectRequestMethod(fetchMock, 'GET', 1);
				expectRequestCredentialsMode(fetchMock, 'same-origin', 1);
				expectRequestBody(fetchMock, undefined, 1);
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toEqual(completedTasks);
				expect(trashedTaskStore.trashedTasks).toEqual(trashedTasks);
			});
		});

		describe('infinite scroll', () => {
			it('does not enable infinite scrolling for exhausted partitions', () => {
				const page = render(Page);
				const completedColumn = getTaskColumn(page, 'Completed');
				const trashedColumn = page.getByLabelText('Trashed tasks').closest('section');
				if (!(trashedColumn instanceof HTMLElement)) throw new Error('Expected the trashed task column.');

				expect(completedColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(trashedColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(fetchMock).not.toHaveBeenCalled();
			});
		});

		describe('drag and drop', () => {
			describe('within untrashed task partition', () => {
				it('moves a task to the first index', async () => {
					const expectedRank = generateKeyBetween(undefined, testData.pastCompletedPartition.firstTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const taskList = page.getByLabelText(`Completed tasks dated ${testData.pastCompletedPartitionLabel}`);
					const reorderedTasks = [
						testData.pastCompletedPartition.thirdTask,
						testData.pastCompletedPartition.firstTask,
						testData.pastCompletedPartition.secondTask
					];
					await dropTask(taskList, reorderedTasks, testData.pastCompletedPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.pastCompletedPartitionDate.toString(),
						done: true,
						trashed: false
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.pastCompletedPartition.thirdTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(taskList, reorderedTasks);
				});

				it('moves a task to the middle index', async () => {
					const expectedRank = generateKeyBetween(
						testData.pastCompletedPartition.firstTask.rank,
						testData.pastCompletedPartition.secondTask.rank
					);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const taskList = page.getByLabelText(`Completed tasks dated ${testData.pastCompletedPartitionLabel}`);
					const reorderedTasks = [
						testData.pastCompletedPartition.firstTask,
						testData.pastCompletedPartition.thirdTask,
						testData.pastCompletedPartition.secondTask
					];
					await dropTask(taskList, reorderedTasks, testData.pastCompletedPartition.thirdTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.pastCompletedPartitionDate.toString(),
						done: true,
						trashed: false,
						previousTaskId: testData.pastCompletedPartition.firstTask.id
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.pastCompletedPartition.thirdTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(taskList, reorderedTasks);
				});

				it('moves a task to the last index', async () => {
					const expectedRank = generateKeyBetween(testData.pastCompletedPartition.thirdTask.rank, undefined);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const taskList = page.getByLabelText(`Completed tasks dated ${testData.pastCompletedPartitionLabel}`);
					const reorderedTasks = [
						testData.pastCompletedPartition.secondTask,
						testData.pastCompletedPartition.thirdTask,
						testData.pastCompletedPartition.firstTask
					];
					await dropTask(taskList, reorderedTasks, testData.pastCompletedPartition.firstTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.pastCompletedPartitionDate.toString(),
						done: true,
						trashed: false,
						previousTaskId: testData.pastCompletedPartition.thirdTask.id
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.pastCompletedPartition.firstTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(taskList, reorderedTasks);
				});
			});

			describe('within trashed task partition', () => {
				it('moves a task to the first index', async () => {
					const expectedRank = generateKeyBetween(undefined, testData.trashedPartition.undatedOpenTask.rank);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const taskList = page.getByLabelText('Trashed tasks');
					const reorderedTasks = [
						testData.trashedPartition.datedDoneTask,
						testData.trashedPartition.undatedOpenTask,
						testData.trashedPartition.datedOpenTask
					];
					await dropTask(taskList, reorderedTasks, testData.trashedPartition.datedDoneTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.trashedPartition.datedDoneTask.date?.toString(),
						done: true,
						trashed: true
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.trashedPartition.datedDoneTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(taskList, reorderedTasks);
				});

				it('moves a task to the middle index', async () => {
					const expectedRank = generateKeyBetween(
						testData.trashedPartition.undatedOpenTask.rank,
						testData.trashedPartition.datedOpenTask.rank
					);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const taskList = page.getByLabelText('Trashed tasks');
					const reorderedTasks = [
						testData.trashedPartition.undatedOpenTask,
						testData.trashedPartition.datedDoneTask,
						testData.trashedPartition.datedOpenTask
					];
					await dropTask(taskList, reorderedTasks, testData.trashedPartition.datedDoneTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						date: testData.trashedPartition.datedDoneTask.date?.toString(),
						done: true,
						trashed: true,
						previousTaskId: testData.trashedPartition.undatedOpenTask.id
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.trashedPartition.datedDoneTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(taskList, reorderedTasks);
				});

				it('moves a task to the last index', async () => {
					const expectedRank = generateKeyBetween(testData.trashedPartition.datedDoneTask.rank, undefined);
					const response = getResponse({ rank: expectedRank });
					fetchMock.mockResolvedValue(response);

					const page = render(Page);
					const taskList = page.getByLabelText('Trashed tasks');
					const reorderedTasks = [
						testData.trashedPartition.datedOpenTask,
						testData.trashedPartition.datedDoneTask,
						testData.trashedPartition.undatedOpenTask
					];
					await dropTask(taskList, reorderedTasks, testData.trashedPartition.undatedOpenTask.id);

					await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
					const requestBody = JSON.stringify({
						done: false,
						trashed: true,
						previousTaskId: testData.trashedPartition.datedDoneTask.id
					});
					expectRequestUrl(fetchMock, `/api/tasks/${testData.trashedPartition.undatedOpenTask.id}/drag-and-drop`);
					expectRequestMethod(fetchMock, 'PATCH');
					expectRequestCredentialsMode(fetchMock, 'same-origin');
					expectRequestBody(fetchMock, requestBody);
					expectRequestContentType(fetchMock, 'application/json');
					expectTaskOrder(taskList, reorderedTasks);
				});
			});

			describe('from trashed task partition', () => {
				describe('to untrashed task partition', () => {
					it('moves a task to the first index', async () => {
						const expectedRank = generateKeyBetween(undefined, testData.futureCompletedPartition.firstTask.rank);
						const response = getResponse({ rank: expectedRank });
						const trashedTasksExistResponse = getResponse({ trashedTasksExist: true });
						fetchMock.mockResolvedValueOnce(response).mockResolvedValueOnce(trashedTasksExistResponse);

						const page = render(Page);
						const sourceTaskList = page.getByLabelText('Trashed tasks');
						const destinationTaskList = page.getByLabelText(`Completed tasks dated ${testData.futureCompletedPartitionLabel}`);
						const reorderedTasks = [testData.trashedPartition.datedDoneTask, ...testData.futureCompletedPartition.tasks];
						await dropTask(destinationTaskList, reorderedTasks, testData.trashedPartition.datedDoneTask.id);

						await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
						const requestBody = JSON.stringify({
							date: testData.futureCompletedPartitionDate.toString(),
							done: true,
							trashed: false
						});
						expectRequestUrl(fetchMock, `/api/tasks/${testData.trashedPartition.datedDoneTask.id}/drag-and-drop`);
						expectRequestMethod(fetchMock, 'PATCH');
						expectRequestCredentialsMode(fetchMock, 'same-origin');
						expectRequestBody(fetchMock, requestBody);
						expectRequestContentType(fetchMock, 'application/json');
						expectRequestUrl(fetchMock, '/api/tasks/trashed/exists', 1);
						expectRequestMethod(fetchMock, 'GET', 1);
						expectRequestCredentialsMode(fetchMock, 'same-origin', 1);
						expectRequestBody(fetchMock, undefined, 1);
						expectTaskOrder(destinationTaskList, reorderedTasks);
						await waitFor(() => expectTaskNotToBePresent(sourceTaskList, testData.trashedPartition.datedDoneTask.title));
					});

					it('moves a task to the middle index', async () => {
						const expectedRank = generateKeyBetween(
							testData.futureCompletedPartition.firstTask.rank,
							testData.futureCompletedPartition.secondTask.rank
						);
						const response = getResponse({ rank: expectedRank });
						const trashedTasksExistResponse = getResponse({ trashedTasksExist: true });
						fetchMock.mockResolvedValueOnce(response).mockResolvedValueOnce(trashedTasksExistResponse);

						const page = render(Page);
						const sourceTaskList = page.getByLabelText('Trashed tasks');
						const destinationTaskList = page.getByLabelText(`Completed tasks dated ${testData.futureCompletedPartitionLabel}`);
						const reorderedTasks = [
							testData.futureCompletedPartition.firstTask,
							testData.trashedPartition.datedDoneTask,
							testData.futureCompletedPartition.secondTask,
							testData.futureCompletedPartition.thirdTask
						];
						await dropTask(destinationTaskList, reorderedTasks, testData.trashedPartition.datedDoneTask.id);

						await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
						const requestBody = JSON.stringify({
							date: testData.futureCompletedPartitionDate.toString(),
							done: true,
							trashed: false,
							previousTaskId: testData.futureCompletedPartition.firstTask.id
						});
						expectRequestUrl(fetchMock, `/api/tasks/${testData.trashedPartition.datedDoneTask.id}/drag-and-drop`);
						expectRequestMethod(fetchMock, 'PATCH');
						expectRequestCredentialsMode(fetchMock, 'same-origin');
						expectRequestBody(fetchMock, requestBody);
						expectRequestContentType(fetchMock, 'application/json');
						expectRequestUrl(fetchMock, '/api/tasks/trashed/exists', 1);
						expectRequestMethod(fetchMock, 'GET', 1);
						expectRequestCredentialsMode(fetchMock, 'same-origin', 1);
						expectRequestBody(fetchMock, undefined, 1);
						expectTaskOrder(destinationTaskList, reorderedTasks);
						await waitFor(() => expectTaskNotToBePresent(sourceTaskList, testData.trashedPartition.datedDoneTask.title));
					});

					it('moves a task to the last index', async () => {
						const expectedRank = generateKeyBetween(testData.futureCompletedPartition.thirdTask.rank, undefined);
						const response = getResponse({ rank: expectedRank });
						const trashedTasksExistResponse = getResponse({ trashedTasksExist: true });
						fetchMock.mockResolvedValueOnce(response).mockResolvedValueOnce(trashedTasksExistResponse);

						const page = render(Page);
						const sourceTaskList = page.getByLabelText('Trashed tasks');
						const destinationTaskList = page.getByLabelText(`Completed tasks dated ${testData.futureCompletedPartitionLabel}`);
						const reorderedTasks = [...testData.futureCompletedPartition.tasks, testData.trashedPartition.datedDoneTask];
						await dropTask(destinationTaskList, reorderedTasks, testData.trashedPartition.datedDoneTask.id);

						await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
						const requestBody = JSON.stringify({
							date: testData.futureCompletedPartitionDate.toString(),
							done: true,
							trashed: false,
							previousTaskId: testData.futureCompletedPartition.thirdTask.id
						});
						expectRequestUrl(fetchMock, `/api/tasks/${testData.trashedPartition.datedDoneTask.id}/drag-and-drop`);
						expectRequestMethod(fetchMock, 'PATCH');
						expectRequestCredentialsMode(fetchMock, 'same-origin');
						expectRequestBody(fetchMock, requestBody);
						expectRequestContentType(fetchMock, 'application/json');
						expectRequestUrl(fetchMock, '/api/tasks/trashed/exists', 1);
						expectRequestMethod(fetchMock, 'GET', 1);
						expectRequestCredentialsMode(fetchMock, 'same-origin', 1);
						expectRequestBody(fetchMock, undefined, 1);
						expectTaskOrder(destinationTaskList, reorderedTasks);
						await waitFor(() => expectTaskNotToBePresent(sourceTaskList, testData.trashedPartition.datedDoneTask.title));
					});
				});
			});

			describe('from untrashed task partition', () => {
				describe('to untrashed task partition', () => {
					it('moves a task to the first index', async () => {
						const expectedRank = generateKeyBetween(undefined, testData.futureCompletedPartition.firstTask.rank);
						const response = getResponse({ rank: expectedRank });
						fetchMock.mockResolvedValue(response);

						const page = render(Page);
						const sourceTaskList = page.getByLabelText(`Completed tasks dated ${testData.pastCompletedPartitionLabel}`);
						const destinationTaskList = page.getByLabelText(`Completed tasks dated ${testData.futureCompletedPartitionLabel}`);
						const reorderedTasks = [testData.pastCompletedPartition.thirdTask, ...testData.futureCompletedPartition.tasks];
						await dropTask(destinationTaskList, reorderedTasks, testData.pastCompletedPartition.thirdTask.id);

						await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
						const requestBody = JSON.stringify({
							date: testData.futureCompletedPartitionDate.toString(),
							done: true,
							trashed: false
						});
						expectRequestUrl(fetchMock, `/api/tasks/${testData.pastCompletedPartition.thirdTask.id}/drag-and-drop`);
						expectRequestMethod(fetchMock, 'PATCH');
						expectRequestCredentialsMode(fetchMock, 'same-origin');
						expectRequestBody(fetchMock, requestBody);
						expectRequestContentType(fetchMock, 'application/json');
						expectTaskOrder(destinationTaskList, reorderedTasks);
						await waitFor(() => expectTaskNotToBePresent(sourceTaskList, testData.pastCompletedPartition.thirdTask.title));
					});

					it('moves a task to the middle index', async () => {
						const expectedRank = generateKeyBetween(
							testData.futureCompletedPartition.firstTask.rank,
							testData.futureCompletedPartition.secondTask.rank
						);
						const response = getResponse({ rank: expectedRank });
						fetchMock.mockResolvedValue(response);

						const page = render(Page);
						const sourceTaskList = page.getByLabelText(`Completed tasks dated ${testData.pastCompletedPartitionLabel}`);
						const destinationTaskList = page.getByLabelText(`Completed tasks dated ${testData.futureCompletedPartitionLabel}`);
						const reorderedTasks = [
							testData.futureCompletedPartition.firstTask,
							testData.pastCompletedPartition.thirdTask,
							testData.futureCompletedPartition.secondTask,
							testData.futureCompletedPartition.thirdTask
						];
						await dropTask(destinationTaskList, reorderedTasks, testData.pastCompletedPartition.thirdTask.id);

						await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
						const requestBody = JSON.stringify({
							date: testData.futureCompletedPartitionDate.toString(),
							done: true,
							trashed: false,
							previousTaskId: testData.futureCompletedPartition.firstTask.id
						});
						expectRequestUrl(fetchMock, `/api/tasks/${testData.pastCompletedPartition.thirdTask.id}/drag-and-drop`);
						expectRequestMethod(fetchMock, 'PATCH');
						expectRequestCredentialsMode(fetchMock, 'same-origin');
						expectRequestBody(fetchMock, requestBody);
						expectRequestContentType(fetchMock, 'application/json');
						expectTaskOrder(destinationTaskList, reorderedTasks);
						await waitFor(() => expectTaskNotToBePresent(sourceTaskList, testData.pastCompletedPartition.thirdTask.title));
					});

					it('moves a task to the last index', async () => {
						const expectedRank = generateKeyBetween(testData.futureCompletedPartition.thirdTask.rank, undefined);
						const response = getResponse({ rank: expectedRank });
						fetchMock.mockResolvedValue(response);

						const page = render(Page);
						const sourceTaskList = page.getByLabelText(`Completed tasks dated ${testData.pastCompletedPartitionLabel}`);
						const destinationTaskList = page.getByLabelText(`Completed tasks dated ${testData.futureCompletedPartitionLabel}`);
						const reorderedTasks = [...testData.futureCompletedPartition.tasks, testData.pastCompletedPartition.thirdTask];
						await dropTask(destinationTaskList, reorderedTasks, testData.pastCompletedPartition.thirdTask.id);

						await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
						const requestBody = JSON.stringify({
							date: testData.futureCompletedPartitionDate.toString(),
							done: true,
							trashed: false,
							previousTaskId: testData.futureCompletedPartition.thirdTask.id
						});
						expectRequestUrl(fetchMock, `/api/tasks/${testData.pastCompletedPartition.thirdTask.id}/drag-and-drop`);
						expectRequestMethod(fetchMock, 'PATCH');
						expectRequestCredentialsMode(fetchMock, 'same-origin');
						expectRequestBody(fetchMock, requestBody);
						expectRequestContentType(fetchMock, 'application/json');
						expectTaskOrder(destinationTaskList, reorderedTasks);
						await waitFor(() => expectTaskNotToBePresent(sourceTaskList, testData.pastCompletedPartition.thirdTask.title));
					});
				});

				describe('to trashed task partition', () => {
					it('moves a task to the first index', async () => {
						const expectedRank = generateKeyBetween(undefined, testData.trashedPartition.undatedOpenTask.rank);
						const response = getResponse({ rank: expectedRank });
						fetchMock.mockResolvedValue(response);

						const page = render(Page);
						const sourceTaskList = page.getByLabelText(`Completed tasks dated ${testData.pastCompletedPartitionLabel}`);
						const destinationTaskList = page.getByLabelText('Trashed tasks');
						const reorderedTasks = [testData.pastCompletedPartition.thirdTask, ...testData.trashedPartition.tasks];
						await dropTask(destinationTaskList, reorderedTasks, testData.pastCompletedPartition.thirdTask.id);

						await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
						const requestBody = JSON.stringify({
							date: testData.pastCompletedPartitionDate.toString(),
							done: true,
							trashed: true
						});
						expectRequestUrl(fetchMock, `/api/tasks/${testData.pastCompletedPartition.thirdTask.id}/drag-and-drop`);
						expectRequestMethod(fetchMock, 'PATCH');
						expectRequestCredentialsMode(fetchMock, 'same-origin');
						expectRequestBody(fetchMock, requestBody);
						expectRequestContentType(fetchMock, 'application/json');
						expectTaskOrder(destinationTaskList, reorderedTasks);
						await waitFor(() => expectTaskNotToBePresent(sourceTaskList, testData.pastCompletedPartition.thirdTask.title));
					});

					it('moves a task to the middle index', async () => {
						const expectedRank = generateKeyBetween(
							testData.trashedPartition.undatedOpenTask.rank,
							testData.trashedPartition.datedOpenTask.rank
						);
						const response = getResponse({ rank: expectedRank });
						fetchMock.mockResolvedValue(response);

						const page = render(Page);
						const sourceTaskList = page.getByLabelText(`Completed tasks dated ${testData.pastCompletedPartitionLabel}`);
						const destinationTaskList = page.getByLabelText('Trashed tasks');
						const reorderedTasks = [
							testData.trashedPartition.undatedOpenTask,
							testData.pastCompletedPartition.thirdTask,
							testData.trashedPartition.datedOpenTask,
							testData.trashedPartition.datedDoneTask
						];
						await dropTask(destinationTaskList, reorderedTasks, testData.pastCompletedPartition.thirdTask.id);

						await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
						const requestBody = JSON.stringify({
							date: testData.pastCompletedPartitionDate.toString(),
							done: true,
							trashed: true,
							previousTaskId: testData.trashedPartition.undatedOpenTask.id
						});
						expectRequestUrl(fetchMock, `/api/tasks/${testData.pastCompletedPartition.thirdTask.id}/drag-and-drop`);
						expectRequestMethod(fetchMock, 'PATCH');
						expectRequestCredentialsMode(fetchMock, 'same-origin');
						expectRequestBody(fetchMock, requestBody);
						expectRequestContentType(fetchMock, 'application/json');
						expectTaskOrder(destinationTaskList, reorderedTasks);
						await waitFor(() => expectTaskNotToBePresent(sourceTaskList, testData.pastCompletedPartition.thirdTask.title));
					});

					it('moves a task to the last index', async () => {
						const expectedRank = generateKeyBetween(testData.trashedPartition.datedDoneTask.rank, undefined);
						const response = getResponse({ rank: expectedRank });
						fetchMock.mockResolvedValue(response);

						const page = render(Page);
						const sourceTaskList = page.getByLabelText(`Completed tasks dated ${testData.pastCompletedPartitionLabel}`);
						const destinationTaskList = page.getByLabelText('Trashed tasks');
						const reorderedTasks = [...testData.trashedPartition.tasks, testData.pastCompletedPartition.thirdTask];
						await dropTask(destinationTaskList, reorderedTasks, testData.pastCompletedPartition.thirdTask.id);

						await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
						const requestBody = JSON.stringify({
							date: testData.pastCompletedPartitionDate.toString(),
							done: true,
							trashed: true,
							previousTaskId: testData.trashedPartition.datedDoneTask.id
						});
						expectRequestUrl(fetchMock, `/api/tasks/${testData.pastCompletedPartition.thirdTask.id}/drag-and-drop`);
						expectRequestMethod(fetchMock, 'PATCH');
						expectRequestCredentialsMode(fetchMock, 'same-origin');
						expectRequestBody(fetchMock, requestBody);
						expectRequestContentType(fetchMock, 'application/json');
						expectTaskOrder(destinationTaskList, reorderedTasks);
						await waitFor(() => expectTaskNotToBePresent(sourceTaskList, testData.pastCompletedPartition.thirdTask.title));
					});
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
			it('loads more untrashed tasks without applied filters', async () => {
				const nextTaskResponse = {
					id: 13,
					categoryId: workCategory.id,
					title: 'Publish the meeting minutes',
					date: testData.futureCompletedPartitionDate.toString(),
					rank: generateKeyBetween(testData.futureCompletedPartition.thirdTask.rank, undefined)
				};
				const nextTask: DatedUntrashedDoneTaskResponse = {
					...nextTaskResponse,
					date: testData.futureCompletedPartitionDate,
					done: true
				};
				const existingTrashedTasks = [...trashedTaskStore.trashedTasks];
				const resolveFetch = mockPendingFetch(fetchMock);

				const page = render(Page);
				const completedColumn = getTaskColumn(page, 'Completed');
				const loadingSentinel = getLoadingSentinel(completedColumn);
				intersectionObserverMock.enterNode(loadingSentinel);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, '/api/tasks/dated-untrashed-done');
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expect(getRequestSearchParameter(fetchMock, 'cursorDate')).toBe(testData.futureCompletedPartitionDate.toString());
				expect(getRequestSearchParameter(fetchMock, 'cursorRank')).toBe(testData.futureCompletedPartition.thirdTask.rank);
				expect(within(completedColumn).getByLabelText('Loading more items')).toBeInTheDocument();
				expect(within(getTaskColumn(page, 'Trashed')).queryByLabelText('Loading more items')).not.toBeInTheDocument();

				resolveFetch(getResponse({ tasks: [nextTaskResponse], hasMore: false }));

				await waitFor(() => expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toContainEqual(nextTask));
				expect(datedUntrashedDoneTaskStore.hasMore).toBe(false);
				const taskList = page.getByLabelText(`Completed tasks dated ${testData.futureCompletedPartitionLabel}`);
				expectTaskToBePresent(taskList, nextTask.title, nextTask.description, workCategory.title);
				expect(completedColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(trashedTaskStore.trashedTasks).toEqual(existingTrashedTasks);
			});

			it('loads more untrashed tasks with applied filters', async () => {
				const filteredCompletedTasks = [testData.futureCompletedPartition.firstTask, testData.futureCompletedPartition.thirdTask];
				const filteredTrashedTasks = [testData.trashedPartition.datedOpenTask];
				mockFilterResponses(filteredCompletedTasks, filteredTrashedTasks, true);
				mockFilterResponses(filteredCompletedTasks, filteredTrashedTasks, true);
				mockFilterResponses(filteredCompletedTasks, filteredTrashedTasks, true);

				const page = render(Page);
				const excludePersonalButtons = page.getAllByRole('button', { name: 'Exclude Personal' });
				await userEvent.click(excludePersonalButtons[0]!);
				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
				await enterFilterDate(page, 'Start', todayDate);
				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(4));
				await enterFilterDate(page, 'End', testData.futureCompletedPartitionDate);
				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(6));
				fetchMock.mockClear();

				const nextTaskResponse = {
					id: 13,
					categoryId: workCategory.id,
					title: 'Publish the filtered meeting minutes',
					date: testData.futureCompletedPartitionDate.toString(),
					rank: generateKeyBetween(testData.futureCompletedPartition.thirdTask.rank, undefined)
				};
				const nextTask: DatedUntrashedDoneTaskResponse = {
					...nextTaskResponse,
					date: testData.futureCompletedPartitionDate,
					done: true
				};
				const resolveFetch = mockPendingFetch(fetchMock);
				const completedColumn = getTaskColumn(page, 'Completed');
				intersectionObserverMock.enterNode(getLoadingSentinel(completedColumn));

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, '/api/tasks/dated-untrashed-done');
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expect(getRequestSearchParameter(fetchMock, 'categoryId')).toBe(String(workCategory.id));
				expect(getRequestSearchParameter(fetchMock, 'startDate')).toBe(todayDate.toString());
				expect(getRequestSearchParameter(fetchMock, 'endDate')).toBe(testData.futureCompletedPartitionDate.toString());
				expect(getRequestSearchParameter(fetchMock, 'cursorDate')).toBe(testData.futureCompletedPartitionDate.toString());
				expect(getRequestSearchParameter(fetchMock, 'cursorRank')).toBe(testData.futureCompletedPartition.thirdTask.rank);

				resolveFetch(getResponse({ tasks: [nextTaskResponse], hasMore: false }));

				await waitFor(() => expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toContainEqual(nextTask));
				expect(datedUntrashedDoneTaskStore.hasMore).toBe(false);
				expect(completedColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
			});

			it('loads more trashed tasks without applied filters', async () => {
				const nextTaskResponse = {
					id: 13,
					categoryId: personalCategory.id,
					title: 'Discard obsolete project notes',
					done: false,
					rank: generateKeyBetween(testData.trashedPartition.datedDoneTask.rank, undefined)
				};
				const nextTask: TrashedTaskResponse = nextTaskResponse;
				const existingCompletedTasks = [...datedUntrashedDoneTaskStore.datedUntrashedDoneTasks];
				const resolveFetch = mockPendingFetch(fetchMock);

				const page = render(Page);
				const trashedColumn = getTaskColumn(page, 'Trashed');
				const loadingSentinel = getLoadingSentinel(trashedColumn);
				intersectionObserverMock.enterNode(loadingSentinel);

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, '/api/tasks/trashed');
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expect(getRequestSearchParameter(fetchMock, 'rank')).toBe(testData.trashedPartition.datedDoneTask.rank);
				expect(within(trashedColumn).getByLabelText('Loading more items')).toBeInTheDocument();
				expect(within(getTaskColumn(page, 'Completed')).queryByLabelText('Loading more items')).not.toBeInTheDocument();

				resolveFetch(getResponse({ tasks: [nextTaskResponse], hasMore: false }));

				await waitFor(() => expect(trashedTaskStore.trashedTasks).toContainEqual(nextTask));
				expect(trashedTaskStore.hasMore).toBe(false);
				expectTaskToBePresent(trashedColumn, nextTask.title, nextTask.description, personalCategory.title);
				expect(trashedColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
				expect(datedUntrashedDoneTaskStore.datedUntrashedDoneTasks).toEqual(existingCompletedTasks);
			});

			it('loads more trashed tasks with applied filters', async () => {
				const filteredCompletedTasks = [testData.futureCompletedPartition.firstTask, testData.futureCompletedPartition.thirdTask];
				const filteredTrashedTasks = [testData.trashedPartition.datedOpenTask];
				mockFilterResponses(filteredCompletedTasks, filteredTrashedTasks, true);
				mockFilterResponses(filteredCompletedTasks, filteredTrashedTasks, true);
				mockFilterResponses(filteredCompletedTasks, filteredTrashedTasks, true);

				const page = render(Page);
				const excludePersonalButtons = page.getAllByRole('button', { name: 'Exclude Personal' });
				await userEvent.click(excludePersonalButtons[0]!);
				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
				await enterFilterDate(page, 'Start', todayDate);
				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(4));
				await enterFilterDate(page, 'End', testData.futureCompletedPartitionDate);
				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(6));
				fetchMock.mockClear();

				const nextTaskResponse = {
					id: 13,
					categoryId: workCategory.id,
					title: 'Discard filtered project notes',
					done: false,
					rank: generateKeyBetween(testData.trashedPartition.datedOpenTask.rank, undefined)
				};
				const resolveFetch = mockPendingFetch(fetchMock);
				const trashedColumn = getTaskColumn(page, 'Trashed');
				intersectionObserverMock.enterNode(getLoadingSentinel(trashedColumn));

				await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
				expectRequestPath(fetchMock, '/api/tasks/trashed');
				expectRequestMethod(fetchMock, 'GET');
				expectRequestCredentialsMode(fetchMock, 'same-origin');
				expectRequestBody(fetchMock, undefined);
				expect(getRequestSearchParameter(fetchMock, 'categoryId')).toBe(String(workCategory.id));
				expect(getRequestSearchParameter(fetchMock, 'startDate')).toBe(todayDate.toString());
				expect(getRequestSearchParameter(fetchMock, 'endDate')).toBe(testData.futureCompletedPartitionDate.toString());
				expect(getRequestSearchParameter(fetchMock, 'rank')).toBe(testData.trashedPartition.datedOpenTask.rank);

				resolveFetch(getResponse({ tasks: [nextTaskResponse], hasMore: false }));

				await waitFor(() => expect(trashedTaskStore.trashedTasks).toContainEqual(nextTaskResponse));
				expect(trashedTaskStore.hasMore).toBe(false);
				expect(trashedColumn.querySelector('.loading-sentinel')).not.toBeInTheDocument();
			});

			it('disables infinite scrolling while a task is being dragged', async () => {
				const page = render(Page);
				const completedColumn = getTaskColumn(page, 'Completed');
				const trashedColumn = getTaskColumn(page, 'Trashed');
				const completedLoadingSentinel = getLoadingSentinel(completedColumn);
				const trashedLoadingSentinel = getLoadingSentinel(trashedColumn);

				const completedTaskList = page.getByLabelText(`Completed tasks dated ${testData.pastCompletedPartitionLabel}`);
				const event = new CustomEvent<DndEvent<DatedUntrashedDoneTaskResponse>>('consider', {
					detail: {
						items: testData.pastCompletedPartition.tasks,
						info: {
							trigger: TRIGGERS.DRAG_STARTED,
							id: String(testData.pastCompletedPartition.firstTask.id),
							source: SOURCES.POINTER
						}
					}
				});
				await fireEvent(completedTaskList, event);

				intersectionObserverMock.enterNodes([completedLoadingSentinel, trashedLoadingSentinel]);
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
	const pastDate = todayDate.subtract({ days: 1 });
	const futureDate = todayDate.add({ days: 1 });
	const pastDateLabel = `${String(pastDate.day).padStart(2, '0')}.${String(pastDate.month).padStart(2, '0')}.${pastDate.year}`;
	const futureDateLabel = `${String(futureDate.day).padStart(2, '0')}.${String(futureDate.month).padStart(2, '0')}.${futureDate.year}`;
	let taskId = 1;

	const pastCompletedTaskDetails = [
		{
			categoryId: personalCategory.id,
			title: 'Send the signed contract',
			description: 'Attach the final PDF and copy the project lead.'
		},
		{ categoryId: workCategory.id, title: 'Book the conference hotel' },
		{
			categoryId: personalCategory.id,
			title: 'Submit the travel expenses',
			description: 'Include the train and taxi receipts.'
		}
	] as const;
	const pastCompletedTaskRanks = generateNKeysBetween(undefined, undefined, pastCompletedTaskDetails.length);
	const pastCompletedTasks: DatedUntrashedDoneTaskResponse[] = pastCompletedTaskDetails.map((details, index) => ({
		...details,
		id: taskId++,
		date: pastDate,
		rank: pastCompletedTaskRanks[index],
		done: true
	}));

	const futureCompletedTaskDetails = [
		{ categoryId: workCategory.id, title: 'Prepare the workshop agenda' },
		{
			categoryId: personalCategory.id,
			title: 'Review the quarterly report',
			description: 'Check the revenue figures and highlighted risks.'
		},
		{ categoryId: workCategory.id, title: 'Confirm the catering order' }
	] as const;
	const futureCompletedTaskRanks = generateNKeysBetween(undefined, undefined, futureCompletedTaskDetails.length);
	const futureCompletedTasks: DatedUntrashedDoneTaskResponse[] = futureCompletedTaskDetails.map((details, index) => ({
		...details,
		id: taskId++,
		date: futureDate,
		rank: futureCompletedTaskRanks[index],
		done: true
	}));

	const trashedTaskRanks = generateNKeysBetween(undefined, undefined, 3);
	const undatedOpenTrashedTask: TrashedTaskResponse = {
		id: taskId++,
		categoryId: personalCategory.id,
		title: 'Compare mobile phone plans',
		description: 'Check monthly costs and available data packages.',
		done: false,
		rank: trashedTaskRanks[0]
	};
	const datedOpenTrashedTask: TrashedTaskResponse = {
		id: taskId++,
		categoryId: workCategory.id,
		title: 'Schedule the annual checkup',
		date: futureDate,
		done: false,
		rank: trashedTaskRanks[1]
	};
	const datedDoneTrashedTask: TrashedTaskResponse = {
		id: taskId++,
		categoryId: personalCategory.id,
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
		categoryId: personalCategory.id,
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

async function dropTask(
	taskList: HTMLElement,
	newOrderedTasks: (DatedUntrashedDoneTaskResponse | TrashedTaskResponse)[],
	draggedTaskId: number
): Promise<void> {
	const event = new CustomEvent<DndEvent<DatedUntrashedDoneTaskResponse | TrashedTaskResponse>>('finalize', {
		detail: {
			items: newOrderedTasks,
			info: { trigger: TRIGGERS.DROPPED_INTO_ZONE, id: String(draggedTaskId), source: SOURCES.POINTER }
		}
	});
	await fireEvent(taskList, event);
}

function expectTaskOrder(taskList: HTMLElement, tasks: (DatedUntrashedDoneTaskResponse | TrashedTaskResponse)[]): void {
	const taskListContent = within(taskList);
	const renderedTasks = taskListContent.getAllByRole('listitem');
	const renderedTaskTitles = renderedTasks.map((task) => task.getAttribute('aria-label'));
	const expectedTaskTitles = tasks.map((task) => task.title);
	expect(renderedTaskTitles).toEqual(expectedTaskTitles);
}

function expectTaskNotToBePresent(taskList: HTMLElement, title: string): void {
	const task = within(taskList).queryByRole('listitem', { name: title });
	expect(task).not.toBeInTheDocument();
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

function mockFilterResponses(
	completedTasks: DatedUntrashedDoneTaskResponse[],
	trashedTasks: TrashedTaskResponse[],
	hasMore = false
): void {
	fetchMock
		.mockResolvedValueOnce(
			getResponse({
				tasks: completedTasks.map((task) => ({
					id: task.id,
					categoryId: task.categoryId,
					title: task.title,
					description: task.description,
					date: task.date.toString(),
					rank: task.rank
				})),
				hasMore
			})
		)
		.mockResolvedValueOnce(
			getResponse({
				tasks: trashedTasks.map((task) => ({
					id: task.id,
					categoryId: task.categoryId,
					title: task.title,
					description: task.description,
					date: task.date?.toString(),
					rank: task.rank,
					done: task.done
				})),
				hasMore
			})
		);
}

async function enterFilterDate(page: RenderResult<typeof Page>, label: 'Start' | 'End', date: DateValue): Promise<void> {
	await fireEvent.input(page.getByRole('textbox', { name: `${label} day` }), {
		target: { value: String(date.day).padStart(2, '0') }
	});
	await fireEvent.input(page.getByRole('textbox', { name: `${label} month` }), {
		target: { value: String(date.month).padStart(2, '0') }
	});
	await fireEvent.input(page.getByRole('textbox', { name: `${label} year` }), { target: { value: String(date.year) } });
}
