import '@testing-library/jest-dom/vitest';
import { vi } from 'vitest';

HTMLElement.prototype.hasPointerCapture ??= () => false;
HTMLElement.prototype.setPointerCapture ??= () => {};
HTMLElement.prototype.releasePointerCapture ??= () => {};

vi.mock('$app/navigation');

vi.mock('$app/paths', () => ({
	resolve: (path: string) => path
}));

vi.mock('svelte-sonner');
