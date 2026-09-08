import type { CalendarDate } from '@internationalized/date';

export type ChangeTaskDateAndOrRankRequest = {
	id: number;
	date?: CalendarDate;
	rank: string;
};
