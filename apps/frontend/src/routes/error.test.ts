import { render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import { readable } from 'svelte/store';

vi.mock('$app/stores', () => ({
	page: readable({ status: 404, error: { message: 'Not Found' } })
}));

import ErrorPage from './+error.svelte';

describe('+error.svelte', () => {
	it('shows the status code and error message', () => {
		render(ErrorPage);
		expect(screen.getByText('404')).toBeTruthy();
		expect(screen.getByText('Not Found')).toBeTruthy();
	});
});
