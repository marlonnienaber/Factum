export type UpdateTaskRequest = {
	id: number;
	title: string;
	description?: string;
	date?: string;
	categoryId: number;
};
