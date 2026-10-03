import { categoryService } from '$lib/services/category-service.svelte';
import { taskService } from '$lib/services/task-service.svelte';
import type { LayoutLoad } from './$types';

export const load: LayoutLoad = async () => {
	await Promise.all([
		categoryService.fetchCategories(),
		taskService.fetchUndatedTasks(),
		taskService.fetchDatedUntrashedOpenNonTodayTasks(),
		taskService.fetchTrashedTasks(),
		taskService.fetchTrashedTasksExist()
	]);
};
