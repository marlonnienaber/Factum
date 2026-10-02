import { z } from 'zod';

export const ErrorResponseSchema = z.strictObject({
	errorCode: z.string(),
	message: z.string()
});

export type ErrorResponse = z.infer<typeof ErrorResponseSchema>;
