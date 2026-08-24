import type { AuthenticationResponse } from '../types/responses/authentication-response';

let authentication = $state<AuthenticationResponse | undefined>(undefined);

function setAuthentication(newAuthentication: AuthenticationResponse): void {
	authentication = newAuthentication;
}

function clearAuthentication(): void {
	authentication = undefined;
}

export const authenticationStore = {
	get authentication() {
		return authentication;
	},
	setAuthentication,
	clearAuthentication
};
