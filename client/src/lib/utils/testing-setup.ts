import '@testing-library/jest-dom/vitest';
import { configMocks } from 'jsdom-testing-mocks';
import { afterAll, vi } from 'vitest';

HTMLElement.prototype.hasPointerCapture ??= () => false;
HTMLElement.prototype.setPointerCapture ??= () => {};
HTMLElement.prototype.releasePointerCapture ??= () => {};

configMocks({ afterAll });

vi.mock('$app/navigation');

vi.mock('$app/paths', () => ({
	resolve: (path: string) => path
}));

vi.mock('svelte-sonner');
