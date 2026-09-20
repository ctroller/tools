<script lang="ts">
	import ConversionStatus from './ConversionStatus.svelte';
	import { Trash } from '@lucide/svelte';

	export interface FileInfo {
		handle: string;
		detectedSource: string;
		targets: string[];
	}

	export interface FileUpload {
		id: string;
		status: 'uploading' | 'done' | 'error';
		name: string;
		progress: number | undefined;
		error?: string;

		info?: FileInfo;
		target?: string;
		converted?: boolean;
	}

	let {
		files,
		remove,
		convert
	}: { files: FileUpload[]; remove: (id: string) => void; convert: () => void } = $props();
</script>

{#if files.length > 0}
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
			{#each files as file (file.id)}
				<tr>
					<td>{file.name}</td>
					{#if file.status === 'uploading'}
						<td colspan="2"><span aria-busy="true">Uploading...</span></td>
						<td>
							{#if file.progress !== undefined}
								<progress value={file.progress} max="1"></progress>
							{:else}
								<progress></progress>
							{/if}
						</td>
					{:else if file.status === 'error'}
						<td colspan="2">{file.error}</td>
						<td>
							<Trash class="interactive-icon" onclick={() => remove(file.id)} />
						</td>
					{:else}
						<td>{file.info!.detectedSource}</td>
						<td>
							{#if file.converted}
								{file.target}
							{:else}
								<select bind:value={file.target}>
									{#each file.info!.targets as target (target)}
										<option value={target}>{target}</option>
									{/each}
								</select>
							{/if}
						</td>
						<td>
							{#if file.converted}
								<ConversionStatus handle={file.info!.handle} />
							{:else}
								<button
									type="button"
									class="outline"
									aria-label="Remove file"
									onclick={() => remove(file.id)}
								>
									<Trash />
								</button>
							{/if}
						</td>
					{/if}
				</tr>
			{/each}
		</tbody>
		<tfoot>
			<tr>
				<td colspan="4">
					<button
						onclick={convert}
						disabled={!files.some((f) => f.info !== undefined && !f.converted)}
						>Convert Files
					</button>
				</td>
			</tr>
		</tfoot>
	</table>
{/if}
