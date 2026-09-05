import {cleanup, fireEvent, render, screen} from '@testing-library/svelte';
import {afterEach, describe, expect, it, vi} from 'vitest';
import Page from './+page.svelte';

vi.mock('$lib/tools', () => ({
    Tools: [
        {
            partNumber: '001',
            name: 'Alpha',
            description: 'a',
            href: '/tools/alpha',
            tags: ['shared', 'only-a']
        },
        {
            partNumber: '002',
            name: 'Beta',
            description: 'b',
            href: '/tools/beta',
            tags: ['shared', 'only-b']
        }
    ]
}));

afterEach(cleanup);

describe('tag counts', () => {
    it('shows total counts with no filters active', () => {
        render(Page);
        expect(screen.getByRole('button', {name: 'shared (2)'})).toBeTruthy();
        expect(screen.getByRole('button', {name: 'only-a (1)'})).toBeTruthy();
        expect(screen.getByRole('button', {name: 'only-b (1)'})).toBeTruthy();
    });

    it('recomputes counts against the currently filtered tools', async () => {
        render(Page);

        await fireEvent.click(screen.getByRole('button', {name: 'only-a (1)'}));

        expect(screen.getByRole('button', {name: 'shared (1)'})).toBeTruthy();
        expect(screen.getByRole('button', {name: 'only-a (1)'})).toBeTruthy();
        expect(screen.getByRole('button', {name: 'only-b (0)'})).toBeTruthy();
    });

    it('narrows results with AND semantics when multiple filters are active', async () => {
        render(Page);

        await fireEvent.click(screen.getByRole('button', {name: 'shared (2)'}));
        await fireEvent.click(screen.getByRole('button', {name: 'only-b (1)'}));

        expect(screen.getByRole('link', {name: 'Beta'})).toBeTruthy();
        expect(screen.queryByRole('link', {name: 'Alpha'})).toBeNull();
        expect(screen.getByRole('button', {name: 'only-a (0)'})).toBeTruthy();
    });
});
