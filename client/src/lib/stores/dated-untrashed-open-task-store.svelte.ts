import type { DatedUntrashedOpenTaskResponse } from '$lib/types/responses/dated-untrashed-open-task-response';

let datedUntrashedOpenTasks: DatedUntrashedOpenTaskResponse[] = $state<DatedUntrashedOpenTaskResponse[]>([]);

function addTasks(tasks: DatedUntrashedOpenTaskResponse[]): void {
	const taskIds = new Set(tasks.map((task) => task.id));
	const deduplicatedTasks = datedUntrashedOpenTasks.filter((task) => !taskIds.has(task.id));
	datedUntrashedOpenTasks = [...deduplicatedTasks, ...tasks].sort(compareTasksByDateAndRank);
}

function addTask(task: DatedUntrashedOpenTaskResponse): void {
	addTasks([task]);
}

function removeTask(id: number): void {
	datedUntrashedOpenTasks = datedUntrashedOpenTasks.filter((task) => task.id !== id);
}

function compareTasksByDateAndRank(first: DatedUntrashedOpenTaskResponse, second: DatedUntrashedOpenTaskResponse): number {
	const dateComparison = first.date.compare(second.date);
	if (dateComparison !== 0) return dateComparison;
	return first.rank < second.rank ? -1 : first.rank > second.rank ? 1 : 0;
}

function reset(): void {
	datedUntrashedOpenTasks = [];
}

export const datedUntrashedOpenTaskStore = {
	get datedUntrashedOpenTasks(): DatedUntrashedOpenTaskResponse[] {
		return datedUntrashedOpenTasks;
	},
	addTasks,
	addTask,
	removeTask,
	reset
};
