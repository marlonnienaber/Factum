import { DatedUntrashedTaskResponseSchema } from '$lib/types/responses/dated-untrashed-task-response';
import { TrashedTaskResponseSchema } from '$lib/types/responses/trashed-task-response';
import { z } from 'zod';

export const BatchCompleteResponseSchema = z.object({
	datedUntrashedDoneTasks: z.array(DatedUntrashedTaskResponseSchema),
	trashedTasks: z.array(TrashedTaskResponseSchema)
});

export type BatchCompleteResponse = z.infer<typeof BatchCompleteResponseSchema>;
