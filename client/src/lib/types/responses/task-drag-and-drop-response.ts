import { z } from 'zod';

export const TaskDragAndDropResponseSchema = z.strictObject({
	rank: z.string().min(1)
});

export type TaskDragAndDropResponse = z.infer<typeof TaskDragAndDropResponseSchema>;
