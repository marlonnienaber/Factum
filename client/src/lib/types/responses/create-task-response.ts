import { z } from 'zod';

export const CreateTaskResponseSchema = z.strictObject({
	id: z.number().int().positive(),
	rank: z.string().min(1)
});

export type CreateTaskResponse = z.infer<typeof CreateTaskResponseSchema>;
