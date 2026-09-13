<script lang="ts">
	import ToolShell from '$lib/components/ToolShell.svelte';
	import FileDrop from '$lib/components/FileDrop.svelte';
	import {
		fetchJobStatus,
		type JobStatusResult,
		startConversion,
		uploadFile
	} from '$lib/api/file-converter';
	import type { ProblemDetails } from '$lib/response-types';

	interface FileUpload {
		handle: string;
		name: string;
		detectedSource: string;
		targets: string[];
		target: string;
		promise: Promise<JobStatusResult>;
	}

	let uploadedFiles: FileUpload[] = $state([]);
	let errors: ProblemDetails[] = $state([]);
	let conversion: boolean = $state(false);

	function ondrop(files: File[]) {
		uploadFile(files[0]).then((res) => {
			if (res.success) {
				uploadedFiles.push({
					target: res.data.targets[0],
					...res.data
				} as FileUpload);
			} else {
				errors.push(res);
			}
		});
	}

	function convert() {
		conversion = true;
		uploadedFiles
			.filter((file) => !file.promise)
			.forEach((file) => {
				startConversion(file.handle, file.target).then((res) => {
					if (res.success) {
						file.promise = fetchJobStatus(file.handle).then((r) => {
							if (!r.success) throw r;
							return r.data;
						});
					} else {
						errors.push(res);
					}
				});
			});
	}
</script>

<ToolShell>
	<FileDrop {ondrop} />
	{#if uploadedFiles.length > 0}
		<table class="file-handler">
			<thead>
				<tr>
					<th>Name</th>
					<th>Source</th>
					<th>Target</th>
					{#if conversion}
						<th></th>
					{/if}
				</tr>
			</thead>
			<tbody>
				{#each uploadedFiles as file (file.handle)}
					<tr>
						<td>{file.name}</td>
						<td>{file.detectedSource}</td>
						<td>
							{#if file.promise !== undefined}
								{file.target}
							{:else}
								<select bind:value={file.target}>
									{#each file.targets as target (target)}
										<option value={target}>{target}</option>
									{/each}
								</select>
							{/if}
						</td>
						{#if conversion}
							<td>
								{#await file.promise}
									Converting...
								{:then jobStatus}
									{#if jobStatus?.status === 'done'}
										<a
											href="/api/file-converter/{file.handle}/download"
											target="_blank"
											rel="external">Download</a
										>
									{:else}
										{JSON.stringify(jobStatus)}
									{/if}
								{:catch pd}
									{pd.detail}
								{/await}
							</td>
						{/if}
					</tr>
				{/each}
			</tbody>
			<tfoot>
				<tr>
					<td colspan="4">
						<button onclick={convert}>Convert Files</button>
					</td>
				</tr>
			</tfoot>
		</table>
	{/if}
	{#if errors.length > 0}
		{#each errors as error (error)}
			<article>
				<header>{error.title}</header>
				{error.detail}
			</article>
		{/each}
	{/if}
</ToolShell>
