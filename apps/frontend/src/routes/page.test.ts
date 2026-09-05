import {cleanup, fireEvent, render, screen} from '@testing-library/svelte';
import {afterEach, describe, expect, it} from 'vitest';
import Page from './+page.svelte';

afterEach(cleanup);

describe('landing page', () => {
	it('links to the file-converter tool', () => {
		render(Page);
		const link = screen.getByRole('link', { name: 'File Converter' });
		expect(link).toBeTruthy();
		expect(link.getAttribute('href')).toBe('/tools/file-converter');
	});

	it('lists every tag alphabetically with its tool count', () => {
		render(Page);
		const chips = screen.getAllByRole('button', {name: /\(\d+\)$/});
		expect(chips.map((chip) => chip.textContent)).toEqual(['converter (1)', 'file (1)']);
	});

	it('toggles a tag filter from the tag bar', async () => {
		render(Page);
		const chip = screen.getByRole('button', {name: 'converter (1)'});

		expect(chip.getAttribute('aria-pressed')).toBe('false');
		expect(screen.getByRole('link', {name: 'File Converter'})).toBeTruthy();

		await fireEvent.click(chip);
		expect(chip.getAttribute('aria-pressed')).toBe('true');
		expect(screen.getByRole('link', {name: 'File Converter'})).toBeTruthy();

		await fireEvent.click(chip);
		expect(chip.getAttribute('aria-pressed')).toBe('false');
	});
});
