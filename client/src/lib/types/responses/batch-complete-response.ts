import { DatedUntrashedDoneTaskResponseSchema } from '$lib/types/responses/dated-untrashed-done-task-response';
import { TrashedTaskResponseSchema } from '$lib/types/responses/trashed-task-response';
import { z } from 'zod';

export const BatchCompleteResponseSchema = z.strictObject({
	datedUntrashedDoneTasks: z.array(DatedUntrashedDoneTaskResponseSchema),
	trashedTasks: z.array(TrashedTaskResponseSchema)
});

export type BatchCompleteResponse = z.infer<typeof BatchCompleteResponseSchema>;
