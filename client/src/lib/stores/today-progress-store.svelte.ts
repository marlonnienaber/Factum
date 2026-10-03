import type { GetTodayProgressResponse } from '$lib/types/responses/get-today-progress-response';
import { getLocalTimeZone, today, type CalendarDate } from '@internationalized/date';

export type ProgressRelevantTaskState = {
	date?: CalendarDate;
	done: boolean;
	trashed: boolean;
};

let openTaskCount = $state(0);
let doneTaskCount = $state(0);

function setProgress(progress: GetTodayProgressResponse): void {
	openTaskCount = progress.openTaskCount;
	doneTaskCount = progress.doneTaskCount;
}

/**
 * Updates today's progress by first removing the task's previous progress contribution and then adding its next contribution.
 * A task contributes to exactly one count when it is untrashed and dated today: to the done count if completed, or to the
 * open count otherwise. An undefined state represents a task that does not exist before or after the transition.
 */
function applyTaskTransition(previousState?: ProgressRelevantTaskState, nextState?: ProgressRelevantTaskState): void {
	const todayDate = today(getLocalTimeZone());
	if (previousState && !previousState.trashed && previousState.date?.compare(todayDate) === 0) {
		if (previousState.done) {
			doneTaskCount--;
		} else {
			openTaskCount--;
		}
	}
	if (nextState && !nextState.trashed && nextState.date?.compare(todayDate) === 0) {
		if (nextState.done) {
			doneTaskCount++;
		} else {
			openTaskCount++;
		}
	}
}

function reset(): void {
	openTaskCount = 0;
	doneTaskCount = 0;
}

export const todayProgressStore = {
	get openTaskCount(): number {
		return openTaskCount;
	},
	get doneTaskCount(): number {
		return doneTaskCount;
	},
	setProgress,
	applyTaskTransition,
	reset
};
