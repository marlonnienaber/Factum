import type { TrashedTaskResponse } from '$lib/types/responses/trashed-task-response';

let trashedTasks: TrashedTaskResponse[] = $state<TrashedTaskResponse[]>([]);
let hasMore = $state(false);

function addOrUpdateTasks(tasks: TrashedTaskResponse[]) {
	const taskIds = new Set(tasks.map((task) => task.id));
	const deduplicatedTasks = trashedTasks.filter((task) => !taskIds.has(task.id));
	trashedTasks = [...deduplicatedTasks, ...tasks].sort(compareTasksByRank);
}

function addOrUpdateTask(task: TrashedTaskResponse): void {
	addOrUpdateTasks([task]);
}

function addPage(tasks: TrashedTaskResponse[], pageHasMore: boolean): void {
	addOrUpdateTasks(tasks);
	hasMore = pageHasMore;
}

function removeTask(id: number): void {
	trashedTasks = trashedTasks.filter((task) => task.id !== id);
}

function compareTasksByRank(first: TrashedTaskResponse, second: TrashedTaskResponse): number {
	return first.rank < second.rank ? -1 : first.rank > second.rank ? 1 : 0;
}

function reset(): void {
	trashedTasks = [];
	hasMore = false;
}

export const trashedTaskStore = {
	get trashedTasks(): TrashedTaskResponse[] {
		return trashedTasks;
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
