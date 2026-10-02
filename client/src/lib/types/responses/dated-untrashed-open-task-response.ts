import { DatedUntrashedTaskResponseSchema } from '$lib/types/responses/dated-untrashed-task-response';
import { z } from 'zod';

export const DatedUntrashedOpenTaskResponseSchema = DatedUntrashedTaskResponseSchema.transform((task) => ({
	...task,
	done: false as const
}));

export type DatedUntrashedOpenTaskResponse = z.infer<typeof DatedUntrashedOpenTaskResponseSchema>;
