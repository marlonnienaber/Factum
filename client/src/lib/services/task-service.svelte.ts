import {
	type NetworkCall,
	jsonGetWithoutBodyWithResponse,
	jsonPatchWithBodyWithResponse,
	jsonPatchWithBodyWithoutResponse,
	jsonPatchWithoutBodyWithResponse,
	jsonPatchWithoutBodyWithoutResponse,
	jsonPostWithBodyWithResponse,
	jsonPutWithBodyWithResponse
} from '$lib/utils/networking';
import { datedUntrashedDoneTaskStore } from '$lib/stores/dated-untrashed-done-task-store.svelte';
import { datedUntrashedOpenTaskStore } from '$lib/stores/dated-untrashed-open-task-store.svelte';
import { trashedTaskStore } from '$lib/stores/trashed-task-store.svelte';
import { undatedTaskStore } from '$lib/stores/undated-task-store.svelte';
import type { BatchTrashTasksRequest } from '$lib/types/requests/batch-trash-tasks-request';
import type { BatchUpdateDoneRequest } from '$lib/types/requests/batch-update-done-request';
import type { CreateTaskRequest } from '$lib/types/requests/create-task-request';
import type { DatedUntrashedOpenTasksCursor } from '$lib/types/requests/dated-untrashed-open-tasks-cursor';
import type { TaskDragAndDropRequest } from '$lib/types/requests/task-drag-and-drop-request';
import type { UpdateTaskRequest } from '$lib/types/requests/update-task-request';
import { BatchCompleteResponseSchema, type BatchCompleteResponse } from '$lib/types/responses/batch-complete-response';
import { CreateTaskResponseSchema, type CreateTaskResponse } from '$lib/types/responses/create-task-response';
import {
	DatedUntrashedOpenTaskResponseSchema,
	type DatedUntrashedOpenTaskResponse
} from '$lib/types/responses/dated-untrashed-open-task-response';
import type { TrashedTaskResponse } from '$lib/types/responses/trashed-task-response';
import { TaskDragAndDropResponseSchema, type TaskDragAndDropResponse } from '$lib/types/responses/task-drag-and-drop-response';
import { ToggleTaskDoneResponseSchema, type ToggleTaskDoneResponse } from '$lib/types/responses/toggle-task-done-response';
import { UndatedTaskResponseSchema, type UndatedTaskResponse } from '$lib/types/responses/undated-task-response';
import { getLocalTimeZone, now, type ZonedDateTime } from '@internationalized/date';
import { toast } from 'svelte-sonner';
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
		if (networkCall.status === 'success') undatedTaskStore.addOrUpdateTasks(networkCall.result);
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
		if (networkCall.status === 'success') datedUntrashedOpenTaskStore.addOrUpdateTasks(networkCall.result);
	} finally {
		datedUntrashedOpenTasksLoading = false;
	}
}

async function createTask(request: CreateTaskRequest): Promise<void> {
	if (taskMutationBlocked) return;
	taskMutationLoading = true;
	try {
		const networkCall: NetworkCall<CreateTaskResponse> = await jsonPostWithBodyWithResponse(
			'/api/tasks',
			request,
			CreateTaskResponseSchema
		);
		if (networkCall.status !== 'success') return;

		if (request.date !== undefined) {
			datedUntrashedOpenTaskStore.addOrUpdateTask({
				...request,
				...networkCall.result,
				date: request.date
			});
		} else {
			undatedTaskStore.addOrUpdateTask({ ...request, ...networkCall.result });
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
			datedUntrashedOpenTaskStore.addOrUpdateTask(networkCall.result);
		} else {
			undatedTaskStore.addOrUpdateTask(networkCall.result);
		}
	} finally {
		taskMutationLoading = false;
	}
}

async function completeTask(id: number): Promise<void> {
	if (taskMutationBlocked) return;

	const datedTask = datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.find((task) => task.id === id);
	const undatedTask = undatedTaskStore.undatedTasks.find((task) => task.id === id);
	const trashedTask = trashedTaskStore.trashedTasks.find((task) => task.id === id);
	const task = datedTask ?? undatedTask ?? trashedTask;
	if (!task) return;

	taskMutationLoading = true;
	try {
		const networkCall: NetworkCall<ToggleTaskDoneResponse> = await jsonPatchWithoutBodyWithResponse(
			`/api/tasks/${id}/toggle-done`,
			ToggleTaskDoneResponseSchema
		);
		if (networkCall.status !== 'success') return;
		if (trashedTask !== undefined) {
			trashedTaskStore.addOrUpdateTask({
				...trashedTask,
				date: networkCall.result.date,
				done: true
			});
			return;
		}
		if (networkCall.result.rank === undefined) {
			toast.error('An unexpected error occurred. Your data might be inconsistent. Please reload the page to resolve the issue.');
			return;
		}

		undatedTaskStore.removeTask(id);
		datedUntrashedOpenTaskStore.removeTask(id);
		datedUntrashedDoneTaskStore.addOrUpdateTask({
			...task,
			date: networkCall.result.date,
			rank: networkCall.result.rank
		});
	} finally {
		taskMutationLoading = false;
	}
}

async function completeTasks(taskIds: number[]): Promise<boolean> {
	if (taskMutationBlocked) return false;

	const request: BatchUpdateDoneRequest = { taskIds };
	taskMutationLoading = true;
	try {
		const networkCall: NetworkCall<BatchCompleteResponse> = await jsonPatchWithBodyWithResponse(
			'/api/tasks/batch-complete',
			request,
			BatchCompleteResponseSchema
		);
		if (networkCall.status !== 'success') return false;

		undatedTaskStore.removeTasks(taskIds);
		datedUntrashedOpenTaskStore.removeTasks(taskIds);
		datedUntrashedDoneTaskStore.addOrUpdateTasks(networkCall.result.datedUntrashedDoneTasks);
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
		trashedTaskStore.addOrUpdateTask({
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

async function trashTasks(taskIds: number[]): Promise<boolean> {
	if (taskMutationBlocked) return false;

	const tasks = taskIds
		.map((taskId) => {
			const datedTask = datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.find((task) => task.id === taskId);
			const undatedTask = undatedTaskStore.undatedTasks.find((task) => task.id === taskId);
			return datedTask ?? undatedTask;
		})
		.filter((task) => task !== undefined);
	if (tasks.length !== taskIds.length) return false;

	const trashedDate: ZonedDateTime = now(getLocalTimeZone());
	const encodedTrashedDate: string = encodeURIComponent(trashedDate.toAbsoluteString());
	const request: BatchTrashTasksRequest = { taskIds };
	taskMutationLoading = true;
	try {
		const networkCall: NetworkCall<void> = await jsonPatchWithBodyWithoutResponse(
			`/api/tasks/batch-trash?trashedDate=${encodedTrashedDate}`,
			request
		);
		if (networkCall.status !== 'success') return false;

		const trashedTasks: TrashedTaskResponse[] = tasks.map((task) => {
			const datedTask = datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.find((datedTask) => datedTask.id === task.id);
			return {
				id: task.id,
				categoryId: task.categoryId,
				title: task.title,
				description: task.description,
				date: datedTask?.date,
				done: false,
				trashedDate
			};
		});
		undatedTaskStore.removeTasks(taskIds);
		datedUntrashedOpenTaskStore.removeTasks(taskIds);
		trashedTaskStore.addOrUpdateTasks(trashedTasks);
		return true;
	} finally {
		taskMutationLoading = false;
	}
}

async function dragAndDropTask(id: number, request: TaskDragAndDropRequest): Promise<void> {
	if (taskMutationBlocked) return;

	const undatedTask = undatedTaskStore.undatedTasks.find((task) => task.id === id);
	const datedTask = datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.find((task) => task.id === id);
	const task = undatedTask ?? datedTask;
	if (!task) return;

	taskMutationLoading = true;

	try {
		const networkCall: NetworkCall<TaskDragAndDropResponse> = await jsonPatchWithBodyWithResponse(
			`/api/tasks/${id}/drag-and-drop`,
			request,
			TaskDragAndDropResponseSchema
		);
		if (networkCall.status === 'failed') return;

		undatedTaskStore.removeTask(id);
		datedUntrashedOpenTaskStore.removeTask(id);
		if (request.date) {
			datedUntrashedOpenTaskStore.addOrUpdateTask({ ...task, date: request.date, rank: networkCall.result.rank });
		} else {
			undatedTaskStore.addOrUpdateTask({ ...task, rank: networkCall.result.rank });
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
	trashTasks,
	dragAndDropTask
};
