<script lang="ts">
	import { resolve } from '$app/paths';
	import type { RouteId } from '$app/types';

	let {
		partNumber,
		name,
		description,
		href,
		tags
	}: { partNumber: string; name: string; description: string; href: RouteId; tags: string[] } =
		$props();
</script>

<a aria-label={name} class="tool-tag contrast" href={resolve(href)}>
	<span aria-hidden="true" class="grommet"></span>
	<span class="part-number">#{partNumber}</span>
	<h3>{name}</h3>
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
		padding: 0.7rem 1.25rem 1.25rem 1.75rem;
		background-color: var(--pico-secondary-background);
		border: 1px solid var(--pico-primary);
		border-radius: 0.25rem;
		text-decoration: none;
		transition: box-shadow 0.15s ease;
	}

	.tool-tag:hover,
	.tool-tag:focus-visible {
		box-shadow: 0 6px 16px rgba(0, 0, 0, 0.35);
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
		border: 0.1rem solid #4a4740;
		background: #1c1b19;
	}

	.part-number {
		font-family: 'Allerta Stencil', monospace;
		font-size: 0.75rem;
		letter-spacing: 0.08em;
		color: var(--pico-primary);
	}

	.description {
		font-size: 0.7rem;
	}
</style>
