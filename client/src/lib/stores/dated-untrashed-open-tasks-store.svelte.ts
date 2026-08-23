import { type NetworkCall, jsonGetWithoutBodyWithResponse, jsonPostWithBodyWithResponse } from '$lib/api/api';
import type { CreateTaskRequest } from '$lib/types/requests/create-task-request';
import type { DatedUntrashedOpenTasksCursor } from '$lib/types/requests/dated-untrashed-open-tasks-cursor';
import {
	DatedUntrashedOpenTaskResponseSchema,
	type DatedUntrashedOpenTaskResponse
} from '$lib/types/responses/dated-untrashed-open-task-response';
import { z } from 'zod';

let datedUntrashedOpenTasks: DatedUntrashedOpenTaskResponse[] = $state<DatedUntrashedOpenTaskResponse[]>([]);
let loading = $state(false);

async function fetchDatedUntrashedOpenTasks(cursor?: DatedUntrashedOpenTasksCursor): Promise<void> {
	if (loading) return;
	loading = true;
	const queryParameters = cursor
		? `&date=${encodeURIComponent(cursor.date.toString())}&rank=${encodeURIComponent(cursor.rank)}`
		: '';
	const networkCall: NetworkCall<DatedUntrashedOpenTaskResponse[]> = await jsonGetWithoutBodyWithResponse(
		`/api/tasks/dated-untrashed?done=false${queryParameters}`,
		z.array(DatedUntrashedOpenTaskResponseSchema)
	);
	if (networkCall.status === 'success') {
		const fetchedTasks = networkCall.result;
		const fetchedIds = fetchedTasks.map((task) => task.id);
		const deduplicatedTasks = datedUntrashedOpenTasks.filter((task) => !fetchedIds.includes(task.id));
		datedUntrashedOpenTasks = [...deduplicatedTasks, ...fetchedTasks].sort(compareTasksByDateAndRank);
	}
	loading = false;
}

async function createTask(request: CreateTaskRequest): Promise<void> {
	if (loading) return;
	loading = true;
	const networkCall: NetworkCall<DatedUntrashedOpenTaskResponse> = await jsonPostWithBodyWithResponse(
		'/api/tasks',
		request,
		DatedUntrashedOpenTaskResponseSchema
	);
	if (networkCall.status === 'success') {
		const createdTask = networkCall.result;
		datedUntrashedOpenTasks = [...datedUntrashedOpenTasks.filter((task) => task.id !== createdTask.id), createdTask].sort(
			compareTasksByDateAndRank
		);
	}
	loading = false;
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
	get loading(): boolean {
		return loading;
	},
	fetchDatedUntrashedOpenTasks,
	createTask,
	reset
};
