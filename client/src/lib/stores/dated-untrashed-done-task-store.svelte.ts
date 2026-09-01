import type { DatedUntrashedDoneTaskResponse } from '$lib/types/responses/dated-untrashed-done-task-response';

let datedUntrashedDoneTasks: DatedUntrashedDoneTaskResponse[] = $state<DatedUntrashedDoneTaskResponse[]>([]);

function addDatedUntrashedDoneTask(task: DatedUntrashedDoneTaskResponse): void {
	const deduplicatedTasks = datedUntrashedDoneTasks.filter((existingTask) => existingTask.id !== task.id);
	datedUntrashedDoneTasks = [...deduplicatedTasks, task].sort(compareTasksByDateAndRank);
	console.log('Dated untrashed done tasks:', $state.snapshot(datedUntrashedDoneTasks));
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
	addDatedUntrashedDoneTask
};
