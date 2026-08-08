import { z } from 'zod';

export const signInInputSchema = z.object({
	email: z.email(),
	password: z.string().min(10).max(128)
});

export type SignInInput = z.infer<typeof signInInputSchema>;
