import type { UndatedTaskResponse } from '$lib/types/responses/undated-task-response';

let undatedTasks: UndatedTaskResponse[] = $state<UndatedTaskResponse[]>([]);
let hasMore = $state(false);

function addOrUpdateTasks(tasks: UndatedTaskResponse[]): void {
	const taskIds = new Set(tasks.map((task) => task.id));
	const deduplicatedTasks = undatedTasks.filter((task) => !taskIds.has(task.id));
	undatedTasks = [...deduplicatedTasks, ...tasks].sort(compareTasksByRank);
}

function addOrUpdateTask(task: UndatedTaskResponse): void {
	addOrUpdateTasks([task]);
}

function addPage(tasks: UndatedTaskResponse[], pageHasMore: boolean): void {
	addOrUpdateTasks(tasks);
	hasMore = pageHasMore;
}

function removeTask(id: number): void {
	removeTasks([id]);
}

function removeTasks(ids: number[]): void {
	const taskIds = new Set(ids);
	undatedTasks = undatedTasks.filter((task) => !taskIds.has(task.id));
}

function compareTasksByRank(first: UndatedTaskResponse, second: UndatedTaskResponse): number {
	return first.rank < second.rank ? -1 : first.rank > second.rank ? 1 : 0;
}

function reset(): void {
	undatedTasks = [];
	hasMore = false;
}

export const undatedTaskStore = {
	get undatedTasks(): UndatedTaskResponse[] {
		return undatedTasks;
	},
	get hasMore(): boolean {
		return hasMore;
	},
	addOrUpdateTasks,
	addOrUpdateTask,
	addPage,
	removeTask,
	removeTasks,
	reset
};
