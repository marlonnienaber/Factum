import { z } from 'zod';

export const signUpInputSchema = z
	.object({
		token: z
			.string()
			.regex(
				/^[0-9a-fA-F]{64}$/,
				'This sign-up link is invalid. Please use the exact link from our latest email without changing it.'
			),
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
