import { type NetworkCall, jsonGetWithoutBodyWithResponse, jsonPostWithBodyWithResponse } from '$lib/api/api';
import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
import type { CreateTaskRequest } from '$lib/types/requests/create-task-request';
import type { DatedUntrashedOpenTasksCursor } from '$lib/types/requests/dated-untrashed-open-tasks-cursor';
import {
	DatedUntrashedOpenTaskResponseSchema,
	type DatedUntrashedOpenTaskResponse
} from '$lib/types/responses/dated-untrashed-open-task-response';
import { UndatedTaskResponseSchema, type UndatedTaskResponse } from '$lib/types/responses/undated-task-response';
import { z } from 'zod';

let undatedTasksLoading = $state(false);
let datedUntrashedOpenTasksLoading = $state(false);
let taskCreationLoading = $state(false);

async function fetchUndatedTasks(rank?: string): Promise<void> {
	if (undatedTasksLoading) return;
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
	if (datedUntrashedOpenTasksLoading) return;
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
	if (taskCreationLoading) return;
	taskCreationLoading = true;
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
		taskCreationLoading = false;
	}
}

export const taskService = {
	get taskCreationLoading(): boolean {
		return taskCreationLoading;
	},
	fetchUndatedTasks,
	fetchDatedUntrashedOpenTasks,
	createTask
};
