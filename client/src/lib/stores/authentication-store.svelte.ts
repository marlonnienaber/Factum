import { jsonGet, jsonPostWithoutBodyWithoutResponse, jsonPostWithResponse } from '$lib/api/api';
import type { AuthenticationRequest } from '$lib/types/requests/authentication-request';
import { AuthenticationResponseSchema, type AuthenticationResponse } from '../types/responses/authentication-response';

let authentication = $state<AuthenticationResponse | undefined>(undefined);
let loading = $state(false);
let fetchAttempted = $state(false);

async function signIn(request: AuthenticationRequest) {
	loading = true;

	try {
		authentication = await jsonPostWithResponse('/api/sign-in', request, AuthenticationResponseSchema);
	} finally {
		fetchAttempted = true;
		loading = false;
	}
}

async function signUp(request: AuthenticationRequest) {
	loading = true;
	try {
		authentication = await jsonPostWithResponse('/api/sign-up', request, AuthenticationResponseSchema);
	} finally {
		fetchAttempted = true;
		loading = false;
	}
}

async function signOut() {
	loading = true;
	try {
		await jsonPostWithoutBodyWithoutResponse('/api/sign-out');
		authentication = undefined;
	} finally {
		fetchAttempted = true;
		loading = false;
	}
}

async function fetchCurrentAuthenticationIfNecessary() {
	if (authentication !== undefined || loading) return;
	loading = true;
	try {
		authentication = await jsonGet('/api/current-authentication', AuthenticationResponseSchema);
	} finally {
		fetchAttempted = true;
		loading = false;
	}
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
