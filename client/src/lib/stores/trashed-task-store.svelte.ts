import type { TrashedTaskResponse } from '$lib/types/responses/trashed-task-response';

let trashedTasks: TrashedTaskResponse[] = $state<TrashedTaskResponse[]>([]);
let hasMore = $state(false);
let trashedTasksExist = $state(false);

function setTrashedTasksExist(value: boolean): void {
	trashedTasksExist = value;
}

function addOrUpdateTasks(tasks: TrashedTaskResponse[]) {
	if (tasks.length > 0) trashedTasksExist = true;
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

function clearPages(): void {
	trashedTasks = [];
	hasMore = false;
}

function reset(): void {
	clearPages();
	trashedTasksExist = false;
}

function compareTasksByRank(first: TrashedTaskResponse, second: TrashedTaskResponse): number {
	return first.rank < second.rank ? -1 : first.rank > second.rank ? 1 : 0;
}

export const trashedTaskStore = {
	get trashedTasks(): TrashedTaskResponse[] {
		return trashedTasks;
	},
	get hasMore(): boolean {
		return hasMore;
	},
	get trashedTasksExist(): boolean {
		return trashedTasksExist;
	},
	addOrUpdateTasks,
	addOrUpdateTask,
	addPage,
	removeTask,
	setTrashedTasksExist,
	clearPages,
	reset
};
