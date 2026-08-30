<script lang="ts">
	import {page} from '$app/state';
</script>

<div class="error-tag">
	<h1>{page.status}</h1>
	<p>{page.error?.message}</p>
</div>

<style>
	.error-tag {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
</style>
