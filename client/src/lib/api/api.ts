import { browser } from '$app/environment';
import { toast } from 'svelte-sonner';
import { z, type ZodType } from 'zod';
import { ErrorResponseSchema } from '../types/responses/error-response';

export async function jsonGetWithoutBodyWithResponse<R>(url: string, responseSchema: ZodType<R>): Promise<R> {
	return await jsonFetchWithoutBodyWithResponse('GET', url, responseSchema);
}

export async function jsonPostWithoutBodyWithResponse<R>(url: string, responseSchema: ZodType<R>): Promise<R> {
	return await jsonFetchWithoutBodyWithResponse('POST', url, responseSchema);
}

export async function jsonPostWithBodyWithResponse<B, R>(url: string, requestBody: B, responseSchema: ZodType<R>): Promise<R> {
	return await jsonFetchWithBodyWithResponse('POST', url, requestBody, responseSchema);
}

export async function jsonPostWithoutBodyWithoutResponse(url: string): Promise<void> {
	await jsonFetchWithoutBodyWithoutResponse('POST', url);
}

export async function jsonDeleteWithoutBodyWithoutResponse(url: string): Promise<void> {
	await jsonFetchWithoutBodyWithoutResponse('DELETE', url);
}

async function jsonFetchWithoutBodyWithResponse<R>(
	httpMethod: 'GET' | 'POST',
	url: string,
	responseSchema: ZodType<R>
): Promise<R> {
	const response = await fetchOrThrow(url, configureRequestOptions(httpMethod));
	await checkIfResponseOkOrThrow(response);
	return await parseResponseOrThrow(response, responseSchema);
}

async function jsonFetchWithBodyWithResponse<B, R>(
	httpMethod: 'POST',
	url: string,
	requestBody: B,
	responseSchema: ZodType<R>
): Promise<R> {
	const response = await fetchOrThrow(url, configureRequestOptions(httpMethod, requestBody));
	await checkIfResponseOkOrThrow(response);
	return await parseResponseOrThrow(response, responseSchema);
}

async function jsonFetchWithoutBodyWithoutResponse(httpMethod: 'POST' | 'DELETE', url: string): Promise<void> {
	const response = await fetchOrThrow(url, configureRequestOptions(httpMethod));
	await checkIfResponseOkOrThrow(response);
}

function configureRequestOptions<B>(method: 'GET' | 'POST' | 'DELETE', requestBody?: B): RequestInit {
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

async function fetchOrThrow(url: string, requestOptions: RequestInit): Promise<Response> {
	try {
		return await fetch(url, requestOptions);
	} catch (error) {
		showErrorToast('Request failed before receiving a response.');
		throw error;
	}
}

async function checkIfResponseOkOrThrow(response: Response): Promise<void> {
	if (response.ok) return;

	const errorMessage = await getErrorResponseMessage(response);
	showErrorToast(errorMessage);
	throw new Error(errorMessage);
}

async function parseResponseOrThrow<R>(response: Response, responseSchema: ZodType<R>): Promise<R> {
	let responseBody: unknown;

	try {
		if (isPrimitiveResponseSchema(responseSchema)) {
			const text = await response.text();
			responseBody = coerceResponseToPrimitiveType(responseSchema, text);
		} else {
			responseBody = await response.json();
		}
	} catch (error) {
		showErrorToast('Failed to parse response body.');
		throw error;
	}

	const parsedResponse = responseSchema.safeParse(responseBody);
	if (parsedResponse.success) return parsedResponse.data;

	showErrorToast('Response validation failed.');
	throw parsedResponse.error;
}

async function getErrorResponseMessage(response: Response): Promise<string> {
	try {
		const parsedErrorResponse = ErrorResponseSchema.safeParse(await response.json());
		if (!parsedErrorResponse.success) {
			return 'Received a malformed error response.';
		}
		return parsedErrorResponse.data.message;
	} catch {
		return 'Failed to parse error response body.';
	}
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
