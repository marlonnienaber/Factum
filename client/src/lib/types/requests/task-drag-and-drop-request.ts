import type { CalendarDate } from '@internationalized/date';

export type TaskDragAndDropRequest = {
	date?: CalendarDate;
	previousTaskId?: number;
};
