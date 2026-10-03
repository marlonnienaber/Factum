<script lang="ts">
	import { LoaderCircle } from '@lucide/svelte';
	import type { Snippet } from 'svelte';

	type Props = {
		children: Snippet;
		loadMore: () => void | Promise<void>;
		hasMore: boolean;
		disabled?: boolean;
	};

	const { children, loadMore, hasMore, disabled = false }: Props = $props();
	let scrollContainer = $state<HTMLDivElement>();
	let loadingSentinel = $state<HTMLDivElement>();
	let loading = $state(false);

	async function handleLoadMore(): Promise<void> {
		if (loading) return;
		loading = true;
		try {
			await loadMore();
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		if (scrollContainer === undefined || loadingSentinel === undefined || disabled || !hasMore) return;

		const observer = new IntersectionObserver(
			(entries) => {
				if (entries.some((entry) => entry.isIntersecting)) void handleLoadMore();
			},
			{ root: scrollContainer }
		);
		observer.observe(loadingSentinel);

		return () => observer.disconnect();
	});
</script>

<div class="scroll-container" bind:this={scrollContainer}>
	<div class="content">
		{@render children()}
	</div>
	{#if hasMore}
		<div class="loading-sentinel" bind:this={loadingSentinel}>
			{#if loading}
				<LoaderCircle size={18} aria-label="Loading more items" />
			{/if}
		</div>
	{/if}
</div>

<style lang="scss">
	.scroll-container {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		scrollbar-width: none;

		&::-webkit-scrollbar {
			display: none;
		}

		.content {
			display: grid;
			min-height: 100%;
		}

		.loading-sentinel {
			display: flex;
			min-height: 1px;
			margin-top: -1px;
			align-items: center;
			justify-content: center;

			:global(svg) {
				margin: 0.75rem;
				color: var(--text-subtle);
				animation: spin 1s linear infinite;
			}
		}
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
