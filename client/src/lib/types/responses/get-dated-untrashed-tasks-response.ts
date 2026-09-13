import { DatedUntrashedTaskResponseSchema } from '$lib/types/responses/dated-untrashed-task-response';
import { z } from 'zod';

export const GetDatedUntrashedTasksResponseSchema = z.object({
	tasks: z.array(DatedUntrashedTaskResponseSchema),
	hasMore: z.boolean()
});

export type GetDatedUntrashedTasksResponse = z.infer<typeof GetDatedUntrashedTasksResponseSchema>;
