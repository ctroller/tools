<script lang="ts">
	import { FilePlusCorner } from '@lucide/svelte';

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
		ondrop(f);
	}}
>
	<div class="box__input">
		<input
			class="box__file"
			id="file"
			multiple
			name="files[]"
			onchange={(e) => ondrop(Array.from((e?.target as HTMLInputElement)?.files ?? []))}
			type="file"
		/>
		<label for="file">
			<span class="icon">
				<FilePlusCorner size={64} />
			</span>
			<strong>Choose a file</strong> <span class="box__dragdrop">or drag it here</span>.</label
		>
	</div>
</form>

<style>
	.box {
		font-family: var(--pico-font-family), sans-serif;
		max-width: 32rem;
		margin: 0 auto;
		padding: 2rem 1.5rem;
		text-align: center;
		color: var(--pico-secondary);
		border: 1px solid var(--pico-muted-border-color);
		border-radius: var(--pico-border-radius);
		background-color: var(--pico-secondary-background);
	}

	.box strong {
		color: var(--pico-secondary-inverse);
	}

	.box.has-advanced-upload {
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

	.box.is-dragover {
		background-color: var(--pico-secondary-hover);
		outline-color: var(--pico-secondary);
	}

	.box.is-dragover .icon {
		color: var(--pico-secondary);
	}

	.box.is-dragover span {
		color: var(--pico-secondary-inverse);
	}

	.icon {
		display: block;
		color: var(--pico-primary);
	}
</style>
