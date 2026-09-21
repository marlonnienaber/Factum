import { DatedUntrashedTaskResponseSchema } from '$lib/types/responses/dated-untrashed-task-response';
import { z } from 'zod';

export const DatedUntrashedDoneTaskResponseSchema = DatedUntrashedTaskResponseSchema.transform((task) => ({
	...task,
	done: true as const
}));

export type DatedUntrashedDoneTaskResponse = z.infer<typeof DatedUntrashedDoneTaskResponseSchema>;
