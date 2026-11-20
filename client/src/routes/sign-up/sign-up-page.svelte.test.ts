import { render, waitFor, type RenderResult } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { goto } from '$app/navigation';
import { authenticationStore } from '$lib/stores/authentication-store.svelte';
import {
	expectInputToBeInvalid,
	expectInputToBeValid,
	expectRequestContentType,
	getInputById,
	getResponse,
	mockPendingFetch
} from '$lib/utils/testing';
import { toast } from 'svelte-sonner';
import { tick } from 'svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import Page from './+page.svelte';

describe('sign-up page tests', () => {
	beforeEach(() => {
		vi.unstubAllGlobals();
		vi.resetAllMocks();
		vi.stubGlobal('fetch', fetchMock);
		authenticationStore.clearAuthentication();
	});

	describe('email step', () => {
		it('exposes sign-in navigation', () => {
			const page = render(Page);

			expect(page.getByRole('link', { name: 'Sign in' })).toHaveAttribute('href', '/sign-in');
		});

		it('reacts correctly to invalid email', async () => {
			const user = userEvent.setup();
			const page = render(Page);

			await user.type(getInputById(page, 'email-input'), 'not-an-email');
			await clickNext(page, user);

			expect(fetchMock).not.toHaveBeenCalled();
			expect(gotoMock).not.toHaveBeenCalled();
			await expectInputToBeInvalid(page, 'email-input', 'Provide a valid email address.');
		});

		it('stops displaying validation errors on editing', async () => {
			const user = userEvent.setup();
			const page = render(Page);
			const input = getInputById(page, 'email-input');
			await user.type(input, 'not-an-email');
			await clickNext(page, user);
			await expectInputToBeInvalid(page, 'email-input', 'Provide a valid email address.');

			await user.type(input, 'x');

			expectInputToBeValid(page, 'email-input');
		});

		it('reacts correctly to backend error response', async () => {
			const user = userEvent.setup();
			fetchMock.mockResolvedValue(getResponse({ errorCode: 'unexpected_error', message: 'An unexpected error occurred.' }, 500));
			const page = render(Page);

			await user.type(getInputById(page, 'email-input'), 'person@example.com');
			await clickNext(page, user);

			await waitFor(() => expect(toastErrorMock).toHaveBeenCalledWith('An unexpected error occurred.'));
			expect(fetchMock).toHaveBeenCalledTimes(1);
			expect(gotoMock).not.toHaveBeenCalled();
			expect(getInputById(page, 'email-input')).toHaveValue('person@example.com');
			expect(page.getByRole('button', { name: 'Next' })).toBeEnabled();
		});

		it('reacts correctly to request failure before receiving a response', async () => {
			const user = userEvent.setup();
			fetchMock.mockRejectedValue(new Error('Network failure'));
			const page = render(Page);

			await user.type(getInputById(page, 'email-input'), 'person@example.com');
			await clickNext(page, user);

			expect(fetchMock).toHaveBeenCalledTimes(1);
			expect(gotoMock).not.toHaveBeenCalled();
			expect(toastErrorMock).toHaveBeenCalledWith('Request failed before receiving a response.');
			expect(getInputById(page, 'email-input')).toHaveValue('person@example.com');
			expect(page.getByRole('button', { name: 'Next' })).toBeEnabled();
		});

		it('reacts correctly while verify-email request is loading', async () => {
			const user = userEvent.setup();
			const resolveFetch = mockPendingFetch(fetchMock);
			const page = render(Page);
			const input = getInputById(page, 'email-input');

			await user.type(input, 'person@example.com');
			await clickNext(page, user);
			const button = page.getByRole('button', { name: 'Requesting link...' });
			await user.click(button);

			expect(fetchMock).toHaveBeenCalledTimes(1);
			expect(input).toBeDisabled();
			expect(button).toBeDisabled();

			resolveFetch(new Response(undefined, { status: 204 }));
			await expectInCheckEmailStep(page);
		});

		it('successfully requests email verification', async () => {
			const user = userEvent.setup();
			fetchMock.mockResolvedValue(new Response(undefined, { status: 204 }));
			const page = render(Page);

			await user.type(getInputById(page, 'email-input'), 'Person@Example.com');
			await clickNext(page, user);

			expect(fetchMock).toHaveBeenCalledTimes(1);
			expect(fetchMock).toHaveBeenCalledWith('/api/verify-email', {
				method: 'POST',
				credentials: 'same-origin',
				headers: expect.any(Headers),
				body: JSON.stringify({
					email: 'Person@Example.com'
				})
			});
			expectRequestContentType(fetchMock, 'application/json');
			await expectInCheckEmailStep(page);
			expect(page.getByText('Person@Example.com', { exact: true })).toBeInTheDocument();
		});
	});

	describe('check email step', () => {
		beforeEach(() => {
			vi.useFakeTimers({ toFake: ['Date', 'setInterval', 'clearInterval'] });
			vi.setSystemTime(new Date('2026-01-01T00:00:00Z'));
		});

		afterEach(() => {
			vi.useRealTimers();
		});

		it('exposes sign-in navigation', async () => {
			const { page } = await openCheckEmailStep();

			expect(page.getByRole('link', { name: 'Sign in' })).toHaveAttribute('href', '/sign-in');
		});

		it('starts with a disabled resend button and a one-minute countdown', async () => {
			const { page } = await openCheckEmailStep();

			expect(page.getByRole('button', { name: 'Resend in 1m 0s' })).toBeDisabled();
		});

		it('updates the resend countdown as time passes', async () => {
			const { page } = await openCheckEmailStep();

			await advanceClock(10 * second);
			expect(page.getByRole('button', { name: 'Resend in 50s' })).toBeDisabled();

			await advanceClock(15 * second);
			expect(page.getByRole('button', { name: 'Resend in 35s' })).toBeDisabled();
		});

		it('keeps resend disabled just before the minutely window ends', async () => {
			const { page, user } = await openCheckEmailStep();

			await advanceClock(minute - second);
			const button = page.getByRole('button', { name: 'Resend in 1s' });
			await user.click(button);

			expect(button).toBeDisabled();
			expect(fetchMock).toHaveBeenCalledTimes(1);
		});

		it('enables resend exactly when the minutely window ends', async () => {
			const { page } = await openCheckEmailStep();

			await advanceClock(minute);

			expect(page.getByRole('button', { name: 'Resend link' })).toBeEnabled();
		});

		it('keeps resend enabled after the minutely window ends', async () => {
			const { page } = await openCheckEmailStep();

			await advanceClock(2 * minute);

			expect(page.getByRole('button', { name: 'Resend link' })).toBeEnabled();
		});

		it('counts the initial request toward the hourly limit and keeps the window anchored to that request', async () => {
			const { page, user } = await openCheckEmailStep();
			await makeResendRequests(page, user, [minute, minute, minute]);

			await advanceClock(minute);
			const resendButton = page.getByRole('button', { name: 'Resend link' });
			await user.click(resendButton);

			expect(fetchMock).toHaveBeenCalledTimes(5);
			expect(page.getByRole('button', { name: 'Resend in 56m 0s' })).toBeDisabled();
		});

		it('keeps resend disabled just before the hourly window ends', async () => {
			const { page, user } = await openCheckEmailStep();
			await makeResendRequests(page, user, [minute, minute, minute, minute]);

			await advanceClock(56 * minute - second);
			const button = page.getByRole('button', { name: 'Resend in 1s' });
			await user.click(button);

			expect(button).toBeDisabled();
			expect(fetchMock).toHaveBeenCalledTimes(5);
		});

		it('enables resend exactly when the hourly window ends if no other limit is active', async () => {
			const { page, user } = await openCheckEmailStep();
			await makeResendRequests(page, user, [minute, minute, minute, minute]);

			await advanceClock(56 * minute);

			expect(page.getByRole('button', { name: 'Resend link' })).toBeEnabled();
		});

		it('starts a new hourly window with a count of one on the first request after the previous window ends', async () => {
			const { page, user } = await openCheckEmailStep();
			await makeResendRequests(page, user, [minute, minute, minute, hour, minute, minute, minute]);

			await advanceClock(minute);
			const resendButton = page.getByRole('button', { name: 'Resend link' });
			await user.click(resendButton);

			expect(fetchMock).toHaveBeenCalledTimes(9);
			expect(page.getByRole('button', { name: 'Resend in 56m 0s' })).toBeDisabled();
		});

		it('counts the initial request toward the daily limit and preserves the count across hourly resets', async () => {
			const { page, user } = await openCheckEmailStep();
			await makeResendRequests(page, user, [minute, minute, minute, minute, hour - 4 * minute, minute, minute, minute]);

			await advanceClock(minute);
			const resendButton = page.getByRole('button', { name: 'Resend link' });
			await user.click(resendButton);

			expect(fetchMock).toHaveBeenCalledTimes(10);
			expect(page.getByRole('button', { name: 'Resend in 22h 56m' })).toBeDisabled();
		});

		it('keeps resend disabled just before the daily window ends', async () => {
			const { page, user } = await openCheckEmailStep();
			await reachDailyLimit(page, user);

			await advanceClock(day - hour - 4 * minute - second);
			const button = page.getByRole('button', { name: 'Resend in 1s' });
			await user.click(button);

			expect(button).toBeDisabled();
			expect(fetchMock).toHaveBeenCalledTimes(10);
		});

		it('enables resend exactly when the daily window ends if no other limit is active', async () => {
			const { page, user } = await openCheckEmailStep();
			await reachDailyLimit(page, user);

			await advanceClock(day - hour - 4 * minute);

			expect(page.getByRole('button', { name: 'Resend link' })).toBeEnabled();
		});

		it('starts a new daily window with a count of one on the first request after the previous window ends', async () => {
			const { page, user } = await openCheckEmailStep();
			await reachDailyLimit(page, user);

			await advanceClock(day - 4 * minute);
			const resendButton = page.getByRole('button', { name: 'Resend link' });
			await user.click(resendButton);

			expect(fetchMock).toHaveBeenCalledTimes(11);
			expect(page.getByRole('button', { name: 'Resend in 1m 0s' })).toBeDisabled();
			await reachDailyLimit(page, user);
			expect(fetchMock).toHaveBeenCalledTimes(20);
			expect(page.getByRole('button', { name: 'Resend in 22h 56m' })).toBeDisabled();
		});

		it('keeps resend disabled when the minutely window ends while the hourly limit is active', async () => {
			const { page, user } = await openCheckEmailStep();
			await makeResendRequests(page, user, [minute, minute, minute, minute]);

			await advanceClock(minute);

			expect(page.getByRole('button', { name: 'Resend in 55m 0s' })).toBeDisabled();
		});

		it('keeps resend disabled when the minutely window ends while the daily limit is active', async () => {
			const { page, user } = await openCheckEmailStep();
			await reachDailyLimit(page, user);

			await advanceClock(minute);

			expect(page.getByRole('button', { name: 'Resend in 22h 55m' })).toBeDisabled();
		});

		it('keeps resend disabled when the hourly window ends while the minutely window is active', async () => {
			const { page, user } = await openCheckEmailStep();
			await makeResendRequests(page, user, [minute, minute, minute, 56 * minute + 30 * second]);

			await advanceClock(30 * second);

			expect(page.getByRole('button', { name: 'Resend in 30s' })).toBeDisabled();
		});

		it('keeps resend disabled when the hourly window ends while the daily limit is active', async () => {
			const { page, user } = await openCheckEmailStep();
			await reachDailyLimit(page, user);

			await advanceClock(56 * minute);

			expect(page.getByRole('button', { name: 'Resend in 22h 0m' })).toBeDisabled();
		});

		it('keeps resend disabled when the daily window ends while the minutely window is active', async () => {
			const { page, user } = await openCheckEmailStep();
			await makeResendRequests(page, user, [
				minute,
				minute,
				minute,
				minute,
				hour - 4 * minute,
				minute,
				minute,
				minute,
				22 * hour + 56 * minute + 30 * second
			]);

			await advanceClock(30 * second);

			expect(page.getByRole('button', { name: 'Resend in 30s' })).toBeDisabled();
		});

		it('keeps resend disabled when the daily window ends while the hourly limit is active', async () => {
			const { page, user } = await openCheckEmailStep();
			await makeResendRequests(page, user, [
				minute,
				minute,
				minute,
				minute,
				23 * hour + 26 * minute,
				minute,
				minute,
				minute,
				minute
			]);

			await advanceClock(26 * minute);

			expect(page.getByRole('button', { name: 'Resend in 30m 0s' })).toBeDisabled();
		});

		it('rounds a positive fractional second up in the resend countdown', async () => {
			const { page, user } = await openCheckEmailStep();

			await advanceClock(minute + 250);
			const resendButton = page.getByRole('button', { name: 'Resend link' });
			await user.click(resendButton);
			await advanceClock(minute - 250);

			expect(page.getByRole('button', { name: 'Resend in 1s' })).toBeDisabled();
		});

		it('updates the countdown from elapsed time after timer updates have been delayed', async () => {
			const { page } = await openCheckEmailStep();

			await advanceClock(10 * second);
			vi.setSystemTime(Date.now() + 29 * second);
			await advanceClock(second);

			expect(page.getByRole('button', { name: 'Resend in 20s' })).toBeDisabled();
		});

		it('starts the resend cooldown when the response is received', async () => {
			const { page, user } = await openCheckEmailStep();

			await advanceClock(minute);
			const resolveFetch = mockPendingFetch(fetchMock);
			const resendButton = page.getByRole('button', { name: 'Resend link' });
			await user.click(resendButton);
			await advanceClock(10 * second);
			resolveFetch(new Response(undefined, { status: 204 }));

			expect(await page.findByRole('button', { name: 'Resend in 1m 0s' })).toBeDisabled();
		});

		it('prevents additional resend requests while a request is loading even after a minute has passed', async () => {
			const { page, user } = await openCheckEmailStep();
			await advanceClock(minute);
			const resolveFetch = mockPendingFetch(fetchMock);

			const resendButton = page.getByRole('button', { name: 'Resend link' });
			await user.click(resendButton);

			await advanceClock(minute);
			const button = page.getByRole('button', { name: 'Requesting link...' });
			await user.click(button);
			expect(button).toBeDisabled();
			expect(page.getByRole('button', { name: 'Back to email entry' })).toBeDisabled();
			expect(fetchMock).toHaveBeenCalledTimes(2);
			resolveFetch(new Response(undefined, { status: 204 }));
			await page.findByRole('button', { name: 'Resend in 1m 0s' });
		});

		it('counts a resend attempt toward all limits when the backend returns an error', async () => {
			const { page, user } = await openCheckEmailStep();

			await makeResendRequests(page, user, [minute, minute]);
			await advanceClock(minute);
			fetchMock.mockResolvedValueOnce(
				getResponse({ errorCode: 'unexpected_error', message: 'An unexpected error occurred.' }, 500)
			);

			const resendButton = page.getByRole('button', { name: 'Resend link' });
			await user.click(resendButton);

			expect(await page.findByRole('button', { name: 'Resend in 1m 0s' })).toBeDisabled();
			expect(toastErrorMock).toHaveBeenCalledWith('An unexpected error occurred.');

			await makeResendRequests(page, user, [minute]);
			expect(page.getByRole('button', { name: 'Resend in 56m 0s' })).toBeDisabled();

			await makeResendRequests(page, user, [56 * minute, minute, minute, minute, minute]);
			expect(page.getByRole('button', { name: 'Resend in 22h 56m' })).toBeDisabled();
		});

		it('counts a resend attempt toward all limits when the request fails before receiving a response', async () => {
			const { page, user } = await openCheckEmailStep();

			await makeResendRequests(page, user, [minute, minute]);
			await advanceClock(minute);
			fetchMock.mockRejectedValueOnce(new Error('Network failure'));

			const resendButton = page.getByRole('button', { name: 'Resend link' });
			await user.click(resendButton);

			expect(await page.findByRole('button', { name: 'Resend in 1m 0s' })).toBeDisabled();
			expect(toastErrorMock).toHaveBeenCalledWith('Request failed before receiving a response.');

			await makeResendRequests(page, user, [minute]);
			expect(page.getByRole('button', { name: 'Resend in 56m 0s' })).toBeDisabled();

			await makeResendRequests(page, user, [56 * minute, minute, minute, minute, minute]);
			expect(page.getByRole('button', { name: 'Resend in 22h 56m' })).toBeDisabled();
		});

		it('clears all local throttling windows when returning to email entry and submitting the same email', async () => {
			const { page, user } = await openCheckEmailStep();
			await reachDailyLimit(page, user);

			const backButton = page.getByRole('button', { name: 'Back to email entry' });
			await user.click(backButton);
			const input = getInputById(page, 'email-input');
			expect(input).toHaveValue('');
			await user.type(input, 'person@example.com');
			await clickNext(page, user);

			expect(page.getByRole('button', { name: 'Resend in 1m 0s' })).toBeDisabled();
			await advanceClock(minute);
			expect(page.getByRole('button', { name: 'Resend link' })).toBeEnabled();
		});

		it('clears all local throttling windows when returning to email entry and submitting a different email', async () => {
			const { page, user } = await openCheckEmailStep();
			await reachDailyLimit(page, user);

			const backButton = page.getByRole('button', { name: 'Back to email entry' });
			await user.click(backButton);
			await user.type(getInputById(page, 'email-input'), 'other@example.com');
			await clickNext(page, user);

			expect(page.getByText('other@example.com', { exact: true })).toBeInTheDocument();
			expect(page.getByRole('button', { name: 'Resend in 1m 0s' })).toBeDisabled();
			await advanceClock(minute);
			expect(page.getByRole('button', { name: 'Resend link' })).toBeEnabled();
		});

		it('successfully requests resend', async () => {
			const { page, user } = await openCheckEmailStep();
			await advanceClock(minute);

			const resendButton = page.getByRole('button', { name: 'Resend link' });
			await user.click(resendButton);

			expect(fetchMock).toHaveBeenCalledTimes(2);
			expect(fetchMock).toHaveBeenNthCalledWith(2, '/api/verify-email', {
				method: 'POST',
				credentials: 'same-origin',
				headers: expect.any(Headers),
				body: JSON.stringify({ email: 'person@example.com' })
			});
			expectRequestContentType(fetchMock, 'application/json', 1);
			expect(page.getByRole('button', { name: 'Resend in 1m 0s' })).toBeDisabled();
		});
	});

	describe('password step', () => {
		it.todo('exposes sign-in navigation');

		it.todo('reacts correctly to too short password');
		it.todo('reacts correctly to too long password');
		it.todo('stops displaying password validation errors on editing');
		it.todo('reacts correctly to too short repeated password');
		it.todo('reacts correctly to too long repeated password');
		it.todo('stops displaying repeated password validation errors on editing');
		it.todo('reacts correctly to password mismatch');

		it.todo('reacts correctly to too short token');
		it.todo('reacts correctly to long short token');
		it.todo('reacts correctly to invalid characters in token');

		it.todo('reacts correctly to backend error response');
		it.todo('reacts correctly to request failure before receiving a response');
		it.todo('reacts correctly while sign-up request is loading');

		it.todo('successfully signs up');
	});
});

const fetchMock = vi.fn();
const gotoMock = vi.mocked(goto);
const toastErrorMock = vi.mocked(toast.error);

async function clickNext(page: RenderResult<typeof Page>, user: ReturnType<typeof userEvent.setup>): Promise<void> {
	const nextButton = page.getByRole('button', { name: 'Next' });
	await user.click(nextButton);
	await tick();
}

async function expectInCheckEmailStep(page: RenderResult<typeof Page>): Promise<void> {
	expect(await page.findByRole('button', { name: 'Back to email entry' })).toBeInTheDocument();
}

const second = 1000;
const minute = 60 * second;
const hour = 60 * minute;
const day = 24 * hour;

async function openCheckEmailStep(): Promise<{ page: RenderResult<typeof Page>; user: ReturnType<typeof userEvent.setup> }> {
	const user = userEvent.setup();
	fetchMock.mockImplementation(() => Promise.resolve(new Response(undefined, { status: 204 })));
	const page = render(Page);
	await user.type(getInputById(page, 'email-input'), 'person@example.com');
	await clickNext(page, user);
	await expectInCheckEmailStep(page);
	return { page, user };
}

async function advanceClock(milliseconds: number): Promise<void> {
	vi.advanceTimersByTime(milliseconds);
	await tick();
}

async function makeResendRequests(
	page: RenderResult<typeof Page>,
	user: ReturnType<typeof userEvent.setup>,
	delays: number[]
): Promise<void> {
	for (const delay of delays) {
		await advanceClock(delay);
		const resendButton = page.getByRole('button', { name: 'Resend link' });
		await user.click(resendButton);
	}
}

async function reachDailyLimit(page: RenderResult<typeof Page>, user: ReturnType<typeof userEvent.setup>): Promise<void> {
	await makeResendRequests(page, user, [minute, minute, minute, minute, hour - 4 * minute, minute, minute, minute, minute]);
}
