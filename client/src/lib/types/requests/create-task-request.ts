export type CreateTaskRequest = {
	title: string;
	description?: string;
	date?: string;
	categoryId: number;
};
