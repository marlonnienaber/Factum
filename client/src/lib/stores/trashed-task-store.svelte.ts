import type { TrashedTaskResponse } from '$lib/types/responses/trashed-task-response';

let trashedTasks: TrashedTaskResponse[] = $state<TrashedTaskResponse[]>([]);

function addOrUpdateTasks(tasks: TrashedTaskResponse[]) {
	const taskIds = new Set(tasks.map((task) => task.id));
	const deduplicatedTasks = trashedTasks.filter((task) => !taskIds.has(task.id));
	trashedTasks = [...deduplicatedTasks, ...tasks].sort(compareTasksByTrashedDateAndId);
}

function addOrUpdateTask(task: TrashedTaskResponse): void {
	addOrUpdateTasks([task]);
}

function compareTasksByTrashedDateAndId(first: TrashedTaskResponse, second: TrashedTaskResponse): number {
	const trashedDateComparison = second.trashedDate.compare(first.trashedDate);
	if (trashedDateComparison !== 0) return trashedDateComparison;
	return second.id - first.id;
}

function reset(): void {
	trashedTasks = [];
}

export const trashedTaskStore = {
	get trashedTasks(): TrashedTaskResponse[] {
		return trashedTasks;
	},
	addOrUpdateTasks,
	addOrUpdateTask,
	reset
};
