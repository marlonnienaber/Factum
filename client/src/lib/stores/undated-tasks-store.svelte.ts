import { jsonGetWithoutBodyWithResponse } from '$lib/api/api';
import { UndatedTaskResponseSchema, type UndatedTaskResponse } from '$lib/types/responses/undated-task-response';
import { z } from 'zod';

let undatedTasks = $state<UndatedTaskResponse[]>([]);
let loading = $state<boolean>(false);

async function fetchUndatedTasks(rank?: string): Promise<void> {
	loading = true;

	try {
		const fetchedTasks = await jsonGetWithoutBodyWithResponse(
			`/api/tasks/undated${rank ? `?rank=${encodeURIComponent(rank)}` : ''}`,
			z.array(UndatedTaskResponseSchema)
		);
		if (fetchedTasks === undefined) return;

		const fetchedIds = new Set(fetchedTasks.map((task) => task.id));
		const deduplicatedTasks = undatedTasks.filter((task) => !fetchedIds.has(task.id));
		undatedTasks = [...deduplicatedTasks, ...fetchedTasks].sort(compareTasks);
	} finally {
		loading = false;
	}
}

function compareTasks(first: UndatedTaskResponse, second: UndatedTaskResponse): number {
	if (first.rank !== second.rank) {
		return first.rank < second.rank ? -1 : 1;
	}

	return first.id - second.id;
}

export const undatedTaskStore = {
	get undatedTasks() {
		return undatedTasks;
	},
	get loading() {
		return loading;
	},
	fetchUndatedTasks
};
