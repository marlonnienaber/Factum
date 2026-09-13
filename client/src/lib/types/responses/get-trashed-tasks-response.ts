import { TrashedTaskResponseSchema } from '$lib/types/responses/trashed-task-response';
import { z } from 'zod';

export const GetTrashedTasksResponseSchema = z.object({
	tasks: z.array(TrashedTaskResponseSchema),
	hasMore: z.boolean()
});

export type GetTrashedTasksResponse = z.infer<typeof GetTrashedTasksResponseSchema>;
