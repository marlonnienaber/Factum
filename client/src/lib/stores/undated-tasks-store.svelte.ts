import { type NetworkCall, jsonGetWithoutBodyWithResponse, jsonPostWithBodyWithResponse } from '$lib/api/api';
import type { CreateTaskRequest } from '$lib/types/requests/create-task-request';
import { UndatedTaskResponseSchema, type UndatedTaskResponse } from '$lib/types/responses/undated-task-response';
import { z } from 'zod';

let undatedTasks: UndatedTaskResponse[] = $state<UndatedTaskResponse[]>([]);
let loading = $state(false);

async function fetchUndatedTasks(rank?: string): Promise<void> {
	if (loading) return;
	loading = true;
	const networkCall: NetworkCall<UndatedTaskResponse[]> = await jsonGetWithoutBodyWithResponse(
		`/api/tasks/undated${rank ? `?rank=${encodeURIComponent(rank)}` : ''}`,
		z.array(UndatedTaskResponseSchema)
	);
	if (networkCall.status === 'success') {
		const fetchedTasks = networkCall.result;
		const fetchedIds = fetchedTasks.map((task) => task.id);
		const deduplicatedTasks = undatedTasks.filter((task) => !fetchedIds.includes(task.id));
		undatedTasks = [...deduplicatedTasks, ...fetchedTasks].sort(compareTasksByRank);
	}
	loading = false;
}
async function createTask(request: CreateTaskRequest): Promise<void> {
	if (loading) return;
	loading = true;
	const networkCall: NetworkCall<UndatedTaskResponse> = await jsonPostWithBodyWithResponse(
		'/api/tasks',
		request,
		UndatedTaskResponseSchema
	);
	if (networkCall.status === 'success') {
		const createdTask = networkCall.result;
		undatedTasks = [...undatedTasks.filter((task) => task.id !== createdTask.id), createdTask].sort(compareTasksByRank);
	}
	loading = false;
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
	get loading(): boolean {
		return loading;
	},
	fetchUndatedTasks,
	createTask,
	reset
};
