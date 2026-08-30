<script lang="ts">
	import type { Snippet } from 'svelte';
	import SiteHeader from '$lib/components/SiteHeader.svelte';

	let { children }: { children: Snippet } = $props();
</script>

<SiteHeader />
<main>
	{@render children()}
</main>
