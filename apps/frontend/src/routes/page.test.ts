import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import Page from './+page.svelte';

describe('landing page', () => {
	it('links to the file-converter tool', () => {
		render(Page);
		const link = screen.getByRole('link', { name: 'File Converter' });
		expect(link).toBeTruthy();
		expect(link.getAttribute('href')).toBe('/tools/file-converter');
	});
});
