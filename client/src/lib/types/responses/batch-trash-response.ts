import { z } from 'zod';

export const BatchTrashResponseSchema = z.array(
	z.strictObject({
		id: z.number().int().positive(),
		rank: z.string().min(1)
	})
);

export type BatchTrashResponse = z.infer<typeof BatchTrashResponseSchema>;
