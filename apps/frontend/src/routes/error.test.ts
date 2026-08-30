import {render, screen} from '@testing-library/svelte';
import {describe, expect, it, vi} from 'vitest';
import ErrorPage from './+error.svelte';

vi.mock('$app/state', () => ({
	page: {status: 404, error: {message: 'Not Found'}}
}));

describe('+error.svelte', () => {
	it('shows the status code and error message', () => {
		render(ErrorPage);
		expect(screen.getByText('404')).toBeTruthy();
		expect(screen.getByText('Not Found')).toBeTruthy();
	});
});
