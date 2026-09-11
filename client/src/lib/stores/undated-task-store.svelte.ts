import type { UndatedTaskResponse } from '$lib/types/responses/undated-task-response';

let undatedTasks: UndatedTaskResponse[] = $state<UndatedTaskResponse[]>([]);

function addOrUpdateTasks(tasks: UndatedTaskResponse[]): void {
	const taskIds = new Set(tasks.map((task) => task.id));
	const deduplicatedTasks = undatedTasks.filter((task) => !taskIds.has(task.id));
	undatedTasks = [...deduplicatedTasks, ...tasks].sort(compareTasksByRank);
}

function addOrUpdateTask(task: UndatedTaskResponse): void {
	addOrUpdateTasks([task]);
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
}

export const undatedTaskStore = {
	get undatedTasks(): UndatedTaskResponse[] {
		return undatedTasks;
	},
	addOrUpdateTasks,
	addOrUpdateTask,
	removeTask,
	removeTasks,
	reset
};
