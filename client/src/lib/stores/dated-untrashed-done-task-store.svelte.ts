import type { DatedUntrashedDoneTaskResponse } from '$lib/types/responses/dated-untrashed-done-task-response';

let datedUntrashedDoneTasks: DatedUntrashedDoneTaskResponse[] = $state<DatedUntrashedDoneTaskResponse[]>([]);
let hasMore = $state(false);

function addOrUpdateTasks(tasks: DatedUntrashedDoneTaskResponse[]): void {
	const taskIds = new Set(tasks.map((task) => task.id));
	const deduplicatedTasks: DatedUntrashedDoneTaskResponse[] = datedUntrashedDoneTasks.filter((task) => !taskIds.has(task.id));
	datedUntrashedDoneTasks = [...deduplicatedTasks, ...tasks].sort(compareTasksByDateAndRank);
}

function addOrUpdateTask(task: DatedUntrashedDoneTaskResponse): void {
	addOrUpdateTasks([task]);
}

function addPage(tasks: DatedUntrashedDoneTaskResponse[], pageHasMore: boolean): void {
	addOrUpdateTasks(tasks);
	hasMore = pageHasMore;
}

function removeTask(id: number): void {
	datedUntrashedDoneTasks = datedUntrashedDoneTasks.filter((task) => task.id !== id);
}

function reset(): void {
	datedUntrashedDoneTasks = [];
	hasMore = false;
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
	get hasMore(): boolean {
		return hasMore;
	},
	addOrUpdateTasks,
	addOrUpdateTask,
	addPage,
	removeTask,
	reset
};
