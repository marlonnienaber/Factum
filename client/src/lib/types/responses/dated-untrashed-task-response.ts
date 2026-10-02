import { parseDate } from '@internationalized/date';
import { z } from 'zod';

export const DatedUntrashedTaskResponseSchema = z.strictObject({
	id: z.number().int().positive(),
	categoryId: z.number().int().positive(),
	title: z.string().min(1).max(100),
	description: z.string().min(1).max(500).optional(),
	date: z.iso.date().transform((date) => parseDate(date)),
	rank: z.string().min(1)
});

export type DatedUntrashedTaskResponse = z.infer<typeof DatedUntrashedTaskResponseSchema>;
