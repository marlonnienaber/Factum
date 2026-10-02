import { categoryStore } from '$lib/stores/category-store.svelte';
import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
import { trashedTaskStore } from '$lib/stores/trashed-task-store.svelte';
import type { DatedUntrashedDoneTaskResponse } from '$lib/types/responses/dated-untrashed-done-task-response';
import type { TrashedTaskResponse } from '$lib/types/responses/trashed-task-response';
import { getLocalTimeZone, today } from '@internationalized/date';
import {generateNKeysBetween} from 'fractional-indexing';
import {beforeEach, describe, it} from 'vitest';

describe('archive page tests', () => {
	beforeEach(() => {
		datedUntrashedDoneTaskStore.reset();
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
			it.todo('successfully trashes all untrashed tasks');

			it.todo('successfully opens all untrashed tasks');

			it.todo('successfully cancels a batch operation');

			it.todo('disables batch actions when no tasks are selected');

			it.todo('disables batch open when more than 50 tasks are selected');

			it.todo('disables batch trash when more than 50 tasks are selected');
		});

		describe('task operations', () => {
			it.todo('edits untrashed task');

			it.todo('edits trashed task');

			it.todo('trashes task');

			it.todo('restores task');
		});

		describe('empty trash', () => {
			it.todo('successfully empties trash without applied filters');

			it.todo('successfully empties trash with applied filters');

			it.todo('hides empty trash if there are no trashed tasks');
		});

		it.todo('shows trash info popover on info icon hover');

		describe('filtering', () => {
			it.todo('applies category filter');

			it.todo('applies start date filter');

			it.todo('applies end date filter');

			it.todo('applies date range filter');
		});

		describe('infinite scroll', () => {
			it.todo('does not enable infinite scrolling for exhausted partitions');
		});

		describe('drag and drop', () => {
			describe('within untrashed task partition', () => {
				it.todo('moves a task to the first index');

				it.todo('moves a task to the middle index');

				it.todo('moves a task to the last index');
			});

			describe('within trashed task partition', () => {
				it.todo('moves a task to the first index');

				it.todo('moves a task to the middle index');

				it.todo('moves a task to the last index');
			});

			describe('from trashed task partition', () => {
				describe('to other trashed task partition', () => {
					it.todo('moves a task to the first index');

					it.todo('moves a task to the middle index');

					it.todo('moves a task to the last index');
				});

				describe('to untrashed task partition', () => {
					it.todo('moves a task to the first index');

					it.todo('moves a task to the middle index');

					it.todo('moves a task to the last index');
				});
			});

			describe('from untrashed task partition', () => {
				describe('to untrashed task partition', () => {
					it.todo('moves a task to the first index');

					it.todo('moves a task to the middle index');

					it.todo('moves a task to the last index');
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
			it.todo('loads more untrashed tasks without applied filters');

			it.todo('loads more untrashed tasks with applied filters');

			it.todo('loads more trashed tasks without applied filters');

			it.todo('loads more trashed tasks with applied filters');

			it.todo('disables infinite scrolling while a task is being dragged');
		});
	});
});

const categories = [
	{ id: 7, title: 'Personal', color: '#22c55e' },
	{ id: 8, title: 'Work', color: '#3b82f6' }
];
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
	let taskId = 1;

	const pastCompletedTaskDetails = [
		{ title: 'Send the signed contract', description: 'Attach the final PDF and copy the project lead.' },
		{ title: 'Book the conference hotel' },
		{ title: 'Submit the travel expenses', description: 'Include the train and taxi receipts.' }
	] as const;
	const pastCompletedTaskRanks = generateNKeysBetween(undefined, undefined, pastCompletedTaskDetails.length);
	const pastCompletedTasks: DatedUntrashedDoneTaskResponse[] = pastCompletedTaskDetails.map((details, index) => ({
		...details,
		id: taskId++,
		categoryId: categories[index % categories.length]!.id,
		date: pastDate,
		rank: pastCompletedTaskRanks[index],
		done: true
	}));

	const futureCompletedTaskDetails = [
		{ title: 'Prepare the workshop agenda' },
		{ title: 'Review the quarterly report', description: 'Check the revenue figures and highlighted risks.' },
		{ title: 'Confirm the catering order' }
	] as const;
	const futureCompletedTaskRanks = generateNKeysBetween(undefined, undefined, futureCompletedTaskDetails.length);
	const futureCompletedTasks: DatedUntrashedDoneTaskResponse[] = futureCompletedTaskDetails.map((details, index) => ({
		...details,
		id: taskId++,
		categoryId: categories[(index + 1) % categories.length]!.id,
		date: futureDate,
		rank: futureCompletedTaskRanks[index],
		done: true
	}));

	const trashedTaskRanks = generateNKeysBetween(undefined, undefined, 3);
	const undatedOpenTrashedTask: TrashedTaskResponse = {
		id: taskId++,
		categoryId: categories[0]!.id,
		title: 'Compare mobile phone plans',
		description: 'Check monthly costs and available data packages.',
		done: false,
		rank: trashedTaskRanks[0]
	};
	const datedOpenTrashedTask: TrashedTaskResponse = {
		id: taskId++,
		categoryId: categories[1]!.id,
		title: 'Schedule the annual checkup',
		date: futureDate,
		done: false,
		rank: trashedTaskRanks[1]
	};
	const datedDoneTrashedTask: TrashedTaskResponse = {
		id: taskId++,
		categoryId: categories[0]!.id,
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
		futureCompletedPartition,
		futureCompletedPartitionDate: futureDate,
		trashedPartition: {
			undatedOpenTask: undatedOpenTrashedTask,
			datedOpenTask: datedOpenTrashedTask,
			datedDoneTask: datedDoneTrashedTask,
			tasks: trashedTasks
		}
	};
}
