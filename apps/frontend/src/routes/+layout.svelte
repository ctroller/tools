<script lang="ts">
	import type {Snippet} from 'svelte';
	import '$lib/styles/fonts.css';
	import '$lib/styles/theme.css';
	import SiteHeader from '$lib/components/SiteHeader.svelte';

	let { children }: { children: Snippet } = $props();
</script>

<SiteHeader />
<main class="container-fluid">
	{@render children()}
</main>
