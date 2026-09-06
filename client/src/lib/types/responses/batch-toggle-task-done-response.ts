import { DatedUntrashedDoneTaskResponseSchema } from '$lib/types/responses/dated-untrashed-done-task-response';
import { DatedUntrashedOpenTaskResponseSchema } from '$lib/types/responses/dated-untrashed-open-task-response';
import { TrashedTaskResponseSchema } from '$lib/types/responses/trashed-task-response';
import { z } from 'zod';

export const BatchToggleTaskDoneResponseSchema = z.object({
	datedUntrashedOpenTasks: z.array(DatedUntrashedOpenTaskResponseSchema),
	datedUntrashedDoneTasks: z.array(DatedUntrashedDoneTaskResponseSchema),
	trashedTasks: z.array(TrashedTaskResponseSchema)
});

export type BatchToggleTaskDoneResponse = z.infer<typeof BatchToggleTaskDoneResponseSchema>;
