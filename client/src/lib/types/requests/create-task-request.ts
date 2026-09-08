import type { CalendarDate } from '@internationalized/date';

export type CreateTaskRequest = {
	title: string;
	description?: string;
	date?: CalendarDate;
	categoryId: number;
};
