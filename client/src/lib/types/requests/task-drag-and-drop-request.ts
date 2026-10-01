import type { CalendarDate } from '@internationalized/date';

export type TaskDragAndDropRequest = {
	date?: CalendarDate;
	done: boolean;
	trashed: boolean;
	previousTaskId?: number;
};
