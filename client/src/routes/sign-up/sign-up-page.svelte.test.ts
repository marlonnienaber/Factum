import '@testing-library/jest-dom/vitest';
import { render } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { goto } from '$app/navigation';
import { toast } from 'svelte-sonner';
import { tick } from 'svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Page from './+page.svelte';

vi.mock('$app/navigation', () => ({
	goto: vi.fn()
}));

vi.mock('$app/paths', () => ({
	resolve: (path: string) => path
}));

vi.mock('svelte-sonner', () => ({
	toast: {
		error: vi.fn()
	}
}));

const gotoMock = vi.mocked(goto);
const toastErrorMock = vi.mocked(toast.error);

function renderPage() {
	return render(Page);
}

function mockSuccessfulFetch(response: Response): ReturnType<typeof vi.fn> {
	const fetchMock = vi.fn().mockResolvedValue(response);
	vi.stubGlobal('fetch', fetchMock);
	return fetchMock;
}

function mockErrorFetch(): ReturnType<typeof vi.fn> {
	const fetchMock = vi.fn().mockRejectedValue(new Error('Network failure'));
	vi.stubGlobal('fetch', fetchMock);
	return fetchMock;
}

function mockPendingFetch(): {
	fetchMock: ReturnType<typeof vi.fn>;
	resolveFetch: (response: Response) => void;
} {
	let resolvePendingFetch: (response: Response) => void = () => {
		throw new Error('Expected fetch promise to be created before resolving it.');
	};
	const fetchMock = vi.fn(
		() =>
			new Promise<Response>((resolve) => {
				resolvePendingFetch = resolve;
			})
	);
	vi.stubGlobal('fetch', fetchMock);
	return {
		fetchMock,
		resolveFetch: (response) => resolvePendingFetch(response)
	};
}

function getResponseJson(body: unknown, status = 200): Response {
	return new Response(JSON.stringify(body), {
		status,
		headers: { 'Content-Type': 'application/json' }
	});
}

function getAuthenticationInput(input: HTMLElement): HTMLElement {
	const authenticationInput = input.closest('.authentication-input');
	if (!(authenticationInput instanceof HTMLElement)) {
		throw new Error('Expected input to be wrapped by an authentication input.');
	}
	return authenticationInput;
}

async function enterEmail(page: ReturnType<typeof renderPage>, email: string): Promise<void> {
	const user = userEvent.setup();
	await user.type(page.getByTestId('email-input'), email);
}

async function enterPassword(page: ReturnType<typeof renderPage>, password: string): Promise<void> {
	const user = userEvent.setup();
	await user.type(page.getByTestId('password-input'), password);
}

async function enterRepeatedPassword(page: ReturnType<typeof renderPage>, repeatedPassword: string): Promise<void> {
	const user = userEvent.setup();
	await user.type(page.getByTestId('repeated-password-input'), repeatedPassword);
}

async function submit(page: ReturnType<typeof renderPage>): Promise<void> {
	const user = userEvent.setup();
	await user.click(page.getByTestId('next-button'));
	await tick();
}

describe('sign-up page integration', () => {
	beforeEach(() => {
		vi.unstubAllGlobals();
		gotoMock.mockReset();
		toastErrorMock.mockReset();
	});

	it('exposes sign-in navigation', async () => {
		const fetchMock = mockSuccessfulFetch(getResponseJson({ userId: 1 }));
		const { getByRole } = renderPage();

		expect(getByRole('link', { name: 'Sign in' })).toHaveAttribute('href', '/sign-in');
		expect(fetchMock).not.toHaveBeenCalled();
	});

	it('reacts correctly to invalid email', async () => {
		const fetchMock = mockSuccessfulFetch(getResponseJson({ userId: 1 }));
		const page = renderPage();

		await enterEmail(page, 'not-an-email');
		await enterPassword(page, 'long-enough-password');
		await enterRepeatedPassword(page, 'long-enough-password');

		await submit(page);

		expect(fetchMock).not.toHaveBeenCalled();
		expect(gotoMock).not.toHaveBeenCalled();

		const emailInput = getAuthenticationInput(page.getByTestId('email-input'));
		const emailErrorIcon = emailInput.querySelector('.error-icon');
		expect(emailInput).toHaveClass('invalid');
		expect(emailErrorIcon).toBeInTheDocument();

		await userEvent.click(page.getByTestId('email-input-error-trigger'));
		expect(await page.findByText('Invalid email address')).toBeInTheDocument();
	});

	it('reacts correctly to too short password', async () => {
		const fetchMock = mockSuccessfulFetch(getResponseJson({ userId: 1 }));
		const page = renderPage();

		await enterEmail(page, 'person@example.com');
		await enterPassword(page, 'a'.repeat(9));
		await enterRepeatedPassword(page, 'a'.repeat(9));

		await submit(page);

		expect(fetchMock).not.toHaveBeenCalled();
		expect(gotoMock).not.toHaveBeenCalled();

		const passwordInput = getAuthenticationInput(page.getByTestId('password-input'));
		const passwordErrorIcon = passwordInput.querySelector('.error-icon');
		expect(passwordInput).toHaveClass('invalid');
		expect(passwordErrorIcon).toBeInTheDocument();

		await userEvent.click(page.getByTestId('password-input-error-trigger'));
		expect(await page.findByText('Too small: expected string to have >=10 characters')).toBeInTheDocument();
	});

	it('reacts correctly to too long password', async () => {
		const fetchMock = mockSuccessfulFetch(getResponseJson({ userId: 1 }));
		const page = renderPage();

		await enterEmail(page, 'person@example.com');
		await enterPassword(page, 'a'.repeat(129));
		await enterRepeatedPassword(page, 'a'.repeat(129));

		await submit(page);

		expect(fetchMock).not.toHaveBeenCalled();
		expect(gotoMock).not.toHaveBeenCalled();

		const passwordInput = getAuthenticationInput(page.getByTestId('password-input'));
		const passwordErrorIcon = passwordInput.querySelector('.error-icon');
		expect(passwordInput).toHaveClass('invalid');
		expect(passwordErrorIcon).toBeInTheDocument();

		await userEvent.click(page.getByTestId('password-input-error-trigger'));
		expect(await page.findByText('Too big: expected string to have <=128 characters')).toBeInTheDocument();
	});

	it('reacts correctly to password mismatch', async () => {
		const fetchMock = mockSuccessfulFetch(getResponseJson({ userId: 1 }));
		const page = renderPage();

		await enterEmail(page, 'person@example.com');
		await enterPassword(page, 'long-enough-password');
		await enterRepeatedPassword(page, 'different-password');

		await submit(page);

		expect(fetchMock).not.toHaveBeenCalled();
		expect(gotoMock).not.toHaveBeenCalled();

		const repeatedPasswordInput = getAuthenticationInput(page.getByTestId('repeated-password-input'));
		const repeatedPasswordErrorIcon = repeatedPasswordInput.querySelector('.error-icon');
		expect(repeatedPasswordInput).toHaveClass('invalid');
		expect(repeatedPasswordErrorIcon).toBeInTheDocument();

		await userEvent.click(page.getByTestId('repeated-password-input-error-trigger'));
		expect(await page.findByText('Passwords do not match.')).toBeInTheDocument();
	});

	it('reacts correctly to backend error response', async () => {
		const fetchMock = mockSuccessfulFetch(
			getResponseJson(
				{
					errorCode: 'email_already_exists',
					message: 'Email already exists.'
				},
				409
			)
		);
		const page = renderPage();

		await enterEmail(page, 'person@example.com');
		await enterPassword(page, 'long-enough-password');
		await enterRepeatedPassword(page, 'long-enough-password');

		await submit(page);

		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(gotoMock).not.toHaveBeenCalled();
		expect(toastErrorMock).toHaveBeenCalledTimes(1);
		expect(toastErrorMock).toHaveBeenCalledWith('Email already exists.');
	});

	it('reacts correctly to malformed authentication response', async () => {
		const fetchMock = mockSuccessfulFetch(getResponseJson({ id: 123 }));
		const page = renderPage();

		await enterEmail(page, 'person@example.com');
		await enterPassword(page, 'long-enough-password');
		await enterRepeatedPassword(page, 'long-enough-password');

		await submit(page);

		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(gotoMock).not.toHaveBeenCalled();
		expect(toastErrorMock).toHaveBeenCalledTimes(1);
		expect(toastErrorMock).toHaveBeenCalledWith('Response validation failed.');
	});

	it('reacts correctly to request failure before receiving a response', async () => {
		const fetchMock = mockErrorFetch();
		const page = renderPage();

		await enterEmail(page, 'person@example.com');
		await enterPassword(page, 'long-enough-password');
		await enterRepeatedPassword(page, 'long-enough-password');

		await submit(page);

		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(gotoMock).not.toHaveBeenCalled();
		expect(toastErrorMock).toHaveBeenCalledTimes(1);
		expect(toastErrorMock).toHaveBeenCalledWith('Request failed before receiving a response.');
	});

	it('reacts correctly while sign-up request is loading', async () => {
		const { fetchMock, resolveFetch } = mockPendingFetch();
		const page = renderPage();

		await enterEmail(page, 'person@example.com');
		await enterPassword(page, 'long-enough-password');
		await enterRepeatedPassword(page, 'long-enough-password');

		await userEvent.click(page.getByTestId('next-button'));
		await tick();

		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(page.getByTestId('next-button')).toBeDisabled();
		expect(page.getByTestId('next-button')).toHaveTextContent('Signing up...');

		resolveFetch(getResponseJson({ errorCode: 'example_error_code', message: 'Example error message.' }, 409));
		await tick();
	});

	it('successful sign-up', async () => {
		const fetchMock = mockSuccessfulFetch(getResponseJson({ userId: 123 }));
		const page = renderPage();

		await enterEmail(page, 'person@example.com');
		await enterPassword(page, 'long-enough-password');
		await enterRepeatedPassword(page, 'long-enough-password');

		await submit(page);

		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(fetchMock).toHaveBeenCalledWith('/api/sign-up', {
			method: 'POST',
			credentials: 'same-origin',
			headers: expect.any(Headers),
			body: JSON.stringify({
				email: 'person@example.com',
				password: 'long-enough-password'
			})
		});
		expect((fetchMock.mock.calls[0]?.[1] as RequestInit).headers).toHaveProperty('get');
		expect(((fetchMock.mock.calls[0]?.[1] as RequestInit).headers as Headers).get('Content-Type')).toBe('application/json');

		const emailInput = getAuthenticationInput(page.getByTestId('email-input'));
		const passwordInput = getAuthenticationInput(page.getByTestId('password-input'));
		const repeatedPasswordInput = getAuthenticationInput(page.getByTestId('repeated-password-input'));
		expect(emailInput).not.toHaveClass('invalid');
		expect(emailInput.querySelector('.error-icon')).not.toBeInTheDocument();
		expect(passwordInput).not.toHaveClass('invalid');
		expect(passwordInput.querySelector('.error-icon')).not.toBeInTheDocument();
		expect(repeatedPasswordInput).not.toHaveClass('invalid');
		expect(repeatedPasswordInput.querySelector('.error-icon')).not.toBeInTheDocument();
		expect(gotoMock).toHaveBeenCalledTimes(1);
		expect(gotoMock).toHaveBeenCalledWith('/backlog');
	});
});
