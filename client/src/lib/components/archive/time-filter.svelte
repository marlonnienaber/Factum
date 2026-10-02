<script lang="ts">
	import { RangeCalendar, type DateRange } from 'bits-ui';
	import { ChevronLeft, ChevronRight } from '@lucide/svelte';
	import { getLocalTimeZone, today, type DateValue } from '@internationalized/date';
	import DateInput from './date-input.svelte';

	type Props = {
		onSelectionChange: (startDate: DateValue | undefined, endDate: DateValue | undefined) => void;
	};

	const { onSelectionChange }: Props = $props();
	let startDateInputValue = $state<DateValue | 'partial' | undefined>();
	let endDateInputValue = $state<DateValue | 'partial' | undefined>();
	let placeholder: DateValue = $state(today(getLocalTimeZone()));
	const startDate = $derived(getDateValue(startDateInputValue));
	const endDate = $derived(getDateValue(endDateInputValue));
	const value = $derived<DateRange>(
		startDate === undefined && endDate !== undefined ? { start: endDate, end: endDate } : { start: startDate, end: endDate }
	);

	function handleCalendarValueChange(nextValue: DateRange): void {
		startDateInputValue = nextValue.start;
		endDateInputValue = nextValue.end;
		onSelectionChange(nextValue.start, nextValue.end);
	}

	function handleStartDateInputValueChange(nextValue: DateValue | 'partial' | undefined): void {
		startDateInputValue = nextValue;
		if (startDateInputValue !== 'partial' && endDateInputValue !== 'partial') {
			onSelectionChange(startDateInputValue, endDateInputValue);
		}
	}

	function handleEndDateInputValueChange(nextValue: DateValue | 'partial' | undefined): void {
		endDateInputValue = nextValue;
		if (startDateInputValue !== 'partial' && endDateInputValue !== 'partial') {
			onSelectionChange(startDateInputValue, endDateInputValue);
		}
	}

	function clearStart(): void {
		startDateInputValue = undefined;
		if (endDateInputValue !== 'partial') {
			onSelectionChange(startDateInputValue, endDateInputValue);
		}
	}

	function clearEnd(): void {
		endDateInputValue = undefined;
		if (startDateInputValue !== 'partial') {
			onSelectionChange(startDateInputValue, endDateInputValue);
		}
	}

	function getDateValue(dateInputValue: DateValue | 'partial' | undefined): DateValue | undefined {
		return dateInputValue === 'partial' ? undefined : dateInputValue;
	}

	function pairWeeks(weeks: DateValue[][]): DateValue[][] {
		const pairedWeeks: DateValue[][] = [];
		for (let index = 0; index < weeks.length; index += 2) {
			pairedWeeks.push([...weeks[index], ...(weeks[index + 1] ?? [])]);
		}
		return pairedWeeks;
	}
</script>

<div class="frame">
	<RangeCalendar.Root
		{value}
		onValueChange={handleCalendarValueChange}
		bind:placeholder
		locale="en-GB"
		weekStartsOn={1}
		weekdayFormat="short"
		monthFormat="long"
		yearFormat="numeric"
		fixedWeeks
	>
		{#snippet children({ months, weekdays })}
			<RangeCalendar.Header>
				{#snippet child({ props: headerProps })}
					<header {...headerProps} class="navigation-row">
						<RangeCalendar.PrevButton aria-label="Previous month">
							{#snippet child({ props })}
								<button {...props} class="navigation-button"><ChevronLeft size={18} /></button>
							{/snippet}
						</RangeCalendar.PrevButton>
						<RangeCalendar.Heading>
							{#snippet child({ props, headingValue })}
								<div {...props} class="month-title">{headingValue}</div>
							{/snippet}
						</RangeCalendar.Heading>
						<RangeCalendar.NextButton aria-label="Next month">
							{#snippet child({ props })}
								<button {...props} class="navigation-button"><ChevronRight size={18} /></button>
							{/snippet}
						</RangeCalendar.NextButton>
					</header>
				{/snippet}
			</RangeCalendar.Header>

			{#each months as month (month.value.toString())}
				<RangeCalendar.Grid>
					{#snippet child({ props: gridProps })}
						<table {...gridProps} class="calendar">
							<RangeCalendar.GridHead>
								<RangeCalendar.GridRow>
									{#each [...weekdays, ...weekdays] as weekday, index (`${weekday}-${index}`)}
										<RangeCalendar.HeadCell>
											{#snippet child({ props })}<th {...props} class="weekday">{weekday.slice(0, 2)}</th>{/snippet}
										</RangeCalendar.HeadCell>
									{/each}
								</RangeCalendar.GridRow>
							</RangeCalendar.GridHead>
							<RangeCalendar.GridBody>
								{#each pairWeeks(month.weeks) as dates (dates[0].toString())}
									<RangeCalendar.GridRow>
										{#each dates as date (date.toString())}
											<RangeCalendar.Cell {date} month={month.value}>
												<RangeCalendar.Day>
													{#snippet child({ props, day })}<div {...props} class="day">{day}</div>{/snippet}
												</RangeCalendar.Day>
											</RangeCalendar.Cell>
										{/each}
									</RangeCalendar.GridRow>
								{/each}
							</RangeCalendar.GridBody>
						</table>
					{/snippet}
				</RangeCalendar.Grid>
			{/each}
		{/snippet}
	</RangeCalendar.Root>

	<div class="date-input-row">
		<button type="button" class="clear-button" disabled={startDateInputValue === undefined} onclick={clearStart}>clear</button>
		<DateInput value={startDateInputValue} onValueChange={handleStartDateInputValueChange} label="Start" />
		<span class="date-range-separator">—</span>
		<DateInput value={endDateInputValue} onValueChange={handleEndDateInputValueChange} label="End" />
		<button type="button" class="clear-button" disabled={endDateInputValue === undefined} onclick={clearEnd}>clear</button>
	</div>
</div>

<style lang="scss">
	.frame {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1.25rem;
		margin-bottom: 1rem;

		.navigation-row {
			display: grid;
			grid-template-columns: auto 1fr auto;
			align-items: center;
			margin-bottom: 0.75rem;

			.navigation-button {
				display: flex;
				padding: 0;
				border: 0;
				background: transparent;
				color: var(--text);
				cursor: pointer;
			}

			.month-title {
				font-size: 0.8125rem;
				font-weight: 500;
				text-align: center;
			}
		}

		.calendar {
			border-spacing: 0 0.125rem;

			:global(td[data-selected]) {
				background: linear-gradient(var(--surface), var(--surface)) center / 100% 1.25rem no-repeat;
			}

			.weekday {
				color: var(--text-subtle);
				font-size: 0.875rem;
				font-weight: 400;
			}

			.day {
				display: flex;
				align-items: center;
				justify-content: center;
				width: 1.5rem;
				height: 1.5rem;
				border-radius: 0.375rem;
				font-size: 0.875rem;
				cursor: pointer;

				&[data-outside-month] {
					color: var(--text-subtle);
				}

				&[data-range-start],
				&[data-range-end] {
					background: var(--button);
					color: var(--text-inverse);
				}
			}
		}

		.date-input-row {
			display: grid;
			grid-auto-flow: column;
			align-items: center;
			gap: 0.5rem;

			.clear-button {
				padding: 0;
				border: 0;
				background: transparent;
				color: var(--text);
				font-size: 0.75rem;
				cursor: pointer;

				&:disabled {
					color: var(--text-subtle);
					cursor: default;
				}
			}

			.date-range-separator {
				margin: 0 0.75rem;
				font-weight: 600;
			}
		}
	}
</style>
