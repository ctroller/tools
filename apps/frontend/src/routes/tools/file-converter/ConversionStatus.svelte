<script lang="ts">
	import { downloadHref, type JobUpdate, watchJob } from '$lib/api/file-converter';

	let { handle }: { handle: string } = $props();

	let job: JobUpdate = $state({ status: 'pending' });
	$effect(() => watchJob(handle, (update) => (job = update)));
</script>

{#if job.status === 'done'}
	<a href={downloadHref(handle)} rel="external">Download</a>
{:else if job.status === 'failed'}
	{job.error ?? 'Conversion failed'}
{:else}
	<span aria-busy="true">{job.status === 'processing' ? 'Converting...' : 'Queued...'}</span>
{/if}
