import { z } from 'zod';

export const UndatedTaskResponseSchema = z
	.object({
		id: z.number().int().positive(),
		categoryId: z.number().int().positive(),
		title: z.string().min(1).max(100),
		description: z.string().min(1).max(500).optional(),
		rank: z.string().min(1)
	})
	.transform((task) => ({ ...task, done: false as const }));

export type UndatedTaskResponse = z.infer<typeof UndatedTaskResponseSchema>;
