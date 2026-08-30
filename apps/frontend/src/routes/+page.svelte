<script lang="ts">
    import ToolTag from '$lib/components/ToolTag.svelte';
    import {Tools} from "$lib/tools";
</script>

<div class="tool-grid">
    {#each Tools as tool}
        <ToolTag {...tool}/>
    {/each}
</div>

<style>
    .tool-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
        gap: 1.25rem;
        margin-top: 1.5rem;
    }
</style>
