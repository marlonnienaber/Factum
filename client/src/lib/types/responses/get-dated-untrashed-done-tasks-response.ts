import { DatedUntrashedDoneTaskResponseSchema } from '$lib/types/responses/dated-untrashed-done-task-response';
import { z } from 'zod';

export const GetDatedUntrashedDoneTasksResponseSchema = z.strictObject({
	tasks: z.array(DatedUntrashedDoneTaskResponseSchema),
	hasMore: z.boolean()
});

export type GetDatedUntrashedDoneTasksResponse = z.infer<typeof GetDatedUntrashedDoneTasksResponseSchema>;
