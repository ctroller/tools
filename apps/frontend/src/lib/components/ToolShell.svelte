<script lang="ts">
	import type {Snippet} from 'svelte';
	import {page} from '$app/state';
	import {Tools} from '$lib/tools';

	let {children}: { children: Snippet } = $props();
	const tool = Tools.find((t) => t.href === page.url.pathname)!;
</script>

<div class="tool-page">
	<h1><span class="part-number">#{tool.partNumber}</span> - {tool.name}</h1>
	{@render children()}
</div>

<style>
	.tool-page {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.part-number {
		color: #d9531e;
	}
</style>
