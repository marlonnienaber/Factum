import { z } from 'zod';

export const GetTodayProgressResponseSchema = z.strictObject({
	openTaskCount: z.number().int().nonnegative(),
	doneTaskCount: z.number().int().nonnegative()
});

export type GetTodayProgressResponse = z.infer<typeof GetTodayProgressResponseSchema>;
