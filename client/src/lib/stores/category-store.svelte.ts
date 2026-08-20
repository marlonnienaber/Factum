import {
	type NetworkCall,
	jsonDeleteWithoutBodyWithoutResponse,
	jsonGetWithoutBodyWithResponse,
	jsonPostWithBodyWithResponse
} from '$lib/api/api';
import type { CreateNewCategoryRequest } from '$lib/types/requests/create-new-category-request';
import { CategoryResponseSchema, type CategoryResponse } from '$lib/types/responses/category-response';
import { z } from 'zod';

let categories: CategoryResponse[] = $state<CategoryResponse[]>([]);
let loading = $state(false);

async function fetchCategories(): Promise<void> {
	if (loading) return;
	loading = true;
	const networkCall: NetworkCall<CategoryResponse[]> = await jsonGetWithoutBodyWithResponse(
		'/api/categories',
		z.array(CategoryResponseSchema)
	);
	if (networkCall.status === 'success') categories = networkCall.result;
	loading = false;
}

async function createCategory(request: CreateNewCategoryRequest): Promise<void> {
	if (loading) return;
	loading = true;
	const networkCall: NetworkCall<CategoryResponse> = await jsonPostWithBodyWithResponse(
		'/api/categories',
		request,
		CategoryResponseSchema
	);
	if (networkCall.status === 'success') categories = [...categories, networkCall.result];
	loading = false;
}

async function deleteCategory(id: number): Promise<void> {
	if (loading) return;
	loading = true;
	const networkCall: NetworkCall<void> = await jsonDeleteWithoutBodyWithoutResponse(`/api/categories/${id}`);
	if (networkCall.status === 'success') categories = categories.filter((category) => category.id !== id);
	loading = false;
}

function reset(): void {
	categories = [];
}

export const categoryStore = {
	get categories() {
		return categories;
	},
	get loading() {
		return loading;
	},
	fetchCategories,
	createCategory,
	deleteCategory,
	reset
};
