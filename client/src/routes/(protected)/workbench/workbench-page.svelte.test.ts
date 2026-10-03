import { getLocalTimeZone, today } from '@internationalized/date';
import {generateNKeysBetween} from 'fractional-indexing';
import { categoryStore } from '$lib/stores/category-store.svelte';
import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
import { todayProgressStore } from '$lib/stores/today-progress-store.svelte';
import { trashedTaskStore } from '$lib/stores/trashed-task-store.svelte';
import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
import type { DatedUntrashedDoneTaskResponse } from '$lib/types/responses/dated-untrashed-done-task-response';
import type { DatedUntrashedOpenTaskResponse } from '$lib/types/responses/dated-untrashed-open-task-response';
import {beforeEach, describe, it, vi} from 'vitest';

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
			it.todo('successfully trashes all tasks');

			it.todo('successfully completes all open tasks');

			it.todo('successfully opens all done tasks');

			it.todo('disables batch opening if an open task is selected');

			it.todo('disables batch completion if a done task is selected');

			it.todo('disables batch operations when more than 50 tasks are selected');

			it.todo('disables batch operations when no task is selected');

			it.todo('successfully cancels a batch operation');
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
