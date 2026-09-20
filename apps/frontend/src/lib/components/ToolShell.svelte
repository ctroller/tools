<script lang="ts">
	import { page } from '$app/state';
	import { Tools } from '$lib/tools';
	import type { Snippet } from 'svelte';

	const tool = Tools.find((t) => t.href === page.url.pathname)!;

	let { children }: { children: Snippet } = $props();
</script>

<hgroup>
	<p class="chapter">#{tool.partNumber}</p>
	<h1>{tool.name}</h1>
	{#if tool.extendedDesc}
		<p>
			{tool.extendedDesc}
		</p>
	{/if}
</hgroup>
<div id="content" role="document">
	{@render children()}
</div>
