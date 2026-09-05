<script lang="ts">
	import ToolTag from '$lib/components/ToolTag.svelte';
	import {type Tool, Tools} from '$lib/tools';

	let filters: string[] = $state([]);

	function toggleFilter(tag: string) {
		filters = filters.includes(tag) ? filters.filter((f) => f !== tag) : [...filters, tag];
	}

	let filtered: Tool[] = $derived(
			Tools.filter((tool) => filters.every((tag) => tool.tags.includes(tag)))
	);

	let allTags: { tag: string; count: number }[] = $derived(
			Array.from(new Set(Tools.flatMap((tool) => tool.tags)))
					.sort((a, b) => a.localeCompare(b))
					.map((tag) => ({tag, count: filtered.filter((tool) => tool.tags.includes(tag)).length}))
	);
</script>

<div class="tag-filters">
	{#each allTags as {tag, count} (tag)}
		<button
				type="button"
				class="tag-filter"
				class:active={filters.includes(tag)}
				aria-pressed={filters.includes(tag)}
				onclick={() => toggleFilter(tag)}>{tag} ({count})
		</button
		>
	{/each}
</div>

<div class="tool-grid">
	{#each filtered as tool (tool.href)}
		<ToolTag {...tool}/>
	{/each}
</div>

<style>
	.tag-filters {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.35rem;
		margin-top: 1.25rem;
	}

	.tag-filter {
		background: #4a4740;
		color: #ede7da;
		font-size: 0.75rem;
		line-height: normal;
		padding: 0.2rem 0.5rem;
		border: none;
		border-radius: 0.2rem;
		cursor: pointer;
	}

	.tag-filter.active {
		background: #d9531e;
		color: #1c1b19;
	}

	.tool-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
		gap: 1.25rem;
		margin-top: 1.5rem;
	}
</style>
