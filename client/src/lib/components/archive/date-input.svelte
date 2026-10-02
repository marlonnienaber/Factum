<script lang="ts">
	import { CalendarDate, parseDate, type DateValue } from '@internationalized/date';

	type SegmentValues = {
		day: string;
		month: string;
		year: string;
	};

	type Props = {
		value: DateValue | 'partial' | undefined;
		label: string;
		onValueChange: (value: DateValue | 'partial' | undefined) => void;
	};

	const emptySegmentValues = { day: '', month: '', year: '' };
	let { value, label, onValueChange }: Props = $props();
	let segmentValues = $state<SegmentValues>(emptySegmentValues);

	$effect(() => {
		if (value === undefined) {
			segmentValues = emptySegmentValues;
		} else if (value !== 'partial' && getDateValue(segmentValues)?.compare(value) !== 0) {
			segmentValues = getSegmentValues(value);
		}
	});

	function getSegmentValues(date: DateValue): SegmentValues {
		return {
			day: date.day.toString().padStart(2, '0'),
			month: date.month.toString().padStart(2, '0'),
			year: date.year.toString().padStart(4, '0')
		};
	}

	function getDateValue(dateSegments: SegmentValues): DateValue | undefined {
		const { day, month, year } = dateSegments;
		if (day.length !== 2 || month.length !== 2 || year.length !== 4) return undefined;
		try {
			return parseDate(`${year}-${month}-${day}`);
		} catch {
			return undefined;
		}
	}

	function updateSegment(part: 'day' | 'month' | 'year', input: HTMLInputElement): void {
		const sanitizedSegmentValue = input.value.replace(/\D/g, '').slice(0, part === 'year' ? 4 : 2);
		const nextSegmentValues = { ...segmentValues, [part]: sanitizedSegmentValue };
		if (isValidInput(nextSegmentValues)) {
			segmentValues = nextSegmentValues;
			onValueChange(getInputValue(nextSegmentValues));
			input.value = sanitizedSegmentValue;
			if (part !== 'year' && sanitizedSegmentValue.length === 2 && input.nextElementSibling instanceof HTMLInputElement) {
				input.nextElementSibling.focus();
			}
		} else {
			input.value = segmentValues[part];
		}
	}

	function isValidInput(sanitizedSegmentValues: SegmentValues): boolean {
		const { day, month, year } = sanitizedSegmentValues;
		const numericDay = Number(day);
		const numericMonth = Number(month);
		const numericYear = Number(year);
		if (day.length === 2 && (numericDay < 1 || numericDay > 31)) return false;
		if (month.length === 2 && (numericMonth < 1 || numericMonth > 12)) return false;
		if (year.length === 4 && numericYear < 1) return false;
		if (day.length === 2 && month.length === 2 && numericDay > getMaximumDaysInMonth(numericMonth)) return false;

		const isIncomplete = day.length !== 2 || month.length !== 2 || year.length !== 4;
		return isIncomplete || getDateValue({ day, month, year }) !== undefined;
	}

	function getMaximumDaysInMonth(month: number): number {
		const dateInLeapYear = new CalendarDate(2000, month, 1);
		return dateInLeapYear.calendar.getDaysInMonth(dateInLeapYear);
	}

	function getInputValue(inputSegments: SegmentValues): DateValue | 'partial' | undefined {
		if (inputSegments.day === '' && inputSegments.month === '' && inputSegments.year === '') return undefined;
		return getDateValue(inputSegments) ?? 'partial';
	}
</script>

<div class="frame">
	<input
		type="text"
		class="date-segment non-year"
		aria-label={`${label} day`}
		inputmode="numeric"
		maxlength="2"
		placeholder="dd"
		value={segmentValues.day}
		oninput={(event) => updateSegment('day', event.currentTarget)}
	/>
	<input
		type="text"
		class="date-segment non-year"
		aria-label={`${label} month`}
		inputmode="numeric"
		maxlength="2"
		placeholder="mm"
		value={segmentValues.month}
		oninput={(event) => updateSegment('month', event.currentTarget)}
	/>
	<input
		type="text"
		class="date-segment year"
		aria-label={`${label} year`}
		inputmode="numeric"
		maxlength="4"
		placeholder="yyyy"
		value={segmentValues.year}
		oninput={(event) => updateSegment('year', event.currentTarget)}
	/>
</div>

<style lang="scss">
	.frame {
		display: flex;
		gap: 0.5rem;

		.date-segment {
			box-sizing: border-box;
			padding: 0.25rem 0.375rem;
			border: 0;
			border-radius: 0.375rem;
			outline: none;
			background: var(--surface);
			caret-color: transparent;
			color: var(--text);
			font-family: inherit;
			font-size: 0.8125rem;
			text-align: center;

			&:focus {
				outline: 1px solid var(--border-accented);
			}

			&.non-year {
				width: 2.5rem;
			}

			&.year {
				width: 3rem;
			}

			&::placeholder {
				color: var(--text-subtle);
			}
		}
	}
</style>
