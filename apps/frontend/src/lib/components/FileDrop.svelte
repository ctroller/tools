<script lang="ts">
	const advancedUpload = () => {
		const div = document.createElement('div');
		const isDragDropSupported = 'draggable' in div || ('ondragstart' in div && 'ondrop' in div);

		const isFormDataSupported = 'FormData' in window;
		const isFileReaderSupported = 'FileReader' in window;

		return isDragDropSupported && isFormDataSupported && isFileReaderSupported;
	};

	const prevent = (e: Event) => {
		e.preventDefault();
		e.stopPropagation();
	};

	let dragCount = $state(0);

	const ondragenter = (e: Event) => {
		prevent(e);
		dragCount++;
	};

	const ondragleave = (e: Event) => {
		prevent(e);
		dragCount--;
		if (dragCount <= 0) {
			dragCount = 0;
		}
	};

	const ondragend = (e: Event) => {
		prevent(e);
		dragCount = 0;
	};

	let { ondrop = () => {} }: { ondrop: (files: File[]) => void } = $props();

	let count = 0;
	let files: File[] = $state([]);
</script>

<form
	action=""
	class="box"
	class:has-advanced-upload={advancedUpload}
	class:is-dragover={dragCount > 0}
	enctype="multipart/form-data"
	method="post"
	ondrag={prevent}
	{ondragend}
	{ondragenter}
	{ondragleave}
	ondragover={prevent}
	ondragstart={prevent}
	ondrop={(e) => {
		ondragend(e);
		const f: File[] = Array.from(e.dataTransfer?.files ?? []);
		files.push(...f);
		ondrop(f);
	}}
>
	<div class="box__input">
		<input
			class="box__file"
			data-multiple-caption="{count} files selected"
			id="file"
			multiple
			name="files[]"
			type="file"
		/>
		<label for="file"
			><strong>Choose a file</strong> <span class="box__dragdrop">or drag it here</span>.</label
		>
		<button class="box__button" type="submit">Upload</button>
	</div>
	<div class="box__files">
		{#each files as file, i (i)}
			<div class="box__files-file file-{i}">{file.name}</div>
		{/each}
	</div>
	<div class="box__uploading">Uploading…</div>
	<div class="box__success">Done!</div>
	<div class="box__error">Error!</div>
</form>

<style>
	.box {
		font-family: var(--pico-font-family);
		max-width: 32rem;
		margin: 0 auto;
		padding: 2rem 1.5rem;
		text-align: center;
		color: var(--pico-color);
		border: 1px solid var(--pico-muted-border-color);
		border-radius: var(--pico-border-radius);
	}

	.box__dragdrop,
	.box__files,
	.box__uploading,
	.box__success,
	.box__error {
		display: none;
	}

	.box.has-advanced-upload {
		background-color: #ede7da;
		color: #2b2a28;
		border-color: transparent;
		outline: 2px dashed var(--pico-primary);
		outline-offset: -10px;
		transition:
			background-color 0.15s ease,
			outline-color 0.15s ease;
	}
	.box.has-advanced-upload .box__dragdrop {
		display: inline;
	}

	.box.has-advanced-upload .box__files {
		display: inline-block;
	}

	.box__input {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.6rem;
	}

	.box__file {
		width: 0.1px;
		height: 0.1px;
		opacity: 0;
		overflow: hidden;
		position: absolute;
		z-index: -1;
	}

	label {
		cursor: pointer;
		font-size: 0.95rem;
	}

	label strong {
		color: var(--pico-primary);
		text-decoration: underline;
	}

	.box__file:focus-visible + label {
		outline: 2px solid var(--pico-primary);
		outline-offset: 4px;
	}

	.box__button {
		font-family: 'Allerta Stencil', var(--pico-font-family), sans-serif;
		text-transform: uppercase;
		letter-spacing: 0.03em;
		font-size: 0.85rem;
		background: var(--pico-primary-background);
		color: #ede7da;
		border: none;
		border-radius: var(--pico-border-radius);
		padding: 0.5rem 1.25rem;
		cursor: pointer;
		transition: background-color 0.15s ease;
	}

	.box__button:hover {
		background: var(--pico-primary-hover-background);
	}

	.box__uploading,
	.box__success,
	.box__error {
		margin-top: 0.75rem;
		font-size: 0.85rem;
	}

	.box__error {
		color: var(--pico-primary);
	}

	.box.is-dragover {
		background-color: #000;
		outline-color: var(--pico-primary-hover);
	}
</style>
