import { taskService } from '$lib/services/task-service.svelte';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	await taskService.fetchDatedUntrashedDoneTasks();
};
