<script lang="ts">
    import type {UndatedTaskResponse} from "$lib/types/responses/undated-task-response";
    import Checkbox from "$lib/components/shared/checkbox.svelte";
    import { categoryStore } from '$lib/stores/category-store.svelte';
    import Category from "$lib/components/shared/category.svelte";
    import { Ellipsis } from '@lucide/svelte';
    import type {DatedUntrashedOpenTaskResponse} from "$lib/types/responses/dated-untrashed-open-task-response";

    type Props = {
        task: UndatedTaskResponse | DatedUntrashedOpenTaskResponse
    }

    const { task }: Props = $props()
    let checked = $state(false);
    const category = $derived(categoryStore.categories.find((category) => category.id === task.categoryId));
</script>

<div class="frame">
    <div class="header">
        <span class="title">{task.title}</span>
        <div class="controls">
            <button class="more-button" type="button" aria-label={`More options for "${task.title}"`}>
                <Ellipsis size={18} />
            </button>
            <Checkbox bind:checked ariaLabel={`Mark "${task.title}" as completed`} />
        </div>
    </div>
    {#if task.description}
        <span class="description">{task.description}</span>
    {/if}
    {#if category}
        <Category {category} />
    {/if}
</div>

<style lang="scss">
    .frame {
        display: flex;
        flex-direction: column;
        background-color: var(--surface);
        box-sizing: border-box;
        padding: 1rem;
        border-radius: 0.5rem;
        gap: 0.75rem;

        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;


            .title {
                line-height: 1rem;
                font-size: 1rem;
            }

            .controls {
                display: flex;
                align-items: center;
                gap: 0.5rem;

                .more-button {
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    padding: 0;
                    border: 0;
                    color: inherit;
                    background: transparent;
                    cursor: pointer;
                }
            }
        }

        .description {
            font-size: 0.75rem;
        }
    }
</style>
