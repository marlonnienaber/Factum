import { z } from 'zod';

export const BatchOpenResponseSchema = z.array(
	z.strictObject({
		id: z.number().int().positive(),
		rank: z.string().min(1)
	})
);

export type BatchOpenResponse = z.infer<typeof BatchOpenResponseSchema>;
