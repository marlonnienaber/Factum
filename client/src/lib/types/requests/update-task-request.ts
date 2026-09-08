import type { CalendarDate } from '@internationalized/date';

export type UpdateTaskRequest = {
	id: number;
	title: string;
	description?: string;
	date?: CalendarDate;
	categoryId: number;
};
