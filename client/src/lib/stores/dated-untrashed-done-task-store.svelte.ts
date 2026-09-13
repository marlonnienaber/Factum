import type { DatedUntrashedTaskResponse } from '$lib/types/responses/dated-untrashed-task-response';

let datedUntrashedDoneTasks: DatedUntrashedTaskResponse[] = $state<DatedUntrashedTaskResponse[]>([]);
let hasMore = $state(false);

function addOrUpdateTasks(tasks: DatedUntrashedTaskResponse[]): void {
	const taskIds = new Set(tasks.map((task) => task.id));
	const deduplicatedTasks: DatedUntrashedTaskResponse[] = datedUntrashedDoneTasks.filter((task) => !taskIds.has(task.id));
	datedUntrashedDoneTasks = [...deduplicatedTasks, ...tasks].sort(compareTasksByDateAndRank);
}

function addOrUpdateTask(task: DatedUntrashedTaskResponse): void {
	addOrUpdateTasks([task]);
}

function addPage(tasks: DatedUntrashedTaskResponse[], pageHasMore: boolean): void {
	addOrUpdateTasks(tasks);
	hasMore = pageHasMore;
}

function reset(): void {
	datedUntrashedDoneTasks = [];
	hasMore = false;
}

function compareTasksByDateAndRank(first: DatedUntrashedTaskResponse, second: DatedUntrashedTaskResponse): number {
	const dateComparison = first.date.compare(second.date);
	if (dateComparison !== 0) return dateComparison;
	return first.rank < second.rank ? -1 : first.rank > second.rank ? 1 : 0;
}

export const datedUntrashedDoneTaskStore = {
	get datedUntrashedDoneTasks(): DatedUntrashedTaskResponse[] {
		return datedUntrashedDoneTasks;
	},
	get hasMore(): boolean {
		return hasMore;
	},
	addOrUpdateTasks,
	addOrUpdateTask,
	addPage,
	reset
};
