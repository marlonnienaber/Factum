import { parseDate } from '@internationalized/date';
import { z } from 'zod';

export const BatchCompleteResponseSchema = z.array(
	z.strictObject({
		id: z.number().int().positive(),
		date: z.iso.date().transform((date) => parseDate(date)),
		rank: z.string().min(1)
	})
);

export type BatchCompleteResponse = z.infer<typeof BatchCompleteResponseSchema>;
