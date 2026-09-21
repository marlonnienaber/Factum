import { z } from 'zod';

export const RestoreTaskResponseSchema = z.strictObject({
	rank: z.string().min(1)
});

export type RestoreTaskResponse = z.infer<typeof RestoreTaskResponseSchema>;
