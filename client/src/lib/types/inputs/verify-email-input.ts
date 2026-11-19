import { z } from 'zod';

export const verifyEmailInputSchema = z.strictObject({
	email: z.email('Provide a valid email address.')
});

export type VerifyEmailInput = z.infer<typeof verifyEmailInputSchema>;
