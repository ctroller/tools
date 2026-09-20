<script lang="ts">
	import ToolShell from '$lib/components/ToolShell.svelte';
	import FileDrop from '$lib/components/FileDrop.svelte';
	import { deleteJob, downloadUrl, startConversion, uploadFile } from '$lib/api/file-converter';
	import FileTracker, { type FileUpload } from './FileTracker.svelte';

	let uploadedFiles: FileUpload[] = $state([]);

	function onpaste(data: DataTransfer | null) {
		let files: File[] = [...(data?.files || [])];
		if (files.length > 0) {
			ondrop(files);
		} else if (data) {
			const rawUrl = data.getData('text');
			if (!URL.canParse(rawUrl)) return;

			const url = new URL(rawUrl);
			const f: FileUpload = {
				id: crypto.randomUUID(),
				name: url.pathname.substring(url.pathname.lastIndexOf('/') + 1),
				status: 'uploading',
				progress: undefined
			};
			uploadedFiles.push(f);

			downloadUrl(url.toString()).then(
				(res) => {
					const entry = uploadedFiles.find((file) => file.id === f.id);
					if (!entry) return;
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
					const entry = uploadedFiles.find((file) => file.id === f.id);
					if (!entry) return;
					entry.error = err.detail;
					entry.status = 'error';
				}
			);
		}
	}

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
				const entry = uploadedFiles.find((file) => file.id === f.id);
				if (!entry) return;
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
				const entry = uploadedFiles.find((file) => file.id === f.id);
				if (!entry) return;
				entry.error = err.detail;
				entry.status = 'error';
			}
		);
	}

	function convert() {
		uploadedFiles
			.filter((file) => file.info !== undefined && !file.converted)
			.forEach((file) => {
				const handle = file.info!.handle;
				startConversion(handle, file.target!).then((res) => {
					file.converted = true;
					if (!res.success) {
						file.error = res.detail;
						file.status = 'error';
					}
				});
			});
	}

	function remove(id: string) {
		const entry = uploadedFiles.find((file) => file.id === id);
		if (entry) {
			uploadedFiles = uploadedFiles.filter((f) => f.id !== id);
			if (entry.info?.handle) {
				deleteJob(entry.info!.handle);
			}
		}
	}
</script>

<ToolShell>
	<section>
		<FileDrop {ondrop} {onpaste} />
	</section>
	<section>
		<FileTracker {convert} files={uploadedFiles} {remove} />
	</section>
</ToolShell>
