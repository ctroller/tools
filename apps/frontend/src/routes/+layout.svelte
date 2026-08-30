<script lang="ts">
	import type {Snippet} from 'svelte';
	import '$lib/styles/fonts.css';
	import '$lib/styles/theme.css';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import SiteFooter from '$lib/components/SiteFooter.svelte';

	let { children }: { children: Snippet } = $props();
</script>

<SiteHeader />
<main class="container-fluid">
	{@render children()}
</main>
<SiteFooter/>

<style>
	main {
		flex: 1;
	}
</style>
