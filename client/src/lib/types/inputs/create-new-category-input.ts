import { z } from 'zod';

export const createNewCategoryInputSchema = z.strictObject({
	title: z.string().min(1).max(50)
});

export type CreateNewCategoryInput = z.infer<typeof createNewCategoryInputSchema>;
