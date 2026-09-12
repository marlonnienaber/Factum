import { z } from 'zod';

export const signUpInputSchema = z
	.object({
		email: z.email(),
		password: z
			.string()
			.min(10, 'Password must have at least 10 characters')
			.max(128, 'Password must have at most 128 characters'),
		repeatedPassword: z
			.string()
			.min(10, 'Password must have at least 10 characters')
			.max(128, 'Password must have at most 128 characters')
	})
	.refine((input) => input.password === input.repeatedPassword, {
		message: 'Passwords do not match.',
		path: ['repeatedPassword']
	});

export type SignUpInput = z.infer<typeof signUpInputSchema>;
