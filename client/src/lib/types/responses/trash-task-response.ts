import { z } from 'zod';

export const TrashTaskResponseSchema = z.strictObject({
	rank: z.string().min(1)
});

export type TrashTaskResponse = z.infer<typeof TrashTaskResponseSchema>;
