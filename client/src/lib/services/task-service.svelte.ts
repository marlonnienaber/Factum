import {
	type NetworkCall,
	jsonGetWithoutBodyWithResponse,
	jsonPostWithBodyWithResponse,
	jsonPutWithBodyWithResponse
} from '$lib/api/api';
import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
import type { CreateTaskRequest } from '$lib/types/requests/create-task-request';
import type { DatedUntrashedOpenTasksCursor } from '$lib/types/requests/dated-untrashed-open-tasks-cursor';
import type { UpdateTaskRequest } from '$lib/types/requests/update-task-request';
import {
	DatedUntrashedOpenTaskResponseSchema,
	type DatedUntrashedOpenTaskResponse
} from '$lib/types/responses/dated-untrashed-open-task-response';
import { UndatedTaskResponseSchema, type UndatedTaskResponse } from '$lib/types/responses/undated-task-response';
import { z } from 'zod';

let undatedTasksLoading = $state(false);
let datedUntrashedOpenTasksLoading = $state(false);
let taskMutationLoading = $state(false);
const taskMutationBlocked = $derived(taskMutationLoading || undatedTasksLoading || datedUntrashedOpenTasksLoading);

async function fetchUndatedTasks(rank?: string): Promise<void> {
	if (undatedTasksLoading || taskMutationLoading) return;
	undatedTasksLoading = true;
	try {
		const networkCall: NetworkCall<UndatedTaskResponse[]> = await jsonGetWithoutBodyWithResponse(
			`/api/tasks/undated${rank ? `?rank=${encodeURIComponent(rank)}` : ''}`,
			z.array(UndatedTaskResponseSchema)
		);
		if (networkCall.status === 'success') undatedTaskStore.addUndatedTasks(networkCall.result);
	} finally {
		undatedTasksLoading = false;
	}
}

async function fetchDatedUntrashedOpenTasks(cursor?: DatedUntrashedOpenTasksCursor): Promise<void> {
	if (datedUntrashedOpenTasksLoading || taskMutationLoading) return;
	datedUntrashedOpenTasksLoading = true;
	try {
		const queryParameters = cursor
			? `&date=${encodeURIComponent(cursor.date.toString())}&rank=${encodeURIComponent(cursor.rank)}`
			: '';
		const networkCall: NetworkCall<DatedUntrashedOpenTaskResponse[]> = await jsonGetWithoutBodyWithResponse(
			`/api/tasks/dated-untrashed?done=false${queryParameters}`,
			z.array(DatedUntrashedOpenTaskResponseSchema)
		);
		if (networkCall.status === 'success') datedUntrashedOpenTaskStore.addDatedUntrashedOpenTasks(networkCall.result);
	} finally {
		datedUntrashedOpenTasksLoading = false;
	}
}

async function createTask(request: CreateTaskRequest): Promise<void> {
	if (taskMutationBlocked) return;
	taskMutationLoading = true;
	try {
		if (request.date !== undefined) {
			const networkCall: NetworkCall<DatedUntrashedOpenTaskResponse> = await jsonPostWithBodyWithResponse(
				'/api/tasks',
				request,
				DatedUntrashedOpenTaskResponseSchema
			);
			if (networkCall.status === 'success') datedUntrashedOpenTaskStore.addDatedUntrashedOpenTask(networkCall.result);
		} else {
			const networkCall: NetworkCall<UndatedTaskResponse> = await jsonPostWithBodyWithResponse(
				'/api/tasks',
				request,
				UndatedTaskResponseSchema
			);
			if (networkCall.status === 'success') undatedTaskStore.addUndatedTask(networkCall.result);
		}
	} finally {
		taskMutationLoading = false;
	}
}

async function updateTask(request: UpdateTaskRequest): Promise<void> {
	if (taskMutationBlocked) return;
	taskMutationLoading = true;
	try {
		const networkCall: NetworkCall<DatedUntrashedOpenTaskResponse | UndatedTaskResponse> = await jsonPutWithBodyWithResponse(
			'/api/tasks',
			request,
			z.union([DatedUntrashedOpenTaskResponseSchema, UndatedTaskResponseSchema])
		);
		if (networkCall.status !== 'success') return;

		undatedTaskStore.removeTask(request.id);
		datedUntrashedOpenTaskStore.removeTask(request.id);
		if ('date' in networkCall.result) {
			datedUntrashedOpenTaskStore.addDatedUntrashedOpenTask(networkCall.result);
		} else {
			undatedTaskStore.addUndatedTask(networkCall.result);
		}
	} finally {
		taskMutationLoading = false;
	}
}

export const taskService = {
	get taskMutationLoading(): boolean {
		return taskMutationLoading;
	},
	get taskMutationBlocked(): boolean {
		return taskMutationBlocked;
	},
	fetchUndatedTasks,
	fetchDatedUntrashedOpenTasks,
	createTask,
	updateTask
};
