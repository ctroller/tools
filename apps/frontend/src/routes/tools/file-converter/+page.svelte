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
	import { Trash } from '@lucide/svelte';

	interface FileInfo {
		handle: string;
		detectedSource: string;
		targets: string[];
	}

	interface FileUpload {
		id: string;
		status: 'uploading' | 'done' | 'error';
		name: string;
		progress: number;
		error?: string;

		info?: FileInfo;
		target?: string;
		convPromise?: Promise<JobStatusResult>;
	}

	let uploadedFiles: FileUpload[] = $state([]);
	let errors: ProblemDetails[] = $state([]);
	let conversion: boolean = $state(false);

	function ondrop(files: File[]) {
		const f: FileUpload = {
			id: crypto.randomUUID(),
			name: files[0].name,
			status: 'uploading',
			progress: 0
		};

		uploadedFiles.push(f);

		uploadFile(files[0], (sent, total) => {
			const entry = uploadedFiles.find((file) => file.id === f.id);
			if (entry) {
				entry.progress = sent / total;
			}
		}).then(
			(res) => {
				const entry = uploadedFiles.find((file) => file.id === f.id)!;
				if (res.success) {
					entry.info = { ...res.data };
					entry.target = res.data.targets[0];
					entry.status = 'done';
				} else {
					entry.error = res.detail;
					entry.status = 'error';
				}
			},
			(err) => {
				const entry = uploadedFiles.find((file) => file.id === f.id)!;
				entry.error = err.detail;
				entry.status = 'error';
			}
		);
	}

	function convert() {
		conversion = true;
		uploadedFiles
			.filter((file) => file.status === 'done')
			.filter((file) => !file.convPromise)
			.forEach((file) => {
				const handle = file.info!.handle;
				startConversion(handle, file.target!).then((res) => {
					if (res.success) {
						file.convPromise = fetchJobStatus(handle).then((r) => {
							if (!r.success) throw r;
							return r.data;
						});
					} else {
						file.error = res.detail;
						file.status = 'error';
					}
				});
			});
	}

	function remove(id: string) {
		uploadedFiles = uploadedFiles.filter((f) => f.id !== id);
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
					<th></th>
				</tr>
			</thead>
			<tbody>
				{#each uploadedFiles as file (file.id)}
					<tr>
						<td>{file.name}</td>
						{#if file.status === 'uploading'}
							<td colspan="2"><span aria-busy="true">Uploading...</span></td>
							<td>
								<progress value={file.progress} max="1"></progress>
							</td>
						{:else if file.status === 'error'}
							<td colspan="2">{file.error}</td>
							<td>
								<Trash class="interactive-icon" onclick={() => remove(file.id)} />
							</td>
						{:else}
							<td>{file.info!.detectedSource}</td>
							<td>
								{#if file.convPromise !== undefined}
									{file.target}
								{:else}
									<select bind:value={file.target}>
										{#each file.info!.targets as target (target)}
											<option value={target}>{target}</option>
										{/each}
									</select>
								{/if}
							</td>
							{#if conversion}
								<td>
									{#await file.convPromise}
										<span aria-busy="true">Converting...</span>
									{:then jobStatus}
										{#if jobStatus?.status === 'done'}
											<a
												href="/api/file-converter/{file.info!.handle}/download"
												target="_blank"
												rel="external">Download</a
											>
										{:else}
											<span aria-busy="true">Converting...</span>
										{/if}
									{:catch pd}
										{pd.detail}
									{/await}
								</td>
							{:else}
								<td>
									<Trash class="interactive-icon" onclick={() => remove(file.id)} />
								</td>
							{/if}
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

<style>
	:global(.interactive-icon):hover {
		cursor: pointer;
	}
</style>
