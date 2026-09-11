import { UndatedTaskResponseSchema } from '$lib/types/responses/undated-task-response';
import { z } from 'zod';

export const GetUndatedTasksResponseSchema = z.object({
	tasks: z.array(UndatedTaskResponseSchema),
	hasMore: z.boolean()
});

export type GetUndatedTasksResponse = z.infer<typeof GetUndatedTasksResponseSchema>;
