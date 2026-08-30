<script lang="ts">
	import {page} from '$app/state';
</script>

<div class="error-tag">
	<span class="stamp">Error</span>
	<h1>{page.status}</h1>
	<p>{page.error?.message}</p>
</div>

<style>
	.error-tag {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.stamp {
		font-family: 'Allerta Stencil', monospace;
		font-size: 0.85rem;
		letter-spacing: 0.1em;
		color: #d9531e;
	}
</style>
