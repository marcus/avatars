<script lang="ts">
	import { agentPortrait } from './agent-portrait';
	let { agentId, size = 24, fillHeader = false }: { agentId: string; size?: number; fillHeader?: boolean } = $props();
	const portrait = $derived(
		agentPortrait(agentId).replace('<svg ', fillHeader ? '<svg preserveAspectRatio="xMidYMid slice" ' : '<svg ')
	);
</script>

<span class="agent-portrait" class:fill-header={fillHeader} style:width={`${size}px`} style:height={`${size}px`} aria-hidden="true">
	{@html portrait}
</span>

<style>
	.agent-portrait {
		display: inline-block;
		flex-shrink: 0;
		overflow: hidden;
		border: 1px solid #bbb49a55;
		background: #d6d0bb;
		border-radius: 50%;
	}
	.agent-portrait :global(svg) {
		display: block;
		width: 100%;
		height: 100%;
	}
	.fill-header {
		position: relative;
		align-self: center;
	}
	.fill-header :global(svg) {
		position: absolute;
		inset: 0;
	}
</style>
