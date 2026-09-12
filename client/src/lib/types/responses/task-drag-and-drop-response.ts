import { z } from 'zod';

export const TaskDragAndDropResponseSchema = z.object({
	rank: z.string().min(1)
});

export type TaskDragAndDropResponse = z.infer<typeof TaskDragAndDropResponseSchema>;
