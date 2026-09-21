import { z } from 'zod';

export const signInInputSchema = z.strictObject({
	email: z.email(),
	password: z.string().min(10, 'Password must have at least 10 characters').max(128, 'Password must have at most 128 characters')
});

export type SignInInput = z.infer<typeof signInInputSchema>;
