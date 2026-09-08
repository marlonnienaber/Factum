import { waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { expect, vi } from 'vitest';

type InputQueries = {
	container: HTMLElement;
	findByText: (text: string) => Promise<HTMLElement>;
};

export function getInputById(queries: Pick<InputQueries, 'container'>, inputId: string): HTMLInputElement {
	const input = queries.container.querySelector(`#${inputId}`);
	if (!(input instanceof HTMLInputElement)) {
		throw new Error(`Expected an input with id "${inputId}".`);
	}
	return input;
}

export function expectInputToBeValid(queries: Pick<InputQueries, 'container'>, inputId: string): void {
	const input = getInputById(queries, inputId);
	expect(input).toHaveAttribute('aria-invalid', 'false');
	expect(queries.container.querySelector(`#${inputId}-error-indicator-popover-trigger`)).not.toBeInTheDocument();
}

export async function expectInputToBeInvalid(queries: InputQueries, inputId: string, expectedError: string): Promise<void> {
	const input = getInputById(queries, inputId);
	expect(input).toHaveAttribute('aria-invalid', 'true');

	const errorIndicator = queries.container.querySelector(`#${inputId}-error-indicator-popover-trigger`)?.closest('button');
	if (!(errorIndicator instanceof HTMLButtonElement)) {
		throw new Error(`Expected invalid input "${inputId}" to expose an error indicator.`);
	}

	await userEvent.click(errorIndicator);
	expect(await queries.findByText(expectedError)).toBeInTheDocument();

	await userEvent.click(errorIndicator);
	await waitFor(() => expect(queries.container.ownerDocument.body).not.toHaveTextContent(expectedError));
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
