import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import Page from './+page.svelte';

describe('file-converter placeholder page', () => {
	it('renders a heading and a not-yet-built notice', () => {
		render(Page);
		expect(screen.getByRole('heading', { name: 'File Converter' })).toBeTruthy();
		expect(screen.getByText('Coming soon.')).toBeTruthy();
	});
});
