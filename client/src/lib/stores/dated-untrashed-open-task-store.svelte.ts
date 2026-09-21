import type { DatedUntrashedOpenTaskResponse } from '$lib/types/responses/dated-untrashed-open-task-response';
import { getLocalTimeZone, today } from '@internationalized/date';

let datedUntrashedOpenTodayTasks: DatedUntrashedOpenTaskResponse[] = $state<DatedUntrashedOpenTaskResponse[]>([]);
let datedUntrashedOpenNonTodayTasks: DatedUntrashedOpenTaskResponse[] = $state<DatedUntrashedOpenTaskResponse[]>([]);
let todayTasksHasMore = $state(false);
let nonTodayTasksHasMore = $state(false);

function addOrUpdateTasks(tasks: DatedUntrashedOpenTaskResponse[]): void {
	const taskIds = new Set(tasks.map((task) => task.id));
	const currentDate = today(getLocalTimeZone());
	const todayTasks = tasks.filter((task) => task.date.compare(currentDate) === 0);
	const nonTodayTasks = tasks.filter((task) => task.date.compare(currentDate) !== 0);

	datedUntrashedOpenTodayTasks = [
		...datedUntrashedOpenTodayTasks.filter((task) => !taskIds.has(task.id)),
		...todayTasks
	].sort(compareTasksByDateAndRank);
	datedUntrashedOpenNonTodayTasks = [
		...datedUntrashedOpenNonTodayTasks.filter((task) => !taskIds.has(task.id)),
		...nonTodayTasks
	].sort(compareTasksByDateAndRank);
}

function addOrUpdateTask(task: DatedUntrashedOpenTaskResponse): void {
	addOrUpdateTasks([task]);
}

function addTodayPage(tasks: DatedUntrashedOpenTaskResponse[], pageHasMore: boolean): void {
	const taskIds = tasks.map((task) => task.id);
	removeTasks(taskIds);
	datedUntrashedOpenTodayTasks = [...datedUntrashedOpenTodayTasks, ...tasks].sort(compareTasksByDateAndRank);
	todayTasksHasMore = pageHasMore;
}

function addNonTodayPage(tasks: DatedUntrashedOpenTaskResponse[], pageHasMore: boolean): void {
	const taskIds = tasks.map((task) => task.id);
	removeTasks(taskIds);
	datedUntrashedOpenNonTodayTasks = [...datedUntrashedOpenNonTodayTasks, ...tasks].sort(compareTasksByDateAndRank);
	nonTodayTasksHasMore = pageHasMore;
}

function removeTask(id: number): void {
	removeTasks([id]);
}

function removeTasks(ids: number[]): void {
	const taskIds = new Set(ids);
	datedUntrashedOpenTodayTasks = datedUntrashedOpenTodayTasks.filter((task) => !taskIds.has(task.id));
	datedUntrashedOpenNonTodayTasks = datedUntrashedOpenNonTodayTasks.filter((task) => !taskIds.has(task.id));
}

function compareTasksByDateAndRank(first: DatedUntrashedOpenTaskResponse, second: DatedUntrashedOpenTaskResponse): number {
	const dateComparison = first.date.compare(second.date);
	if (dateComparison !== 0) return dateComparison;
	return first.rank < second.rank ? -1 : first.rank > second.rank ? 1 : 0;
}

function reset(): void {
	datedUntrashedOpenTodayTasks = [];
	datedUntrashedOpenNonTodayTasks = [];
	todayTasksHasMore = false;
	nonTodayTasksHasMore = false;
}

export const datedUntrashedOpenTaskStore = {
	get datedUntrashedOpenTasks(): DatedUntrashedOpenTaskResponse[] {
		return [...datedUntrashedOpenTodayTasks, ...datedUntrashedOpenNonTodayTasks].sort(compareTasksByDateAndRank);
	},
	get datedUntrashedOpenTodayTasks(): DatedUntrashedOpenTaskResponse[] {
		return datedUntrashedOpenTodayTasks;
	},
	get datedUntrashedOpenNonTodayTasks(): DatedUntrashedOpenTaskResponse[] {
		return datedUntrashedOpenNonTodayTasks;
	},
	get todayTasksHasMore(): boolean {
		return todayTasksHasMore;
	},
	get nonTodayTasksHasMore(): boolean {
		return nonTodayTasksHasMore;
	},
	addOrUpdateTasks,
	addOrUpdateTask,
	addTodayPage,
	addNonTodayPage,
	removeTask,
	removeTasks,
	reset
};
