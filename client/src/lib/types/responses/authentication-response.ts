import { z } from 'zod';

export const AuthenticationResponseSchema = z.strictObject({
	userId: z.number()
});

export type AuthenticationResponse = z.infer<typeof AuthenticationResponseSchema>;
