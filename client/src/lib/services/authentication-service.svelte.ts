import {
	type NetworkCall,
	jsonGetWithoutBodyWithResponse,
	jsonPostWithoutBodyWithoutResponse,
	jsonPostWithBodyWithoutResponse,
	jsonPostWithBodyWithResponse
} from '$lib/utils/networking';
import { authenticationStore } from '$lib/stores/authentication-store.svelte';
import type { SignInRequest } from '$lib/types/requests/sign-in-request';
import type { SignUpRequest } from '$lib/types/requests/sign-up-request';
import type { VerifyEmailRequest } from '$lib/types/requests/verify-email-request';
import { AuthenticationResponseSchema, type AuthenticationResponse } from '$lib/types/responses/authentication-response';

let loading = $state(false);
let fetchAttempted = $state(false);

async function signIn(request: SignInRequest): Promise<void> {
	if (loading) return;
	loading = true;
	try {
		const networkCall: NetworkCall<AuthenticationResponse> = await jsonPostWithBodyWithResponse(
			'/api/sign-in',
			request,
			AuthenticationResponseSchema
		);
		if (networkCall.status === 'success') authenticationStore.setAuthentication(networkCall.result);
		fetchAttempted = true;
	} finally {
		loading = false;
	}
}

async function verifyEmail(request: VerifyEmailRequest): Promise<boolean> {
	if (loading) return false;
	loading = true;
	try {
		const networkCall: NetworkCall<void> = await jsonPostWithBodyWithoutResponse('/api/verify-email', request);
		return networkCall.status === 'success';
	} finally {
		loading = false;
	}
}

async function signUp(request: SignUpRequest): Promise<boolean> {
	if (loading) return false;
	loading = true;
	try {
		const networkCall: NetworkCall<void> = await jsonPostWithBodyWithoutResponse('/api/sign-up', request);
		return networkCall.status === 'success';
	} finally {
		loading = false;
	}
}

async function signOut(): Promise<void> {
	if (loading) return;
	loading = true;
	try {
		const networkCall: NetworkCall<void> = await jsonPostWithoutBodyWithoutResponse('/api/sign-out');
		if (networkCall.status === 'success') authenticationStore.clearAuthentication();
	} finally {
		loading = false;
	}
}

async function fetchCurrentAuthenticationIfNecessary(): Promise<void> {
	if (authenticationStore.authentication !== undefined || loading) return;
	loading = true;
	try {
		const networkCall: NetworkCall<AuthenticationResponse> = await jsonGetWithoutBodyWithResponse(
			'/api/current-authentication',
			AuthenticationResponseSchema
		);
		if (networkCall.status === 'success') authenticationStore.setAuthentication(networkCall.result);
		fetchAttempted = true;
	} finally {
		loading = false;
	}
}

export const authenticationService = {
	get loading(): boolean {
		return loading;
	},
	get fetchAttempted(): boolean {
		return fetchAttempted;
	},
	signIn,
	signUp,
	verifyEmail,
	signOut,
	fetchCurrentAuthenticationIfNecessary
};
