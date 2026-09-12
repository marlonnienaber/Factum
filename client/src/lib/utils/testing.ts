import { waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { mockIntersectionObserver } from 'jsdom-testing-mocks';
import { expect, vi } from 'vitest';

export const intersectionObserverMock = mockIntersectionObserver();

type PageQueries = {
	container: HTMLElement;
	findByText: (text: string) => Promise<HTMLElement>;
};

export function getInputById(page: Pick<PageQueries, 'container'>, inputId: string): HTMLInputElement {
	const input = page.container.ownerDocument.getElementById(inputId);
	if (!(input instanceof HTMLInputElement)) {
		throw new Error(`Expected an input with id "${inputId}".`);
	}
	return input;
}

export function getTextareaById(page: Pick<PageQueries, 'container'>, textareaId: string): HTMLTextAreaElement {
	const textarea = page.container.ownerDocument.getElementById(textareaId);
	if (!(textarea instanceof HTMLTextAreaElement)) {
		throw new Error(`Expected a textarea with id "${textareaId}".`);
	}
	return textarea;
}

export function getSelectById(page: Pick<PageQueries, 'container'>, selectId: string): HTMLButtonElement {
	const select = page.container.ownerDocument.getElementById(selectId);
	if (!(select instanceof HTMLButtonElement)) {
		throw new Error(`Expected a select trigger with id "${selectId}".`);
	}
	return select;
}

export function expectInputToBeValid(page: Pick<PageQueries, 'container'>, inputId: string): void {
	const input = getInputById(page, inputId);
	expect(input).toHaveAttribute('aria-invalid', 'false');
	expect(page.container.ownerDocument.getElementById(`${inputId}-error-indicator-popover-trigger`)).not.toBeInTheDocument();
}

export async function expectInputToBeInvalid(page: PageQueries, inputId: string, expectedError: string): Promise<void> {
	const input = getInputById(page, inputId);
	await expectControlToBeInvalid(page, input, inputId, expectedError);
}

export async function expectTextareaToBeInvalid(page: PageQueries, textareaId: string, expectedError: string): Promise<void> {
	const textarea = getTextareaById(page, textareaId);
	await expectControlToBeInvalid(page, textarea, textareaId, expectedError);
}

export async function expectSelectToBeInvalid(page: PageQueries, selectId: string, expectedError: string): Promise<void> {
	const select = getSelectById(page, selectId);
	await expectControlToBeInvalid(page, select, selectId, expectedError);
}

async function expectControlToBeInvalid(
	page: PageQueries,
	control: HTMLElement,
	controlId: string,
	expectedError: string
): Promise<void> {
	expect(control).toHaveAttribute('aria-invalid', 'true');

	const errorIndicator = page.container.ownerDocument
		.getElementById(`${controlId}-error-indicator-popover-trigger`)
		?.closest('button');
	if (!(errorIndicator instanceof HTMLButtonElement)) {
		throw new Error(`Expected invalid control "${controlId}" to expose an error indicator.`);
	}

	await userEvent.click(errorIndicator);
	expect(await page.findByText(expectedError)).toBeInTheDocument();

	await userEvent.click(errorIndicator);
	await waitFor(() => expect(page.container.ownerDocument.body).not.toHaveTextContent(expectedError));
}

export function mockPendingFetch(fetchMock: ReturnType<typeof vi.fn>): (response: Response) => void {
	let resolvePendingFetch: (response: Response) => void = () => {
		throw new Error('Expected fetch promise to be created before resolving it.');
	};
	fetchMock.mockImplementation(
		() =>
			new Promise<Response>((resolve) => {
				resolvePendingFetch = resolve;
			})
	);
	return (response) => resolvePendingFetch(response);
}

export function getResponse(body: unknown, status = 200): Response {
	return new Response(JSON.stringify(body), {
		status,
		headers: { 'Content-Type': 'application/json' }
	});
}

export function expectRequestUrl(fetchMock: ReturnType<typeof vi.fn>, expectedUrl: string): void {
	const requestUrl = fetchMock.mock.calls[0]?.[0];
	expect(requestUrl).toBe(expectedUrl);
}

export function expectRequestPath(fetchMock: ReturnType<typeof vi.fn>, expectedPath: string): void {
	const parsedRequestUrl = getParsedRequestUrl(fetchMock);
	expect(parsedRequestUrl.pathname).toBe(expectedPath);
}

export function expectRequestMethod(fetchMock: ReturnType<typeof vi.fn>, expectedMethod: string): void {
	const requestOptions = fetchMock.mock.calls[0]?.[1] as RequestInit;
	expect(requestOptions.method).toBe(expectedMethod);
}

export function expectRequestCredentialsMode(fetchMock: ReturnType<typeof vi.fn>, expectedMode: RequestCredentials): void {
	const requestOptions = fetchMock.mock.calls[0]?.[1] as RequestInit;
	expect(requestOptions.credentials).toBe(expectedMode);
}

export function expectRequestBody(fetchMock: ReturnType<typeof vi.fn>, expectedBody: string | undefined): void {
	const requestOptions = fetchMock.mock.calls[0]?.[1] as RequestInit;
	expect(requestOptions.body).toBe(expectedBody);
}

export function expectRequestContentType(fetchMock: ReturnType<typeof vi.fn>, expectedContentType: string): void {
	const requestOptions = fetchMock.mock.calls[0]?.[1] as RequestInit;
	const requestHeaders = requestOptions.headers as Headers;
	expect(requestHeaders).toBeInstanceOf(Headers);
	const contentType = requestHeaders.get('Content-Type');
	expect(contentType).toBe(expectedContentType);
}

export function getRequestSearchParameter(fetchMock: ReturnType<typeof vi.fn>, parameterName: string): string | null {
	const parsedRequestUrl = getParsedRequestUrl(fetchMock);
	return parsedRequestUrl.searchParams.get(parameterName);
}

function getRequestUrl(fetchMock: ReturnType<typeof vi.fn>): string {
	const requestUrl = fetchMock.mock.calls[0]?.[0];
	if (typeof requestUrl !== 'string') throw new Error('Expected the request URL to be a string.');
	return requestUrl;
}

function getParsedRequestUrl(fetchMock: ReturnType<typeof vi.fn>): URL {
	const requestUrl = getRequestUrl(fetchMock);
	return new URL(requestUrl, document.baseURI);
}
