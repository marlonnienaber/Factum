import type { ZonedDateTime } from '@internationalized/date';

export type TrashedTasksCursor = {
	trashedDate: ZonedDateTime;
	id: number;
};
