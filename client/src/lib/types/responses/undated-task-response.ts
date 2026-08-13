import { z } from 'zod';

export const UndatedTaskResponseSchema = z.object({
	id: z.number().int().positive(),
	categoryId: z.number().int().positive(),
	title: z.string().min(1).max(100),
	description: z.string().min(1).max(2000).optional(),
	rank: z.string().min(1),
});

export type UndatedTaskResponse = z.infer<typeof UndatedTaskResponseSchema>;
