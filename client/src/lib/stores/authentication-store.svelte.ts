import { type NetworkCall, jsonGetWithoutBodyWithResponse, jsonPostWithoutBodyWithoutResponse, jsonPostWithBodyWithResponse } from '$lib/api/api';
import type { AuthenticationRequest } from '$lib/types/requests/authentication-request';
import { AuthenticationResponseSchema, type AuthenticationResponse } from '../types/responses/authentication-response';

let authentication = $state<AuthenticationResponse | undefined>(undefined);
let loading = $state(false);
let fetchAttempted = $state(false);

async function signIn(request: AuthenticationRequest) {
	loading = true;
	const networkCall: NetworkCall<AuthenticationResponse> = await jsonPostWithBodyWithResponse('/api/sign-in', request, AuthenticationResponseSchema);
	if (networkCall.status === 'success') authentication = networkCall.result;
	fetchAttempted = true;
	loading = false;
}

async function signUp(request: AuthenticationRequest) {
	loading = true;
	const networkCall: NetworkCall<AuthenticationResponse> = await jsonPostWithBodyWithResponse('/api/sign-up', request, AuthenticationResponseSchema);
	if (networkCall.status === 'success') authentication = networkCall.result;
	loading = false;
}

async function signOut() {
	loading = true;
	const networkCall: NetworkCall<void> = await jsonPostWithoutBodyWithoutResponse('/api/sign-out');
	if (networkCall.status === 'success') authentication = undefined;
	fetchAttempted = true;
	loading = false;
}

async function fetchCurrentAuthenticationIfNecessary() {
	if (authentication !== undefined || loading) return;
	loading = true;
	const networkCall: NetworkCall<AuthenticationResponse> = await jsonGetWithoutBodyWithResponse('/api/current-authentication', AuthenticationResponseSchema);
	if (networkCall.status === 'success') authentication = networkCall.result;
	fetchAttempted = true;
	loading = false;
}

export const authenticationStore = {
	get authentication() {
		return authentication;
	},
	get loading() {
		return loading;
	},
	get fetchAttempted() {
		return fetchAttempted;
	},
	signIn,
	signUp,
	signOut,
	fetchCurrentAuthenticationIfNecessary
};
