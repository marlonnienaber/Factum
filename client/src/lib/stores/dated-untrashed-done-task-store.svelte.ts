import type { DatedUntrashedDoneTaskResponse } from '$lib/types/responses/dated-untrashed-done-task-response';

let datedUntrashedDoneTasks: DatedUntrashedDoneTaskResponse[] = $state<DatedUntrashedDoneTaskResponse[]>([]);

function addOrUpdateTasks(tasks: DatedUntrashedDoneTaskResponse[]): void {
	const taskIds = new Set(tasks.map((task) => task.id));
	const deduplicatedTasks: DatedUntrashedDoneTaskResponse[] = datedUntrashedDoneTasks.filter((task) => !taskIds.has(task.id));
	datedUntrashedDoneTasks = [...deduplicatedTasks, ...tasks].sort(compareTasksByDateAndRank);
	console.log('Dated untrashed done tasks:', $state.snapshot(datedUntrashedDoneTasks));
}

function addOrUpdateTask(task: DatedUntrashedDoneTaskResponse): void {
	addOrUpdateTasks([task]);
}

function reset(): void {
	datedUntrashedDoneTasks = [];
}

function compareTasksByDateAndRank(first: DatedUntrashedDoneTaskResponse, second: DatedUntrashedDoneTaskResponse): number {
	const dateComparison = first.date.compare(second.date);
	if (dateComparison !== 0) return dateComparison;
	return first.rank < second.rank ? -1 : first.rank > second.rank ? 1 : 0;
}

export const datedUntrashedDoneTaskStore = {
	get datedUntrashedDoneTasks(): DatedUntrashedDoneTaskResponse[] {
		return datedUntrashedDoneTasks;
	},
	addOrUpdateTasks,
	addOrUpdateTask,
	reset
};
