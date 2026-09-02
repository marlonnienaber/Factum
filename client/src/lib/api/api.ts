import { browser } from '$app/environment';
import { toast } from 'svelte-sonner';
import { z, type ZodType } from 'zod';
import { ErrorResponseSchema } from '../types/responses/error-response';

export type NetworkCall<T> = { status: 'success'; result: T } | { status: 'failed' };

export async function jsonGetWithoutBodyWithResponse<R>(url: string, responseSchema: ZodType<R>): Promise<NetworkCall<R>> {
	return await jsonFetchWithoutBodyWithResponse('GET', url, responseSchema);
}

export async function jsonPostWithoutBodyWithResponse<R>(url: string, responseSchema: ZodType<R>): Promise<NetworkCall<R>> {
	return await jsonFetchWithoutBodyWithResponse('POST', url, responseSchema);
}

export async function jsonPostWithBodyWithResponse<B, R>(
	url: string,
	requestBody: B,
	responseSchema: ZodType<R>
): Promise<NetworkCall<R>> {
	return await jsonFetchWithBodyWithResponse('POST', url, requestBody, responseSchema);
}

export async function jsonPutWithBodyWithResponse<B, R>(
	url: string,
	requestBody: B,
	responseSchema: ZodType<R>
): Promise<NetworkCall<R>> {
	return await jsonFetchWithBodyWithResponse('PUT', url, requestBody, responseSchema);
}

export async function jsonPatchWithBodyWithResponse<B, R>(
	url: string,
	requestBody: B,
	responseSchema: ZodType<R>
): Promise<NetworkCall<R>> {
	return await jsonFetchWithBodyWithResponse('PATCH', url, requestBody, responseSchema);
}

export async function jsonPatchWithBodyWithoutResponse<B>(url: string, requestBody: B): Promise<NetworkCall<void>> {
	return await jsonFetchWithBodyWithoutResponse('PATCH', url, requestBody);
}

export async function jsonPostWithoutBodyWithoutResponse(url: string): Promise<NetworkCall<void>> {
	return await jsonFetchWithoutBodyWithoutResponse('POST', url);
}

export async function jsonPatchWithoutBodyWithoutResponse(url: string): Promise<NetworkCall<void>> {
	return await jsonFetchWithoutBodyWithoutResponse('PATCH', url);
}

export async function jsonDeleteWithoutBodyWithoutResponse(url: string): Promise<NetworkCall<void>> {
	return await jsonFetchWithoutBodyWithoutResponse('DELETE', url);
}

async function jsonFetchWithoutBodyWithResponse<R>(
	httpMethod: 'GET' | 'POST',
	url: string,
	responseSchema: ZodType<R>
): Promise<NetworkCall<R>> {
	let requestOptions: RequestInit;
	try {
		requestOptions = configureRequestOptions(httpMethod);
	} catch (error) {
		return handleError(error, 'Failed to configure request options.');
	}

	let response: Response;
	try {
		response = await fetch(url, requestOptions);
	} catch (error) {
		return handleError(error, 'Request failed before receiving a response.');
	}

	if (!response.ok) {
		let errorMessage: string;
		try {
			errorMessage = await getErrorResponseMessage(response);
		} catch (error) {
			return handleError(error, 'Failed to parse error response message.');
		}
		return handleErrorResponse(errorMessage);
	}

	try {
		const result = await parseResponse(response, responseSchema);
		return { status: 'success', result: result };
	} catch (error) {
		const message = error instanceof z.ZodError ? 'Response validation failed.' : 'Failed to parse response body.';
		return handleError(error, message);
	}
}

async function jsonFetchWithBodyWithResponse<B, R>(
	httpMethod: 'POST' | 'PUT' | 'PATCH',
	url: string,
	requestBody: B,
	responseSchema: ZodType<R>
): Promise<NetworkCall<R>> {
	let requestOptions: RequestInit;
	try {
		requestOptions = configureRequestOptions(httpMethod, requestBody);
	} catch (error) {
		return handleError(error, 'Failed to configure request.');
	}

	let response: Response;
	try {
		response = await fetch(url, requestOptions);
	} catch (error) {
		return handleError(error, 'Request failed before receiving a response.');
	}

	if (!response.ok) {
		let errorMessage: string;
		try {
			errorMessage = await getErrorResponseMessage(response);
		} catch (error) {
			return handleError(error, 'Failed to parse error response.');
		}
		return handleErrorResponse(errorMessage);
	}

	try {
		const result = await parseResponse(response, responseSchema);
		return { status: 'success', result: result };
	} catch (error) {
		const message = error instanceof z.ZodError ? 'Response validation failed.' : 'Failed to parse response body.';
		return handleError(error, message);
	}
}

async function jsonFetchWithoutBodyWithoutResponse(
	httpMethod: 'POST' | 'PATCH' | 'DELETE',
	url: string
): Promise<NetworkCall<void>> {
	let requestOptions: RequestInit;
	try {
		requestOptions = configureRequestOptions(httpMethod);
	} catch (error) {
		return handleError(error, 'Failed to configure request.');
	}

	let response: Response;
	try {
		response = await fetch(url, requestOptions);
	} catch (error) {
		return handleError(error, 'Request failed before receiving a response.');
	}

	if (!response.ok) {
		let errorMessage: string;
		try {
			errorMessage = await getErrorResponseMessage(response);
		} catch (error) {
			return handleError(error, 'Failed to parse error response.');
		}
		return handleErrorResponse(errorMessage);
	}

	return { status: 'success', result: undefined };
}

async function jsonFetchWithBodyWithoutResponse<B>(httpMethod: 'PATCH', url: string, requestBody: B): Promise<NetworkCall<void>> {
	let requestOptions: RequestInit;
	try {
		requestOptions = configureRequestOptions(httpMethod, requestBody);
	} catch (error) {
		return handleError(error, 'Failed to configure request.');
	}

	let response: Response;
	try {
		response = await fetch(url, requestOptions);
	} catch (error) {
		return handleError(error, 'Request failed before receiving a response.');
	}

	if (!response.ok) {
		let errorMessage: string;
		try {
			errorMessage = await getErrorResponseMessage(response);
		} catch (error) {
			return handleError(error, 'Failed to parse error response.');
		}
		return handleErrorResponse(errorMessage);
	}

	return { status: 'success', result: undefined };
}

function configureRequestOptions<B>(method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE', requestBody?: B): RequestInit {
	const headers = new Headers();

	if (requestBody !== undefined) {
		headers.set('Content-Type', 'application/json');
	}

	return {
		method,
		credentials: 'same-origin',
		headers,
		...(requestBody !== undefined && { body: JSON.stringify(requestBody) })
	};
}

async function parseResponse<R>(response: Response, responseSchema: ZodType<R>): Promise<R> {
	if (isPrimitiveResponseSchema(responseSchema)) {
		const text = await response.text();
		return responseSchema.parse(coerceResponseToPrimitiveType(responseSchema, text));
	}
	return responseSchema.parse(await response.json());
}

async function getErrorResponseMessage(response: Response): Promise<string> {
	const errorResponse = ErrorResponseSchema.parse(await response.json());
	return errorResponse.message;
}

function handleError(error: unknown, message: string): NetworkCall<never> {
	console.error(error);
	showErrorToast(message);
	return { status: 'failed' };
}

function handleErrorResponse(message: string): NetworkCall<never> {
	console.error(message);
	showErrorToast(message);
	return { status: 'failed' };
}

function showErrorToast(message: string): void {
	if (!browser) return;
	toast.error(message);
}

function isPrimitiveResponseSchema(schema: ZodType): boolean {
	return schema instanceof z.ZodString || schema instanceof z.ZodNumber || schema instanceof z.ZodBoolean;
}

function coerceResponseToPrimitiveType<T>(schema: ZodType<T>, text: string): T {
	if (schema instanceof z.ZodString) return text as T;
	if (schema instanceof z.ZodNumber) return Number(text) as T;
	if (schema instanceof z.ZodBoolean) return (text === 'true') as T;
	throw new Error('Unsupported primitive response schema.');
}
