<script lang="ts">
	import type { CategoryResponse } from '$lib/types/responses/category-response';

	type Props = {
		category: CategoryResponse;
	};

	let { category }: Props = $props();
</script>

<span class="category" style:background-color={category.color}>{category.title}</span>

<style lang="scss">
	.category {
		display: inline-flex;
		align-items: center;
		width: fit-content;
        box-sizing: border-box;
		padding: 0.125rem 0.375rem;
		border-radius: 1rem;
		color: var(--text);
		font-size: 0.75rem;
	}
</style>
