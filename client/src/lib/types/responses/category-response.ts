import { z } from 'zod';

export const CategoryResponseSchema = z.object({
	id: z.number().int().positive(),
	title: z.string().min(1).max(50),
	color: z.string().regex(/^#[0-9a-f]{6}$/)
});

export type CategoryResponse = z.infer<typeof CategoryResponseSchema>;
