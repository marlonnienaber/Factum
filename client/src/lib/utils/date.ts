import { getLocalTimeZone, today, type CalendarDate } from '@internationalized/date';

export function getDateLabel(date: CalendarDate): string {
	if (date.compare(today(getLocalTimeZone())) === 0) return 'Today';
	return `${String(date.day).padStart(2, '0')}.${String(date.month).padStart(2, '0')}.${date.year}`;
}
