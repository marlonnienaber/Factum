import { z } from 'zod';

export const signUpInputSchema = z
	.object({
		email: z.email(),
		password: z.string().min(10).max(128),
		repeatedPassword: z.string().min(10).max(128)
	})
	.refine((input) => input.password === input.repeatedPassword, {
		message: 'Passwords do not match.',
		path: ['repeatedPassword']
	});

export type SignUpInput = z.infer<typeof signUpInputSchema>;
