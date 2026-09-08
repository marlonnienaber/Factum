import {
	type NetworkCall,
	jsonGetWithoutBodyWithResponse,
	jsonPostWithoutBodyWithoutResponse,
	jsonPostWithBodyWithResponse
} from '$lib/utils/networking';
import { authenticationStore } from '$lib/stores/authentication-store.svelte';
import type { AuthenticationRequest } from '$lib/types/requests/authentication-request';
import { AuthenticationResponseSchema, type AuthenticationResponse } from '$lib/types/responses/authentication-response';

let loading = $state(false);
let fetchAttempted = $state(false);

async function signIn(request: AuthenticationRequest): Promise<void> {
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

async function signUp(request: AuthenticationRequest): Promise<void> {
	if (loading) return;
	loading = true;
	try {
		const networkCall: NetworkCall<AuthenticationResponse> = await jsonPostWithBodyWithResponse(
			'/api/sign-up',
			request,
			AuthenticationResponseSchema
		);
		if (networkCall.status === 'success') authenticationStore.setAuthentication(networkCall.result);
		fetchAttempted = true;
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
	signOut,
	fetchCurrentAuthenticationIfNecessary
};
