import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import SiteHeader from './SiteHeader.svelte';

describe('SiteHeader', () => {
	it('renders a link back to the tool list', () => {
		render(SiteHeader);
		const link = screen.getByRole('link', { name: 'Tools' });
		expect(link).toBeTruthy();
		expect(link.getAttribute('href')).toBe('/');
	});
});
