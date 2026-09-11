import { parseDate } from '@internationalized/date';
import { z } from 'zod';

export const ToggleTaskDoneResponseSchema = z.object({
	date: z.iso.date().transform((date) => parseDate(date)),
	rank: z.string().min(1).optional()
});

export type ToggleTaskDoneResponse = z.infer<typeof ToggleTaskDoneResponseSchema>;
