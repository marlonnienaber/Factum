import { z } from 'zod';

export const AuthenticationResponseSchema = z.object({
	userId: z.number()
});

export type AuthenticationResponse = z.infer<typeof AuthenticationResponseSchema>;
