import { z } from 'zod';

export const CreateCategoryResponseSchema = z.strictObject({
	id: z.number().int().positive()
});

export type CreateCategoryResponse = z.infer<typeof CreateCategoryResponseSchema>;
