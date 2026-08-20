<script lang="ts">
	import { CalendarDays, ChevronLeft, ChevronRight, X } from '@lucide/svelte';
	import type { CalendarDate } from '@internationalized/date';
	import { DatePicker } from 'bits-ui';

	type Props = {
		id?: string;
		value: CalendarDate | undefined;
	};

	let { id, value = $bindable() }: Props = $props();
	let open = $state(false);

	function isPlaceholder(segmentValue: string): boolean {
		return segmentValue === 'dd' || segmentValue === 'mm' || segmentValue === 'yyyy';
	}
</script>

<DatePicker.Root bind:value bind:open locale="en-GB" weekdayFormat="short" fixedWeeks>
	<DatePicker.Input {id} name="date">
		{#snippet child({ props: inputProps, segments })}
			<div {...inputProps} class:active={open} class="date-input">
				<div class="date-segments">
					{#each segments as { part, value: segmentValue }, index (part + index)}
						<DatePicker.Segment {part}>
							{#snippet child({ props })}
								<span {...props} class:placeholder={isPlaceholder(segmentValue)} class="date-segment">
									{part === 'literal' ? '.' : segmentValue}
								</span>
							{/snippet}
						</DatePicker.Segment>
					{/each}
				</div>
				<div class="date-actions">
					{#if value}
						<button type="button" class="date-action" aria-label="Clear date" onclick={() => (value = undefined)}>
							<X size={18} />
						</button>
					{/if}
					<DatePicker.Trigger>
						{#snippet child({ props: triggerProps })}
							<button {...triggerProps} class="date-action" aria-label="Choose date">
								<CalendarDays size={18} />
							</button>
						{/snippet}
					</DatePicker.Trigger>
				</div>
			</div>
		{/snippet}
	</DatePicker.Input>

	<DatePicker.Portal>
		<DatePicker.Content sideOffset={6}>
			{#snippet child({ props: contentProps, wrapperProps })}
				<div {...wrapperProps}>
					<div {...contentProps} class="calendar-content">
						<DatePicker.Calendar>
							{#snippet children({ months, weekdays })}
								<DatePicker.Header>
									{#snippet child({ props: headerProps })}
										<header {...headerProps} class="calendar-header">
											<DatePicker.PrevButton aria-label="Previous month">
												{#snippet child({ props })}
													<button {...props} class="calendar-navigation"><ChevronLeft size={18} /></button>
												{/snippet}
											</DatePicker.PrevButton>
											<DatePicker.Heading />
											<DatePicker.NextButton aria-label="Next month">
												{#snippet child({ props })}
													<button {...props} class="calendar-navigation"><ChevronRight size={18} /></button>
												{/snippet}
											</DatePicker.NextButton>
										</header>
									{/snippet}
								</DatePicker.Header>
								{#each months as month (month.value.toString())}
									<DatePicker.Grid>
										{#snippet child({ props: gridProps })}
											<table {...gridProps} class="calendar-grid">
												<DatePicker.GridHead>
													<DatePicker.GridRow>
														{#each weekdays as day (day)}
															<DatePicker.HeadCell>
																{#snippet child({ props })}<th {...props} class="calendar-head-cell">{day}</th>{/snippet}
															</DatePicker.HeadCell>
														{/each}
													</DatePicker.GridRow>
												</DatePicker.GridHead>
												<DatePicker.GridBody>
													{#each month.weeks as weekDates (weekDates[0].toString())}
														<DatePicker.GridRow>
															{#each weekDates as date (date.toString())}
																<DatePicker.Cell {date} month={month.value}>
																	<DatePicker.Day>
																		{#snippet child({ props, day })}<div {...props} class="calendar-day">{day}</div>{/snippet}
																	</DatePicker.Day>
																</DatePicker.Cell>
															{/each}
														</DatePicker.GridRow>
													{/each}
												</DatePicker.GridBody>
											</table>
										{/snippet}
									</DatePicker.Grid>
								{/each}
							{/snippet}
						</DatePicker.Calendar>
					</div>
				</div>
			{/snippet}
		</DatePicker.Content>
	</DatePicker.Portal>
</DatePicker.Root>

<style lang="scss">
	.date-input {
		display: flex;
		align-items: center;
		justify-content: space-between;
		width: 100%;
		height: 2.75rem;
		box-sizing: border-box;
		padding: 0.625rem;
		border: 1px solid var(--border);
		border-radius: 0.5rem;
		background: transparent;
		color: var(--text);

		&.active,
		&:has(.date-segment:focus) {
			border-color: var(--border-accented);
		}
	}

	.date-segments {
		display: flex;
	}

	.date-segment {
		padding: 0.125rem 0.375rem;
		border-radius: 0.25rem;
		outline: none;

		&:focus {
			background: var(--surface-subtle);
		}
	}

	.placeholder {
		color: var(--text-subtle);
	}

	.date-actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.date-action {
		display: flex;
		padding: 0;
		border: 0;
		background: transparent;
		color: var(--text);
		cursor: pointer;
	}

	.calendar-content {
		z-index: 60;
		padding: 0.75rem;
		border: 1px solid var(--border);
		border-radius: 0.75rem;
		background: var(--background);
		color: var(--text);
		box-shadow: 0 0.75rem 2rem rgb(0 0 0 / 30%);

		.calendar-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			margin-bottom: 0.5rem;
		}

		.calendar-navigation {
			display: flex;
			padding: 0.25rem;
			border: 0;
			background: transparent;
			color: var(--text);
			cursor: pointer;
		}

		.calendar-grid {
			border-spacing: 0.2rem;
		}

		.calendar-head-cell {
			color: var(--text-subtle);
			font-size: 0.75rem;
			font-weight: 400;
		}

		.calendar-day {
			display: flex;
			align-items: center;
			justify-content: center;
			width: 2rem;
			height: 2rem;
			border-radius: 0.5rem;
			cursor: pointer;
		}

		.calendar-day:hover,
		.calendar-day[data-selected] {
			background: var(--button);
			color: var(--text-inverse);
		}

		.calendar-day[data-outside-month] {
			color: var(--text-subtle);
		}
	}
</style>
