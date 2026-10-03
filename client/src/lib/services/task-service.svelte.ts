import {
	type NetworkCall,
	jsonDeleteWithoutBodyWithoutResponse,
	jsonGetWithoutBodyWithResponse,
	jsonPatchWithBodyWithResponse,
	jsonPatchWithoutBodyWithResponse,
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
import type { DatedUntrashedTasksCursor } from '$lib/types/requests/dated-untrashed-tasks-cursor';
import type { TaskDragAndDropRequest } from '$lib/types/requests/task-drag-and-drop-request';
import type { UpdateTaskRequest } from '$lib/types/requests/update-task-request';
import { BatchCompleteResponseSchema, type BatchCompleteResponse } from '$lib/types/responses/batch-complete-response';
import { BatchOpenResponseSchema, type BatchOpenResponse } from '$lib/types/responses/batch-open-response';
import { BatchTrashResponseSchema, type BatchTrashResponse } from '$lib/types/responses/batch-trash-response';
import { CreateTaskResponseSchema, type CreateTaskResponse } from '$lib/types/responses/create-task-response';
import {
	GetDatedUntrashedDoneTasksResponseSchema,
	type GetDatedUntrashedDoneTasksResponse
} from '$lib/types/responses/get-dated-untrashed-done-tasks-response';
import {
	GetDatedUntrashedOpenTasksResponseSchema,
	type GetDatedUntrashedOpenTasksResponse
} from '$lib/types/responses/get-dated-untrashed-open-tasks-response';
import { GetUndatedTasksResponseSchema, type GetUndatedTasksResponse } from '$lib/types/responses/get-undated-tasks-response';
import { GetTrashedTasksResponseSchema, type GetTrashedTasksResponse } from '$lib/types/responses/get-trashed-tasks-response';
import {
	GetTrashedTasksExistResponseSchema,
	type GetTrashedTasksExistResponse
} from '$lib/types/responses/get-trashed-tasks-exist-response';
import type { TrashedTaskResponse } from '$lib/types/responses/trashed-task-response';
import { TrashTaskResponseSchema, type TrashTaskResponse } from '$lib/types/responses/trash-task-response';
import { RestoreTaskResponseSchema, type RestoreTaskResponse } from '$lib/types/responses/restore-task-response';
import { TaskDragAndDropResponseSchema, type TaskDragAndDropResponse } from '$lib/types/responses/task-drag-and-drop-response';
import { ToggleTaskDoneResponseSchema, type ToggleTaskDoneResponse } from '$lib/types/responses/toggle-task-done-response';
import { UpdateTaskResponseSchema, type UpdateTaskResponse } from '$lib/types/responses/update-task-response';
import type { DateValue } from '@internationalized/date';
import { SvelteURLSearchParams } from 'svelte/reactivity';
import { toast } from 'svelte-sonner';

async function fetchDatedUntrashedDoneTasks(
	cursor?: DatedUntrashedTasksCursor,
	categoryIds: number[] = [],
	startDate?: DateValue,
	endDate?: DateValue
): Promise<void> {
	const queryParameters = new SvelteURLSearchParams();
	for (const categoryId of categoryIds) queryParameters.append('categoryId', String(categoryId));
	if (startDate !== undefined) queryParameters.set('startDate', startDate.toString());
	if (endDate !== undefined) queryParameters.set('endDate', endDate.toString());
	if (cursor) {
		queryParameters.set('cursorDate', cursor.date.toString());
		queryParameters.set('cursorRank', cursor.rank);
	}
	const encodedQueryParameters = queryParameters.toString();
	const url = `/api/tasks/dated-untrashed-done${encodedQueryParameters ? `?${encodedQueryParameters}` : ''}`;
	const networkCall: NetworkCall<GetDatedUntrashedDoneTasksResponse> = await jsonGetWithoutBodyWithResponse(
		url,
		GetDatedUntrashedDoneTasksResponseSchema
	);
	if (networkCall.status === 'success') {
		if (cursor === undefined) {
			datedUntrashedDoneTaskStore.reset();
		}
		datedUntrashedDoneTaskStore.addPage(networkCall.result.tasks, networkCall.result.hasMore);
	}
}

async function fetchUndatedTasks(rank?: string): Promise<void> {
	const networkCall: NetworkCall<GetUndatedTasksResponse> = await jsonGetWithoutBodyWithResponse(
		`/api/tasks/undated${rank ? `?rank=${encodeURIComponent(rank)}` : ''}`,
		GetUndatedTasksResponseSchema
	);
	if (networkCall.status === 'success') {
		if (rank === undefined) undatedTaskStore.reset();
		undatedTaskStore.addPage(networkCall.result.tasks, networkCall.result.hasMore);
	}
}

async function fetchDatedUntrashedOpenTodayTasks(rank?: string, categoryIds: number[] = []): Promise<void> {
	const queryParameters = new SvelteURLSearchParams();
	for (const categoryId of categoryIds) queryParameters.append('categoryId', String(categoryId));
	if (rank !== undefined) queryParameters.set('rank', rank);
	const encodedQueryParameters = queryParameters.toString();
	const networkCall: NetworkCall<GetDatedUntrashedOpenTasksResponse> = await jsonGetWithoutBodyWithResponse(
		`/api/tasks/dated-untrashed-open-today${encodedQueryParameters ? `?${encodedQueryParameters}` : ''}`,
		GetDatedUntrashedOpenTasksResponseSchema
	);
	if (networkCall.status === 'success') {
		if (rank === undefined) datedUntrashedOpenTaskStore.clearTodayPages();
		datedUntrashedOpenTaskStore.addTodayPage(networkCall.result.tasks, networkCall.result.hasMore);
	}
}

async function fetchDatedUntrashedDoneTodayTasks(rank?: string, categoryIds: number[] = []): Promise<void> {
	const queryParameters = new SvelteURLSearchParams();
	for (const categoryId of categoryIds) queryParameters.append('categoryId', String(categoryId));
	if (rank !== undefined) queryParameters.set('rank', rank);
	const encodedQueryParameters = queryParameters.toString();
	const networkCall: NetworkCall<GetDatedUntrashedDoneTasksResponse> = await jsonGetWithoutBodyWithResponse(
		`/api/tasks/dated-untrashed-done-today${encodedQueryParameters ? `?${encodedQueryParameters}` : ''}`,
		GetDatedUntrashedDoneTasksResponseSchema
	);
	if (networkCall.status === 'success') {
		if (rank === undefined) datedUntrashedDoneTaskStore.reset();
		datedUntrashedDoneTaskStore.addPage(networkCall.result.tasks, networkCall.result.hasMore);
	}
}

async function fetchDatedUntrashedOpenNonTodayTasks(cursor?: DatedUntrashedTasksCursor): Promise<void> {
	const queryParameters = cursor
		? `?date=${encodeURIComponent(cursor.date.toString())}&rank=${encodeURIComponent(cursor.rank)}`
		: '';
	const networkCall: NetworkCall<GetDatedUntrashedOpenTasksResponse> = await jsonGetWithoutBodyWithResponse(
		`/api/tasks/dated-untrashed-open-non-today${queryParameters}`,
		GetDatedUntrashedOpenTasksResponseSchema
	);
	if (networkCall.status === 'success') {
		if (cursor === undefined) datedUntrashedOpenTaskStore.clearNonTodayPages();
		datedUntrashedOpenTaskStore.addNonTodayPage(networkCall.result.tasks, networkCall.result.hasMore);
	}
}

async function fetchTrashedTasks(
	rank?: string,
	categoryIds: number[] = [],
	startDate?: DateValue,
	endDate?: DateValue
): Promise<void> {
	const queryParameters = new SvelteURLSearchParams();
	for (const categoryId of categoryIds) queryParameters.append('categoryId', String(categoryId));
	if (startDate !== undefined) queryParameters.set('startDate', startDate.toString());
	if (endDate !== undefined) queryParameters.set('endDate', endDate.toString());
	if (rank !== undefined) queryParameters.set('rank', rank);
	const encodedQueryParameters = queryParameters.toString();
	const networkCall: NetworkCall<GetTrashedTasksResponse> = await jsonGetWithoutBodyWithResponse(
		`/api/tasks/trashed${encodedQueryParameters ? `?${encodedQueryParameters}` : ''}`,
		GetTrashedTasksResponseSchema
	);
	if (networkCall.status === 'success') {
		if (rank === undefined) {
			trashedTaskStore.clearPages();
		}
		trashedTaskStore.addPage(networkCall.result.tasks, networkCall.result.hasMore);
	}
}

async function fetchTrashedTasksExist(): Promise<void> {
	const networkCall: NetworkCall<GetTrashedTasksExistResponse> = await jsonGetWithoutBodyWithResponse(
		'/api/tasks/trashed/exists',
		GetTrashedTasksExistResponseSchema
	);
	if (networkCall.status === 'success') {
		trashedTaskStore.setTrashedTasksExist(networkCall.result.trashedTasksExist);
	}
}

async function createTask(request: CreateTaskRequest): Promise<void> {
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
			date: request.date,
			done: false
		});
	} else {
		undatedTaskStore.addOrUpdateTask({ ...request, ...networkCall.result, done: false });
	}
}

async function updateTask(request: UpdateTaskRequest): Promise<void> {
	const networkCall: NetworkCall<UpdateTaskResponse> = await jsonPutWithBodyWithResponse(
		'/api/tasks',
		request,
		UpdateTaskResponseSchema
	);
	if (networkCall.status !== 'success') return;

	const updatedTask = networkCall.result;
	if (updatedTask.done && updatedTask.date === undefined) {
		toast.error('An unexpected error occurred. Your data might be inconsistent. Please reload the page to resolve the issue.');
		return;
	}

	undatedTaskStore.removeTask(request.id);
	datedUntrashedOpenTaskStore.removeTask(request.id);
	datedUntrashedDoneTaskStore.removeTask(request.id);
	trashedTaskStore.removeTask(request.id);
	if (updatedTask.trashed) {
		trashedTaskStore.addOrUpdateTask(updatedTask);
	} else if (updatedTask.done && updatedTask.date !== undefined) {
		datedUntrashedDoneTaskStore.addOrUpdateTask({ ...updatedTask, date: updatedTask.date, done: true });
	} else if (updatedTask.date !== undefined) {
		datedUntrashedOpenTaskStore.addOrUpdateTask({ ...updatedTask, date: updatedTask.date, done: false });
	} else {
		undatedTaskStore.addOrUpdateTask({ ...updatedTask, done: false });
	}
}

async function toggleTaskDone(id: number): Promise<void> {
	const datedOpenTask = datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.find((task) => task.id === id);
	const datedDoneTask = datedUntrashedDoneTaskStore.datedUntrashedDoneTasks.find((task) => task.id === id);
	const undatedTask = undatedTaskStore.undatedTasks.find((task) => task.id === id);
	const trashedTask = trashedTaskStore.trashedTasks.find((task) => task.id === id);
	const task = datedOpenTask ?? datedDoneTask ?? undatedTask ?? trashedTask;
	if (!task) return;

	const networkCall: NetworkCall<ToggleTaskDoneResponse> = await jsonPatchWithoutBodyWithResponse(
		`/api/tasks/${id}/toggle-done`,
		ToggleTaskDoneResponseSchema
	);
	if (networkCall.status !== 'success') return;
	if (trashedTask !== undefined) {
		trashedTaskStore.addOrUpdateTask({
			...trashedTask,
			date: networkCall.result.date,
			done: !trashedTask.done
		});
		return;
	}
	if (networkCall.result.rank === undefined) {
		toast.error('An unexpected error occurred. Your data might be inconsistent. Please reload the page to resolve the issue.');
		return;
	}

	undatedTaskStore.removeTask(id);
	datedUntrashedOpenTaskStore.removeTask(id);
	datedUntrashedDoneTaskStore.removeTask(id);
	if (datedDoneTask !== undefined) {
		const openedTask = {
			...datedDoneTask,
			date: networkCall.result.date,
			rank: networkCall.result.rank,
			done: false as const
		};
		datedUntrashedOpenTaskStore.addOrUpdateTask(openedTask);
	} else {
		const completedTask = {
			...task,
			date: networkCall.result.date,
			rank: networkCall.result.rank,
			done: true as const
		};
		datedUntrashedDoneTaskStore.addOrUpdateTask(completedTask);
	}
}

async function completeTasks(taskIds: number[]): Promise<boolean> {
	const tasks = taskIds
		.map((taskId) => {
			const datedTask = datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.find((task) => task.id === taskId);
			const undatedTask = undatedTaskStore.undatedTasks.find((task) => task.id === taskId);
			return datedTask ?? undatedTask;
		})
		.filter((task) => task !== undefined);
	if (tasks.length !== taskIds.length) return false;

	const request: BatchUpdateDoneRequest = { taskIds };
	const networkCall: NetworkCall<BatchCompleteResponse> = await jsonPatchWithBodyWithResponse(
		'/api/tasks/batch-complete',
		request,
		BatchCompleteResponseSchema
	);
	if (networkCall.status !== 'success') return false;

	const completedTasks = networkCall.result
		.map((completedTask) => {
			const task = tasks.find((task) => task.id === completedTask.id);
			if (task === undefined) return undefined;
			return { ...task, date: completedTask.date, rank: completedTask.rank, done: true as const };
		})
		.filter((task) => task !== undefined);
	if (completedTasks.length !== taskIds.length) return false;

	undatedTaskStore.removeTasks(taskIds);
	datedUntrashedOpenTaskStore.removeTasks(taskIds);
	datedUntrashedDoneTaskStore.addOrUpdateTasks(completedTasks);
	return true;
}

async function openTasks(taskIds: number[]): Promise<boolean> {
	const tasks = taskIds
		.map((taskId) => datedUntrashedDoneTaskStore.datedUntrashedDoneTasks.find((task) => task.id === taskId))
		.filter((task) => task !== undefined);
	if (tasks.length !== taskIds.length) return false;

	const request: BatchUpdateDoneRequest = { taskIds };
	const networkCall: NetworkCall<BatchOpenResponse> = await jsonPatchWithBodyWithResponse(
		'/api/tasks/batch-open',
		request,
		BatchOpenResponseSchema
	);
	if (networkCall.status !== 'success') return false;

	const openedTasks = networkCall.result
		.map((openedTask) => {
			const task = tasks.find((task) => task.id === openedTask.id);
			if (task === undefined) return undefined;
			return { ...task, rank: openedTask.rank, done: false as const };
		})
		.filter((task) => task !== undefined);
	if (openedTasks.length !== taskIds.length) return false;

	datedUntrashedDoneTaskStore.removeTasks(taskIds);
	datedUntrashedOpenTaskStore.addOrUpdateTasks(openedTasks);
	return true;
}

async function trashTask(id: number): Promise<void> {
	const datedTask = datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.find((task) => task.id === id);
	const completedTask = datedUntrashedDoneTaskStore.datedUntrashedDoneTasks.find((task) => task.id === id);
	const undatedTask = undatedTaskStore.undatedTasks.find((task) => task.id === id);
	const task = datedTask ?? completedTask ?? undatedTask;
	if (!task) return;

	const networkCall: NetworkCall<TrashTaskResponse> = await jsonPatchWithoutBodyWithResponse(
		`/api/tasks/${id}/trash`,
		TrashTaskResponseSchema
	);
	if (networkCall.status !== 'success') return;

	undatedTaskStore.removeTask(id);
	datedUntrashedOpenTaskStore.removeTask(id);
	datedUntrashedDoneTaskStore.removeTask(id);
	trashedTaskStore.addOrUpdateTask({
		id: task.id,
		categoryId: task.categoryId,
		title: task.title,
		description: task.description,
		date: datedTask?.date ?? completedTask?.date,
		done: completedTask !== undefined,
		rank: networkCall.result.rank
	});
}

async function trashTasks(taskIds: number[]): Promise<boolean> {
	const tasks = taskIds
		.map((taskId) => {
			const datedOpenTask = datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.find((task) => task.id === taskId);
			const undatedOpenTask = undatedTaskStore.undatedTasks.find((task) => task.id === taskId);
			const datedDoneTask = datedUntrashedDoneTaskStore.datedUntrashedDoneTasks.find((task) => task.id === taskId);
			return datedOpenTask ?? undatedOpenTask ?? datedDoneTask;
		})
		.filter((task) => task !== undefined);
	if (tasks.length !== taskIds.length) return false;

	const request: BatchTrashTasksRequest = { taskIds };
	const networkCall: NetworkCall<BatchTrashResponse> = await jsonPatchWithBodyWithResponse(
		'/api/tasks/batch-trash',
		request,
		BatchTrashResponseSchema
	);
	if (networkCall.status !== 'success') return false;

	const trashedTasks: TrashedTaskResponse[] = networkCall.result
		.map((trashedTask) => {
			const task = tasks.find((task) => task.id === trashedTask.id);
			if (task === undefined) return undefined;
			return {
				id: task.id,
				categoryId: task.categoryId,
				title: task.title,
				description: task.description,
				date: 'date' in task ? task.date : undefined,
				done: task.done,
				rank: trashedTask.rank
			};
		})
		.filter((task) => task !== undefined);
	if (trashedTasks.length !== taskIds.length) return false;
	undatedTaskStore.removeTasks(taskIds);
	datedUntrashedOpenTaskStore.removeTasks(taskIds);
	datedUntrashedDoneTaskStore.removeTasks(taskIds);
	trashedTaskStore.addOrUpdateTasks(trashedTasks);
	return true;
}

async function restoreTask(id: number): Promise<void> {
	const task = trashedTaskStore.trashedTasks.find((task) => task.id === id);
	if (!task) return;

	const networkCall: NetworkCall<RestoreTaskResponse> = await jsonPatchWithoutBodyWithResponse(
		`/api/tasks/${id}/restore`,
		RestoreTaskResponseSchema
	);
	if (networkCall.status !== 'success') return;

	await fetchTrashedTasksExist();
	trashedTaskStore.removeTask(id);
	if (task.done && task.date !== undefined) {
		datedUntrashedDoneTaskStore.addOrUpdateTask({
			id: task.id,
			categoryId: task.categoryId,
			title: task.title,
			description: task.description,
			date: task.date,
			rank: networkCall.result.rank,
			done: true
		});
	} else if (task.date !== undefined) {
		datedUntrashedOpenTaskStore.addOrUpdateTask({
			id: task.id,
			categoryId: task.categoryId,
			title: task.title,
			description: task.description,
			date: task.date,
			rank: networkCall.result.rank,
			done: false
		});
	} else {
		undatedTaskStore.addOrUpdateTask({
			id: task.id,
			categoryId: task.categoryId,
			title: task.title,
			description: task.description,
			rank: networkCall.result.rank,
			done: false
		});
	}
}

async function deleteTrashedTasks(): Promise<void> {
	if (!trashedTaskStore.trashedTasksExist) return;
	const networkCall = await jsonDeleteWithoutBodyWithoutResponse('/api/tasks/trashed');
	if (networkCall.status !== 'success') return;

	trashedTaskStore.reset();
}

async function dragAndDropTask(id: number, request: TaskDragAndDropRequest): Promise<void> {
	const undatedTask = undatedTaskStore.undatedTasks.find((task) => task.id === id);
	const datedTask = datedUntrashedOpenTaskStore.datedUntrashedOpenTasks.find((task) => task.id === id);
	const completedTask = datedUntrashedDoneTaskStore.datedUntrashedDoneTasks.find((task) => task.id === id);
	const trashedTask = trashedTaskStore.trashedTasks.find((task) => task.id === id);
	const task = undatedTask ?? datedTask ?? completedTask ?? trashedTask;
	if (!task) return;

	const networkCall: NetworkCall<TaskDragAndDropResponse> = await jsonPatchWithBodyWithResponse(
		`/api/tasks/${id}/drag-and-drop`,
		request,
		TaskDragAndDropResponseSchema
	);
	if (networkCall.status === 'failed') return;

	if (request.trashed && completedTask !== undefined) {
		datedUntrashedDoneTaskStore.removeTask(id);
		trashedTaskStore.addOrUpdateTask({ ...completedTask, rank: networkCall.result.rank });
		return;
	}
	if (request.trashed && trashedTask !== undefined) {
		trashedTaskStore.addOrUpdateTask({ ...trashedTask, rank: networkCall.result.rank });
		return;
	}
	if (!request.trashed && trashedTask !== undefined && request.date !== undefined) {
		await fetchTrashedTasksExist();
		trashedTaskStore.removeTask(id);
		datedUntrashedDoneTaskStore.addOrUpdateTask({
			...trashedTask,
			date: request.date,
			done: true,
			rank: networkCall.result.rank
		});
		return;
	}
	if (completedTask !== undefined && request.date !== undefined) {
		datedUntrashedDoneTaskStore.addOrUpdateTask({
			...completedTask,
			date: request.date,
			rank: networkCall.result.rank
		});
		return;
	}

	const openTask = undatedTask ?? datedTask;
	if (openTask === undefined) return;
	undatedTaskStore.removeTask(id);
	datedUntrashedOpenTaskStore.removeTask(id);
	if (request.date) {
		datedUntrashedOpenTaskStore.addOrUpdateTask({
			...openTask,
			date: request.date,
			rank: networkCall.result.rank,
			done: false
		});
	} else {
		undatedTaskStore.addOrUpdateTask({ ...openTask, rank: networkCall.result.rank });
	}
}

export const taskService = {
	fetchDatedUntrashedDoneTasks,
	fetchUndatedTasks,
	fetchDatedUntrashedOpenTodayTasks,
	fetchDatedUntrashedDoneTodayTasks,
	fetchDatedUntrashedOpenNonTodayTasks,
	fetchTrashedTasks,
	fetchTrashedTasksExist,
	createTask,
	updateTask,
	toggleTaskDone,
	completeTasks,
	openTasks,
	trashTask,
	trashTasks,
	restoreTask,
	deleteTrashedTasks,
	dragAndDropTask
};
