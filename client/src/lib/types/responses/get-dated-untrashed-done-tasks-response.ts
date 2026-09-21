import { DatedUntrashedDoneTaskResponseSchema } from '$lib/types/responses/dated-untrashed-done-task-response';
import { z } from 'zod';

export const GetDatedUntrashedDoneTasksResponseSchema = z.object({
	tasks: z.array(DatedUntrashedDoneTaskResponseSchema),
	hasMore: z.boolean()
});

export type GetDatedUntrashedDoneTasksResponse = z.infer<typeof GetDatedUntrashedDoneTasksResponseSchema>;
