<script lang="ts">
	import ToolShell from '$lib/components/ToolShell.svelte';
	import FileDrop from '$lib/components/FileDrop.svelte';

	function ondrop(files: File[]) {
		console.log(files);
	}
</script>

<ToolShell>
	<FileDrop {ondrop} />
</ToolShell>
