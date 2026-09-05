<script lang="ts">
	import {resolve} from '$app/paths';
	import type {RouteId} from '$app/types';

	let {
		partNumber,
		name,
		description,
		href,
		tags
	}: { partNumber: string; name: string; description: string; href: RouteId; tags: string[] } =
			$props();
</script>

<a aria-label={name} class="tool-tag" href={resolve(href)}>
	<span aria-hidden="true" class="grommet"></span>
	<span class="part-number">#{partNumber}</span>
	<span class="name">{name}</span>
	<span class="description">{description}</span>
	<div class="tags">
		{#each tags as tag (tag)}
			<span class="tag">{tag}</span>
		{/each}
	</div>
</a>

<style>
	.tool-tag {
		position: relative;
		display: flex;
		flex-direction: column;
		gap: 0.35rem;
		padding: 1.5rem 1.25rem 1.25rem 1.75rem;
		background: #ede7da;
		border: 1px solid #4a4740;
		border-radius: 0.25rem;
		color: #2b2a28;
		text-decoration: none;
		transition: box-shadow 0.15s ease;
	}

	.tool-tag:hover,
	.tool-tag:focus-visible {
		box-shadow: 0 6px 16px rgba(0, 0, 0, 0.35);
		color: #2b2a28;
	}

	@media (prefers-reduced-motion: reduce) {
		.tool-tag {
			transition: none;
		}
	}

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.25rem;
	}

	.tag {
		background: #4a4740;
		color: #ede7da;
		font-size: 0.5rem;
		padding: 0.1rem 0.2rem;
		border-radius: 0.25rem;
	}

	.grommet {
		position: absolute;
		top: 0.9rem;
		left: 0.7rem;
		width: 0.6rem;
		height: 0.6rem;
		border-radius: 50%;
		border: 1.5px solid #4a4740;
		background: #1c1b19;
	}

	.part-number {
		font-family: 'Allerta Stencil', monospace;
		font-size: 0.75rem;
		letter-spacing: 0.08em;
		color: #d9531e;
	}

	.name {
		font-family: 'Allerta Stencil', monospace;
		font-size: 1.15rem;
		text-transform: uppercase;
		letter-spacing: 0.03em;
	}

	.description {
		font-size: 0.85rem;
		color: #5c7a8a;
	}
</style>
