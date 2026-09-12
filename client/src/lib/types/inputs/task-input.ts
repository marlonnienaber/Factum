import { z } from 'zod';

export const taskInputSchema = z.object({
	title: z.string().min(1, 'Title must have at least 1 character').max(100, 'Title must have at most 100 characters'),
	description: z.preprocess(
		(value) => (value === '' ? undefined : value),
		z.string().max(500, 'Description must have at most 500 characters').optional()
	),
	categoryId: z.number({ error: 'Category is required' }).int().positive()
});

export type TaskInput = z.infer<typeof taskInputSchema>;
