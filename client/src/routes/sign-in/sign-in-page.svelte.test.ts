import { render, type RenderResult } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { goto } from '$app/navigation';
import { authenticationStore } from '$lib/stores/authentication-store.svelte';
import { expectInputToBeInvalid, expectInputToBeValid, getInputById, getResponse, mockPendingFetch } from '$lib/utils/testing';
import { toast } from 'svelte-sonner';
import { tick } from 'svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import Page from './+page.svelte';

describe('sign-in page tests', () => {
	beforeEach(() => {
		vi.unstubAllGlobals();
		vi.resetAllMocks();
		vi.stubGlobal('fetch', fetchMock);
		authenticationStore.clearAuthentication();
	});

	it('exposes sign-up navigation', () => {
		const page = render(Page);

		expect(page.getByRole('link', { name: 'Sign up' })).toHaveAttribute('href', '/sign-up');
	});

	it('reacts correctly to invalid email', async () => {
		fetchMock.mockResolvedValue(getResponse({ userId: 1 }));
		const page = render(Page);

		await enterCredentials(page, 'not-an-email', 'long-enough-password');
		await clickNext(page);

		expect(fetchMock).not.toHaveBeenCalled();
		expect(gotoMock).not.toHaveBeenCalled();
		await expectInputToBeInvalid(page, 'email-input', 'Invalid email address');
	});

	it('reacts correctly to too short password', async () => {
		fetchMock.mockResolvedValue(getResponse({ userId: 1 }));
		const page = render(Page);

		await enterCredentials(page, 'person@example.com', 'a'.repeat(9));
		await clickNext(page);

		expect(fetchMock).not.toHaveBeenCalled();
		expect(gotoMock).not.toHaveBeenCalled();
		await expectInputToBeInvalid(page, 'password-input', 'Password must have at least 10 characters');
	});

	it('reacts correctly to too long password', async () => {
		fetchMock.mockResolvedValue(getResponse({ userId: 1 }));
		const page = render(Page);

		await enterCredentials(page, 'person@example.com', 'a'.repeat(129));
		await clickNext(page);

		expect(fetchMock).not.toHaveBeenCalled();
		expect(gotoMock).not.toHaveBeenCalled();
		await expectInputToBeInvalid(page, 'password-input', 'Password must have at most 128 characters');
	});

	it('reacts correctly to backend error response', async () => {
		fetchMock.mockResolvedValue(getResponse({ errorCode: 'invalid_credentials', message: 'Invalid email or password.' }, 401));
		const page = render(Page);

		await enterCredentials(page, 'person@example.com', 'long-enough-password');
		await clickNext(page);

		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(gotoMock).not.toHaveBeenCalled();
		expect(toastErrorMock).toHaveBeenCalledWith('Invalid email or password.');
	});

	it('reacts correctly to malformed authentication response', async () => {
		fetchMock.mockResolvedValue(getResponse({ id: 123 }));
		const page = render(Page);

		await enterCredentials(page, 'person@example.com', 'long-enough-password');
		await clickNext(page);

		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(gotoMock).not.toHaveBeenCalled();
		expect(toastErrorMock).toHaveBeenCalledWith('Response validation failed.');
	});

	it('reacts correctly to request failure before receiving a response', async () => {
		fetchMock.mockRejectedValue(new Error('Network failure'));
		const page = render(Page);

		await enterCredentials(page, 'person@example.com', 'long-enough-password');
		await clickNext(page);

		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(gotoMock).not.toHaveBeenCalled();
		expect(toastErrorMock).toHaveBeenCalledWith('Request failed before receiving a response.');
	});

	it('reacts correctly while sign-in request is loading', async () => {
		const resolveFetch = mockPendingFetch(fetchMock);
		const page = render(Page);

		await enterCredentials(page, 'person@example.com', 'long-enough-password');
		await clickNext(page);

		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(page.getByRole('button', { name: 'Signing in...' })).toBeDisabled();

		resolveFetch(getResponse({ errorCode: 'invalid_credentials', message: 'Invalid email or password.' }, 401));
		await tick();
	});

	it('successfully signs in', async () => {
		fetchMock.mockResolvedValue(getResponse({ userId: 123 }));
		const page = render(Page);

		await enterCredentials(page, 'person@example.com', 'long-enough-password');
		await clickNext(page);

		expect(fetchMock).toHaveBeenCalledTimes(1);
		expect(fetchMock).toHaveBeenCalledWith('/api/sign-in', {
			method: 'POST',
			credentials: 'same-origin',
			headers: expect.any(Headers),
			body: JSON.stringify({
				email: 'person@example.com',
				password: 'long-enough-password'
			})
		});
		expect(((fetchMock.mock.calls[0]?.[1] as RequestInit).headers as Headers).get('Content-Type')).toBe('application/json');
		expectInputToBeValid(page, 'email-input');
		expectInputToBeValid(page, 'password-input');
		expect(gotoMock).toHaveBeenCalledTimes(1);
		expect(gotoMock).toHaveBeenCalledWith('/workbench');
	});
});

const fetchMock = vi.fn();
const gotoMock = vi.mocked(goto);
const toastErrorMock = vi.mocked(toast.error);

async function enterCredentials(page: RenderResult<typeof Page>, email: string, password: string): Promise<void> {
	const user = userEvent.setup();
	await user.type(getInputById(page, 'email-input'), email);
	await user.type(getInputById(page, 'password-input'), password);
}

async function clickNext(page: RenderResult<typeof Page>): Promise<void> {
	await userEvent.click(page.getByRole('button', { name: 'Next' }));
	await tick();
}
