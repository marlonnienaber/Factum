import { z } from 'zod';

export const GetTrashedTasksExistResponseSchema = z.strictObject({
	trashedTasksExist: z.boolean()
});

export type GetTrashedTasksExistResponse = z.infer<typeof GetTrashedTasksExistResponseSchema>;
