import { z } from 'zod';

export const CreateTaskResponseSchema = z.object({
	id: z.number().int().positive(),
	rank: z.string().min(1)
});

export type CreateTaskResponse = z.infer<typeof CreateTaskResponseSchema>;
