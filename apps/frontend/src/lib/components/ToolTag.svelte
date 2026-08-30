<script lang="ts">
    let {
        partNumber,
        name,
        description,
        href
    }: { partNumber: string; name: string; description: string; href: string } = $props();
</script>

<a aria-label={name} class="tool-tag" {href}>
    <span aria-hidden="true" class="grommet"></span>
    <span class="part-number">#{partNumber}</span>
    <span class="name">{name}</span>
    <span class="description">{description}</span>
</a>

<style>
    .tool-tag {
        position: relative;
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
        padding: 1.5rem 1.25rem 1.25rem 1.75rem;
        background: #ede7da;
        border: 1px solid #4a4740;
        border-radius: 0.25rem;
        color: #2b2a28;
        text-decoration: none;
        transition: transform 0.15s ease,
        box-shadow 0.15s ease;
    }

    .tool-tag:hover,
    .tool-tag:focus-visible {
        transform: translateY(-2px);
        box-shadow: 0 6px 16px rgba(0, 0, 0, 0.35);
        color: #2b2a28;
    }

    @media (prefers-reduced-motion: reduce) {
        .tool-tag {
            transition: none;
        }

        .tool-tag:hover,
        .tool-tag:focus-visible {
            transform: none;
        }
    }

    .grommet {
        position: absolute;
        top: 0.9rem;
        left: 0.7rem;
        width: 0.6rem;
        height: 0.6rem;
        border-radius: 50%;
        border: 1.5px solid #4a4740;
        background: #1c1b19;
    }

    .part-number {
        font-family: 'Allerta Stencil', monospace;
        font-size: 0.75rem;
        letter-spacing: 0.08em;
        color: #d9531e;
    }

    .name {
        font-family: 'Allerta Stencil', monospace;
        font-size: 1.15rem;
        text-transform: uppercase;
        letter-spacing: 0.03em;
    }

    .description {
        font-size: 0.85rem;
        color: #5c7a8a;
    }
</style>
