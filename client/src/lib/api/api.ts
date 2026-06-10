import { browser } from '$app/environment';
import { toast } from 'svelte-sonner';
import { z, type ZodType } from 'zod';
import { ErrorResponseSchema } from '../types/responses/error-response';

export async function jsonGet<R>(url: string, responseSchema: ZodType<R>): Promise<R | undefined> {
	return await jsonFetchWithoutBodyWithResponse('GET', url, responseSchema);
}

export async function jsonPostWithResponse<B, R>(
	url: string,
	requestBody: B,
	responseSchema: ZodType<R>
): Promise<R | undefined> {
	return await jsonFetchWithBodyWithResponse('POST', url, requestBody, responseSchema);
}

export async function jsonPostWithoutBodyWithoutResponse(url: string): Promise<void> {
	await jsonFetchWithoutBodyWithoutResponse('POST', url);
}

async function jsonFetchWithoutBodyWithResponse<R>(
	httpMethod: 'GET',
	url: string,
	responseSchema: ZodType<R>
): Promise<R | undefined> {
	const response = await fetchOrLog(url, configureRequestOptions(httpMethod));
	if (response === undefined) return undefined;
	if (!(await isResponseOkOrLog(response))) return undefined;
	return await parseResponseOrLog(response, responseSchema);
}

async function jsonFetchWithBodyWithResponse<B, R>(
	httpMethod: 'POST',
	url: string,
	requestBody: B,
	responseSchema: ZodType<R>
): Promise<R | undefined> {
	const response = await fetchOrLog(url, configureRequestOptions(httpMethod, requestBody));
	if (response === undefined) return undefined;
	if (!(await isResponseOkOrLog(response))) return undefined;
	return await parseResponseOrLog(response, responseSchema);
}

async function jsonFetchWithoutBodyWithoutResponse(httpMethod: 'POST', url: string): Promise<void> {
	const response = await fetchOrLog(url, configureRequestOptions(httpMethod));
	if (response === undefined) return;
	await isResponseOkOrLog(response);
}

function configureRequestOptions<B>(method: 'GET' | 'POST', requestBody?: B): RequestInit {
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

async function fetchOrLog(url: string, requestOptions: RequestInit): Promise<Response | undefined> {
	try {
		return await fetch(url, requestOptions);
	} catch {
		showErrorToast('Request failed before receiving a response.');
		return undefined;
	}
}

async function isResponseOkOrLog(response: Response): Promise<boolean> {
	if (response.ok) return true;

	await logErrorResponse(response);
	return false;
}

async function parseResponseOrLog<R>(response: Response, responseSchema: ZodType<R>): Promise<R | undefined> {
	try {
		if (isPrimitiveResponseSchema(responseSchema)) {
			const text = await response.text();
			return parseWithSchemaOrLog(responseSchema, coerceResponseToPrimitiveType(responseSchema, text));
		}
		const json = await response.json();
		return parseWithSchemaOrLog(responseSchema, json);
	} catch {
		showErrorToast('Failed to parse response body.');
		return undefined;
	}
}

function parseWithSchemaOrLog<R>(responseSchema: ZodType<R>, value: unknown): R | undefined {
	const parsedResponse = responseSchema.safeParse(value);
	if (parsedResponse.success) return parsedResponse.data;

	showErrorToast('Response validation failed.');
	return undefined;
}

async function logErrorResponse(response: Response) {
	try {
		const parsedErrorResponse = ErrorResponseSchema.safeParse(await response.json());
		if (!parsedErrorResponse.success) {
			showErrorToast('Received a malformed error response.');
			return;
		}
		showErrorToast(parsedErrorResponse.data.message);
	} catch {
		showErrorToast('Failed to parse error response body.');
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
