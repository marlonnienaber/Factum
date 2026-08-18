import { z } from 'zod';

export const CategoryResponseSchema = z.object({
	id: z.number().int().positive(),
	title: z.string().min(1).max(50)
});

export type CategoryResponse = z.infer<typeof CategoryResponseSchema>;
