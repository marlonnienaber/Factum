import '@testing-library/jest-dom/vitest';
import { vi } from 'vitest';

vi.mock('$app/navigation');

vi.mock('$app/paths', () => ({
	resolve: (path: string) => path
}));

vi.mock('svelte-sonner');
