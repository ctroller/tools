<script lang="ts">
	import ToolTag from '$lib/components/ToolTag.svelte';
</script>

<h1>Tools</h1>
<div class="tool-grid">
	<ToolTag
			description="Convert files between various formats."
			href="/tools/file-converter"
			name="File Converter"
			partNumber="001"
	/>
</div>

<style>
	.tool-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
		gap: 1.25rem;
		margin-top: 1.5rem;
	}
</style>
