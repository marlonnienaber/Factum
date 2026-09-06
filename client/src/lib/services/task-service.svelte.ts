import {
	type NetworkCall,
	jsonGetWithoutBodyWithResponse,
	jsonPatchWithBodyWithResponse,
	jsonPatchWithBodyWithoutResponse,
	jsonPatchWithoutBodyWithoutResponse,
	jsonPostWithBodyWithResponse,
	jsonPutWithBodyWithResponse
} from '$lib/api/api';
import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
import { trashedTaskStore } from '$lib/stores/trashed-task-store.svelte';
import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
import type { BatchToggleTaskDoneRequest } from '$lib/types/requests/batch-toggle-task-done-request';
import type { ChangeTaskDateAndOrRankRequest } from '$lib/types/requests/change-task-date-and-or-rank-request';
import type { CreateTaskRequest } from '$lib/types/requests/create-task-request';
import type { DatedUntrashedOpenTasksCursor } from '$lib/types/requests/dated-untrashed-open-tasks-cursor';
import type { UpdateTaskRequest } from '$lib/types/requests/update-task-request';
import {
	BatchToggleTaskDoneResponseSchema,
	type BatchToggleTaskDoneResponse
} from '$lib/types/responses/batch-toggle-task-done-response';
import {
	DatedUntrashedOpenTaskResponseSchema,
	type DatedUntrashedOpenTaskResponse
} from '$lib/types/responses/dated-untrashed-open-task-response';
import { UndatedTaskResponseSchema, type UndatedTaskResponse } from '$lib/types/responses/undated-task-response';
import { getLocalTimeZone, now, today, type CalendarDate } from '@internationalized/date';
import { generateKeyBetween } from 'fractional-indexing';
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
		if (networkCall.status === 'success') undatedTaskStore.addTasks(networkCall.result);
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
		if (networkCall.status === 'success') datedUntrashedOpenTaskStore.addTasks(networkCall.result);
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
			if (networkCall.status === 'success') datedUntrashedOpenTaskStore.addTask(networkCall.result);
		} else {
			const networkCall: NetworkCall<UndatedTaskResponse> = await jsonPostWithBodyWithResponse(
				'/api/tasks',
				request,
				UndatedTaskResponseSchema
			);
			if (networkCall.status === 'success') undatedTaskStore.addTask(networkCall.result);
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
			datedUntrashedOpenTaskStore.addTask(networkCall.result);
		} else {
			undatedTaskStore.addTask(networkCall.result);
		}
	} finally {
		taskMutationLoading = false;
	}
}

async function completeTask(id: number): Promise<void> {
	if (taskMutationBlocked) return;

	const datedTask = datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.find((task) => task.id === id);
	const undatedTask = undatedTaskStore.undatedTasks.find((task) => task.id === id);
	const task = datedTask ?? undatedTask;
	if (!task) return;

	const date = datedTask?.date ?? today(getLocalTimeZone());
	const lastTaskForDate = datedUntrashedDoneTaskStore.datedUntrashedDoneTasks
		.filter((doneTask) => doneTask.date.compare(date) === 0)
		.at(-1);
	const rank = generateKeyBetween(lastTaskForDate?.rank, undefined);

	taskMutationLoading = true;
	try {
		const networkCall = await jsonPatchWithoutBodyWithoutResponse(
			`/api/tasks/${id}/toggle-done?rank=${encodeURIComponent(rank)}`
		);
		if (networkCall.status !== 'success') return;

		undatedTaskStore.removeTask(id);
		datedUntrashedOpenTaskStore.removeTask(id);
		datedUntrashedDoneTaskStore.addTask({ ...task, date, rank });
	} finally {
		taskMutationLoading = false;
	}
}

async function completeTasks(taskIds: number[]): Promise<boolean> {
	if (taskMutationBlocked) return false;

	const request: BatchToggleTaskDoneRequest = { taskIds };
	taskMutationLoading = true;
	try {
		const networkCall: NetworkCall<BatchToggleTaskDoneResponse> = await jsonPatchWithBodyWithResponse(
			'/api/tasks/batch-toggle-done',
			request,
			BatchToggleTaskDoneResponseSchema
		);
		if (networkCall.status !== 'success') return false;

		for (const taskId of taskIds) {
			undatedTaskStore.removeTask(taskId);
			datedUntrashedOpenTaskStore.removeTask(taskId);
		}
		datedUntrashedDoneTaskStore.addTasks(networkCall.result.datedUntrashedDoneTasks);
		return true;
	} finally {
		taskMutationLoading = false;
	}
}

async function trashTask(id: number): Promise<void> {
	if (taskMutationBlocked) return;
	const datedTask = datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.find((task) => task.id === id);
	const undatedTask = undatedTaskStore.undatedTasks.find((task) => task.id === id);
	const task = datedTask ?? undatedTask;
	if (!task) return;

	taskMutationLoading = true;
	try {
		const trashedDate = now(getLocalTimeZone());
		const encodedTrashedDate = encodeURIComponent(trashedDate.toAbsoluteString());
		const networkCall = await jsonPatchWithoutBodyWithoutResponse(`/api/tasks/${id}/trash?trashedDate=${encodedTrashedDate}`);
		if (networkCall.status !== 'success') return;

		undatedTaskStore.removeTask(id);
		datedUntrashedOpenTaskStore.removeTask(id);
		trashedTaskStore.addTask({
			id: task.id,
			categoryId: task.categoryId,
			title: task.title,
			description: task.description,
			date: datedTask?.date,
			done: false,
			trashedDate
		});
	} finally {
		taskMutationLoading = false;
	}
}

async function changeTaskDateAndOrRank(
	id: number,
	newDate: CalendarDate | undefined,
	previousRank: string | undefined,
	nextRank: string | undefined
): Promise<void> {
	if (taskMutationBlocked) return;

	const undatedTask = undatedTaskStore.undatedTasks.find((task) => task.id === id);
	const datedTask = datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.find((task) => task.id === id);
	const task = undatedTask ?? datedTask;
	if (!task) return;

	const rank = generateKeyBetween(previousRank, nextRank);
	const request: ChangeTaskDateAndOrRankRequest = { id, date: newDate?.toString(), rank };

	taskMutationLoading = true;

	undatedTaskStore.removeTask(id);
	datedUntrashedOpenTaskStore.removeTask(id);
	if (newDate) {
		datedUntrashedOpenTaskStore.addTask({ ...task, date: newDate, rank });
	} else {
		undatedTaskStore.addTask({
			id: task.id,
			categoryId: task.categoryId,
			title: task.title,
			description: task.description,
			rank
		});
	}

	try {
		const networkCall = await jsonPatchWithBodyWithoutResponse('/api/tasks/change-date-and-or-rank', request);

		if (networkCall.status === 'failed') {
			datedUntrashedOpenTaskStore.removeTask(id);
			undatedTaskStore.removeTask(id);
			if (datedTask) datedUntrashedOpenTaskStore.addTask(datedTask);
			if (undatedTask) undatedTaskStore.addTask(undatedTask);
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
	updateTask,
	completeTask,
	completeTasks,
	trashTask,
	changeTaskDateAndOrRank
};
