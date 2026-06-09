<script lang="ts">
	import { InfoIcon } from '@lucide/svelte';
	import type { Component, ComponentProps } from 'svelte';
	import type { HTMLInputAttributes } from 'svelte/elements';
	import { InputGroup, InputGroupAddon, InputGroupInput } from '$lib/components/shadcn/input-group';
	import { Popover, PopoverContent, PopoverTrigger } from '$lib/components/shadcn/popover';

	type AuthenticationInputProps = {
		icon: Component;
		type: ComponentProps<typeof InputGroupInput>['type'];
		placeholder: string;
		autocomplete: HTMLInputAttributes['autocomplete'];
		value: string;
		error?: string;
	};

	let {
		icon: Icon,
		type,
		placeholder,
		autocomplete,
		value = $bindable(''),
		error = $bindable<string | undefined>(undefined)
	}: AuthenticationInputProps = $props();

	let errorPopoverOpen = $state(false);
</script>

<InputGroup class="h-11 gap-1 p-1.5">
	<InputGroupInput
		{type}
		{placeholder}
		{autocomplete}
		aria-invalid={error ? true : undefined}
		bind:value
		oninput={() => {
			error = undefined;
		}}
	/>
	<InputGroupAddon align="inline-start" class="text-foreground">
		<Icon />
	</InputGroupAddon>
	{#if error}
		<InputGroupAddon align="inline-end" class="text-destructive">
			<Popover bind:open={errorPopoverOpen}>
				<PopoverTrigger
					onmouseenter={() => {
						errorPopoverOpen = true;
					}}
					onmouseleave={() => {
						errorPopoverOpen = false;
					}}
					onfocus={() => {
						errorPopoverOpen = true;
					}}
					onblur={() => {
						errorPopoverOpen = false;
					}}
					aria-label={error}
					class="flex items-center justify-center"
				>
					<InfoIcon class="size-4" />
				</PopoverTrigger>
				<PopoverContent
					side="right"
					align="center"
					sideOffset={8}
					onmouseenter={() => {
						errorPopoverOpen = true;
					}}
					onmouseleave={() => {
						errorPopoverOpen = false;
					}}
					class="max-w-64"
				>
					{error}
				</PopoverContent>
			</Popover>
		</InputGroupAddon>
	{/if}
</InputGroup>
