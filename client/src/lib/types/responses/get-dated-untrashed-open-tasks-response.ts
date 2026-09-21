import { DatedUntrashedOpenTaskResponseSchema } from '$lib/types/responses/dated-untrashed-open-task-response';
import { z } from 'zod';

export const GetDatedUntrashedOpenTasksResponseSchema = z.strictObject({
	tasks: z.array(DatedUntrashedOpenTaskResponseSchema),
	hasMore: z.boolean()
});

export type GetDatedUntrashedOpenTasksResponse = z.infer<typeof GetDatedUntrashedOpenTasksResponseSchema>;
