import { z } from 'zod';

export const createTaskInputSchema = z.object({
	title: z.string().min(1).max(100),
	description: z.preprocess((value) => (value === '' ? undefined : value), z.string().max(500).optional()),
	categoryId: z.number().int().positive()
});

export type CreateTaskInput = z.infer<typeof createTaskInputSchema>;
