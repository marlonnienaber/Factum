import { parseAbsoluteToLocal, parseDate } from '@internationalized/date';
import { z } from 'zod';

export const UpdateTaskResponseSchema = z.strictObject({
	id: z.number().int().positive(),
	categoryId: z.number().int().positive(),
	title: z.string().min(1).max(100),
	description: z.string().min(1).max(500).optional(),
	date: z.iso
		.date()
		.transform((date) => parseDate(date))
		.optional(),
	rank: z.string().min(1).optional(),
	done: z.boolean(),
	trashedDate: z.iso
		.datetime()
		.transform((trashedDate) => parseAbsoluteToLocal(trashedDate))
		.optional()
});

export type UpdateTaskResponse = z.infer<typeof UpdateTaskResponseSchema>;
